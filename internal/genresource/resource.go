package genresource

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/bwanclient"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/tfschema"
)

// Definition describes one generated resource: the schema to expose and the API
// calls behind each Terraform operation.
type Definition struct {
	// Name is the resource name without the provider prefix, e.g. "segment".
	Name string
	// Schema is the generated schema function.
	Schema func(context.Context) rschema.Schema
	Create Operation
	Read   Operation
	// Update is unset where the API has no update operation; any change then
	// replaces the object.
	Update Operation
	Delete Operation
	// RawJSONAttributes holds dot-separated paths of string attributes carrying
	// an embedded JSON document.
	RawJSONAttributes []string
	// VariantBlocks holds dot-separated paths of the blocks standing for the forms
	// an object or one of its nested objects can take, of which exactly one is set.
	VariantBlocks []string
	// FieldNames maps an API field name to the attribute name it is exposed under,
	// for the handful of names Terraform reserves.
	FieldNames map[string]string
	// RawFeature names the opt-in guarding this resource, set when its
	// configuration is an opaque JSON document the provider does not model yet.
	RawFeature string
	// Variant pins this resource to one kind of object where its endpoints serve
	// several, e.g. the wanlink kind of tag on /tags.
	Variant *Variant
}

// IdentityAttributes lists what it takes to address one of these objects: the
// path placeholders of whatever it lives under, ending with its own id. Joined by
// "/" that is also the resource's import ID, and the example generator documents
// it from here so that what the docs tell a practitioner to type is what
// ImportState parses.
func (d Definition) IdentityAttributes() []string {
	names := parentPlaceholders(d.Create.Path, d.Read.Path, d.Update.Path, d.Delete.Path)

	return append(names, idAttribute)
}

// NewResource returns the factory the provider registers for def.
func NewResource(def Definition) func() resource.Resource {
	return func() resource.Resource {
		return &genericResource{def: def}
	}
}

var (
	_ resource.Resource                   = (*genericResource)(nil)
	_ resource.ResourceWithConfigure      = (*genericResource)(nil)
	_ resource.ResourceWithImportState    = (*genericResource)(nil)
	_ resource.ResourceWithModifyPlan     = (*genericResource)(nil)
	_ resource.ResourceWithValidateConfig = (*genericResource)(nil)
)

// ValidateConfig checks that exactly one of each set of alternative blocks is set.
// The API accepts one form of an object and Terraform has no type for that, so the
// rule is enforced here instead — at validate time, before a plan is even drawn.
func (r *genericResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	r.prepare(ctx)

	resp.Diagnostics.Append(validateVariants(r.model, req.Config.Raw)...)
}

type genericResource struct {
	def  Definition
	meta *Meta

	schema rschema.Schema
	model  *tfschema.Model
}

// prepare builds the schema and its model on first use. The framework creates a
// fresh resource for every call and does not promise Schema runs before the
// operation, so neither can be built in a constructor.
func (r *genericResource) prepare(ctx context.Context) {
	if r.model != nil {
		return
	}

	r.schema = decorate(ctx, r.def)
	r.model = tfschema.FromResource(ctx, r.schema, r.def.RawJSONAttributes, r.def.VariantBlocks, r.def.FieldNames)
}

func (r *genericResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.def.Name
}

func (r *genericResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	r.prepare(ctx)

	resp.Schema = r.schema
}

func (r *genericResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	meta, ok := req.ProviderData.(*Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data",
			fmt.Sprintf("Expected a configured BWAN provider, got %T. This is a bug in the provider.", req.ProviderData),
		)

		return
	}

	r.meta = meta
}

func (r *genericResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.prepare(ctx)

	if !r.ready(&resp.Diagnostics) {
		return
	}

	plan := req.Plan.Raw

	document, found, diags := r.write(ctx, r.def.Create, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.Diagnostics.AddError(
			"Unexpected API response",
			fmt.Sprintf("%s %s returned no document, so the new %s cannot be tracked.", r.def.Create.Method, r.def.Create.Path, r.def.Name),
		)

		return
	}

	state, diags := r.apply(tfschema.AfterWrite, document, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := requireID(state); err != nil {
		resp.Diagnostics.AddError(
			"Unexpected API response",
			fmt.Sprintf("The created %s cannot be tracked: %s.", r.def.Name, err),
		)

		return
	}

	resp.State.Raw = state
}

func (r *genericResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.prepare(ctx)

	if !r.ready(&resp.Diagnostics) {
		return
	}

	state := req.State.Raw

	document, found, diags := r.read(ctx, state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)

		return
	}

	next, diags := r.apply(tfschema.AfterRead, document, state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Raw = next
}

func (r *genericResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.prepare(ctx)

	if !r.ready(&resp.Diagnostics) {
		return
	}

	if !r.def.Update.Supported() {
		resp.Diagnostics.AddError(
			"Update not supported",
			fmt.Sprintf("The API cannot update a %s; it has to be replaced. This is a bug in the provider.", r.def.Name),
		)

		return
	}

	plan := req.Plan.Raw

	document, found, diags := r.write(ctx, r.def.Update, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		// An update answering with no body leaves the computed attributes
		// unresolved, so the object is read back.
		document, found, diags = r.read(ctx, plan)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		if !found {
			resp.Diagnostics.AddError(
				"Unexpected API response",
				fmt.Sprintf("The updated %s could no longer be found.", r.def.Name),
			)

			return
		}
	}

	state, diags := r.apply(tfschema.AfterWrite, document, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Raw = state
}

func (r *genericResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.prepare(ctx)

	if !r.ready(&resp.Diagnostics) {
		return
	}

	requestPath, err := resolvePath(r.def.Delete.Path, req.State.Raw)
	if err != nil {
		resp.Diagnostics.AddError("Incomplete state", fmt.Sprintf("Cannot build the request path for %s: %s.", r.def.Name, err))

		return
	}

	client, diags := r.meta.clientFor(req.State.Raw)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err = client.Do(ctx, bwanclient.Request{Method: r.def.Delete.Method, Path: requestPath})
	if err != nil && !errors.Is(err, bwanclient.ErrNotFound) {
		resp.Diagnostics.AddError("Could not delete "+r.def.Name, err.Error())
	}
}

// ImportState accepts the path placeholders of the resource, ending with its id,
// joined by "/" — just "id" for most resources, "group_id/id" for one nested
// under a parent.
func (r *genericResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.prepare(ctx)

	names := r.def.IdentityAttributes()
	parts := strings.Split(req.ID, "/")

	if len(parts) != len(names) {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected %q, got %q.", strings.Join(names, "/"), req.ID),
		)

		return
	}

	for i, name := range names {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(name), parts[i])...)
	}
}

// ModifyPlan reports the changed attributes of a resource the API cannot update,
// so Terraform replaces it instead of calling an update that does not exist.
func (r *genericResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.prepare(ctx)

	// Planning is the first point at which the provider's configuration is known,
	// so it is where a missing opt-in is reported: failing here means a
	// practitioner never gets as far as applying one of these resources by
	// accident.
	if r.meta != nil && !gate(&resp.Diagnostics, "Resource", r.typeName(), r.def.RawFeature, r.meta.RawEnabled) {
		return
	}

	if r.def.Update.Supported() || req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var prior, planned map[string]tftypes.Value

	if err := req.State.Raw.As(&prior); err != nil {
		resp.Diagnostics.AddError("Unexpected state", err.Error())

		return
	}

	if err := req.Plan.Raw.As(&planned); err != nil {
		resp.Diagnostics.AddError("Unexpected plan", err.Error())

		return
	}

	for _, name := range r.model.ConfigurableNames() {
		if !planned[name].Equal(prior[name]) {
			resp.RequiresReplace = resp.RequiresReplace.Append(path.Root(name))
		}
	}
}

// write sends a create or update request and returns the document it answered
// with, if any.
func (r *genericResource) write(ctx context.Context, op Operation, value tftypes.Value) (any, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	requestPath, err := resolvePath(op.Path, value)
	if err != nil {
		diags.AddError("Incomplete configuration", fmt.Sprintf("Cannot build the request path for %s: %s.", r.def.Name, err))

		return nil, false, diags
	}

	body, err := r.model.Body(value, skipSet(append(placeholders(op.Path), OperatingTenantAttribute)))
	if err != nil {
		diags.AddError("Invalid configuration", fmt.Sprintf("Cannot build the request body for %s: %s.", r.def.Name, err))

		return nil, false, diags
	}

	client, clientDiags := r.meta.clientFor(value)
	diags.Append(clientDiags...)

	if diags.HasError() {
		return nil, false, diags
	}

	document, found, err := fetch(ctx, client, bwanclient.Request{Method: op.Method, Path: requestPath, Body: body})
	if err != nil {
		diags.AddError("Could not write "+r.def.Name, err.Error())

		return nil, false, diags
	}

	return document, found, diags
}

// read fetches the object behind value. Not every object has a single-object GET;
// where it does not, it is looked up in the collection it belongs to.
func (r *genericResource) read(ctx context.Context, value tftypes.Value) (any, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	requestPath, err := resolvePath(r.def.Read.Path, value)
	if err != nil {
		diags.AddError("Incomplete state", fmt.Sprintf("Cannot build the request path for %s: %s.", r.def.Name, err))

		return nil, false, diags
	}

	client, clientDiags := r.meta.clientFor(value)
	diags.Append(clientDiags...)

	if diags.HasError() {
		return nil, false, diags
	}

	if slices.Contains(placeholders(r.def.Read.Path), idAttribute) {
		document, found, err := fetch(ctx, client, bwanclient.Request{Method: r.def.Read.Method, Path: requestPath})

		if errors.Is(err, bwanclient.ErrNotFound) {
			return nil, false, diags
		}

		if err != nil {
			diags.AddError("Could not read "+r.def.Name, err.Error())

			return nil, false, diags
		}

		if found && !r.def.Variant.Matches(document) {
			diags.AddError("Wrong kind of object", r.def.Variant.Mismatch(r.typeName(), document))

			return nil, false, diags
		}

		return document, found, diags
	}

	identity, err := stringMembers(value, []string{idAttribute})
	if err != nil {
		diags.AddError("Incomplete state", fmt.Sprintf("Cannot identify the %s to read: %s.", r.def.Name, err))

		return nil, false, diags
	}

	element, found, err := findByID(ctx, client, requestPath, identity[idAttribute], r.def.Variant)
	if err != nil {
		diags.AddError("Could not read "+r.def.Name, err.Error())

		return nil, false, diags
	}

	return element, found, diags
}

func (r *genericResource) apply(mode tfschema.Mode, document any, current tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	value, err := r.model.Apply(mode, document, current)
	if err != nil {
		diags.AddError(
			"Unexpected API response",
			fmt.Sprintf("Could not map the %s response into state: %s.", r.def.Name, err),
		)
	}

	return value, diags
}

// ready reports whether the resource may talk to the API: the provider has to be
// configured, and one whose configuration is an opaque document has to have been
// opted in to.
func (r *genericResource) ready(diags *diag.Diagnostics) bool {
	if r.meta == nil || r.meta.Client == nil {
		diags.AddError("Provider not configured", "The BWAN provider has no API client. This is a bug in the provider.")

		return false
	}

	return gate(diags, "Resource", r.typeName(), r.def.RawFeature, r.meta.RawEnabled)
}

func (r *genericResource) typeName() string {
	return r.meta.TypeName + "_" + r.def.Name
}

func requireID(value tftypes.Value) error {
	members, err := stringMembers(value, []string{idAttribute})
	if err != nil {
		return err
	}

	if members[idAttribute] == "" {
		return errors.New("the API returned an empty id")
	}

	return nil
}
