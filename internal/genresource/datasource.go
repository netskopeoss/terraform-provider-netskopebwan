package genresource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/bwanclient"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/tfschema"
)

// DataSourceDefinition describes one generated data source: the schema to expose
// and the API call behind it.
type DataSourceDefinition struct {
	// Name is the data source name without the provider prefix, e.g. "segments".
	Name string
	// Schema is the generated schema function.
	Schema func(context.Context) dschema.Schema
	Read   Operation
	// Search is the filterable collection the object read by Read belongs to,
	// set when the API lets that collection be filtered. It is what makes
	// looking an object up by `filter` instead of by `id` possible.
	Search Operation
	// RawJSONAttributes holds dot-separated paths of string attributes carrying
	// an embedded JSON document.
	RawJSONAttributes []string
	// VariantBlocks holds dot-separated paths of the blocks standing for the forms
	// an object or one of its nested objects can take, of which exactly one is set.
	VariantBlocks []string
	// FieldNames maps an API field name to the attribute name it is exposed under,
	// for the handful of names Terraform reserves.
	FieldNames map[string]string
	// RawFeature names the opt-in guarding this data source, set when the object
	// it reads carries its configuration as an opaque JSON document.
	RawFeature string
	// Variant pins this data source to one kind of object where its endpoint serves
	// several.
	Variant *Variant
}

// NewDataSource returns the factory the provider registers for def.
func NewDataSource(def DataSourceDefinition) func() datasource.DataSource {
	return func() datasource.DataSource {
		return &genericDataSource{def: def}
	}
}

var (
	_ datasource.DataSource              = (*genericDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*genericDataSource)(nil)
)

type genericDataSource struct {
	def  DataSourceDefinition
	meta *Meta

	schema dschema.Schema
	model  *tfschema.Model
	// collection records that the endpoint answers with a paginated envelope
	// rather than a single object.
	collection bool
}

func (d *genericDataSource) prepare(ctx context.Context) {
	if d.model != nil {
		return
	}

	d.schema = decorateDataSource(ctx, d.def)
	d.model = tfschema.FromDataSource(ctx, d.schema, d.def.RawJSONAttributes, d.def.VariantBlocks, d.def.FieldNames)
	d.collection = d.model.Attrs[dataField] != nil && d.model.Attrs[totalCountField] != nil
}

// walksCollection reports whether reading the one object this data source stands
// for means finding it in a collection. Not every object has a single-object GET;
// where it does not, the read is the collection the object belongs to, which is
// how a resource without one is refreshed as well.
func (d *genericDataSource) walksCollection() bool {
	return !d.collection && readsWholeCollection(d.def.Read.Path)
}

// readsWholeCollection reports whether a read path names a collection rather than
// one object in it.
func readsWholeCollection(path string) bool {
	return !slices.Contains(placeholders(path), idAttribute)
}

func (d *genericDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.def.Name
}

func (d *genericDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	d.prepare(ctx)

	resp.Schema = d.schema
}

func (d *genericDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.meta = meta
}

func (d *genericDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	d.prepare(ctx)

	if d.meta == nil || d.meta.Client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The BWAN provider has no API client. This is a bug in the provider.")

		return
	}

	if !gate(&resp.Diagnostics, "Data source", d.meta.TypeName+"_"+d.def.Name, d.def.RawFeature, d.meta.RawEnabled) {
		return
	}

	config := req.Config.Raw

	document, diags := d.document(ctx, config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state, err := d.model.Apply(tfschema.AfterWrite, document, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected API response",
			fmt.Sprintf("Could not map the %s response into state: %s.", d.def.Name, err),
		)

		return
	}

	resp.State.Raw = state
}

// document fetches what the data source describes: an object addressed by id, an
// object matched by filter, or a whole collection.
func (d *genericDataSource) document(ctx context.Context, config tftypes.Value) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	skip := append(placeholders(d.def.Read.Path), OperatingTenantAttribute)

	// An id the read walks a collection for is what the walk looks for, not
	// something the API narrows the collection by.
	if d.walksCollection() {
		skip = append(skip, idAttribute)
	}

	// Every configurable attribute that is not part of the path is a filter the
	// API takes as a query parameter.
	query, err := d.model.Query(config, skipSet(skip))
	if err != nil {
		diags.AddError("Invalid configuration", fmt.Sprintf("Cannot build the query for %s: %s.", d.def.Name, err))

		return nil, diags
	}

	client, clientDiags := d.meta.clientFor(config)
	diags.Append(clientDiags...)

	if diags.HasError() {
		return nil, diags
	}

	if d.searchable(config) {
		return d.search(ctx, client, config, url.Values(query))
	}

	requestPath, err := resolvePath(d.def.Read.Path, config)
	if err != nil {
		diags.AddError("Incomplete configuration", fmt.Sprintf("Cannot build the request path for %s: %s.", d.def.Name, err))

		return nil, diags
	}

	if d.walksCollection() {
		return d.element(ctx, client, requestPath, config, url.Values(query))
	}

	return d.fetch(ctx, client, requestPath, url.Values(query))
}

// element finds the one object of a collection carrying the configured id, for an
// object the API has no single-object endpoint for.
func (d *genericDataSource) element(ctx context.Context, client bwanclient.API, requestPath string, config tftypes.Value, query url.Values) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	identity, err := stringMembers(config, []string{idAttribute})
	if err != nil {
		diags.AddError("Incomplete configuration", fmt.Sprintf("Cannot identify the %s to read: %s.", d.def.Name, err))

		return nil, diags
	}

	element, found, err := findByID(ctx, client, requestPath, identity[idAttribute], d.def.Variant, query)
	if err != nil {
		diags.AddError("Could not read "+d.def.Name, err.Error())

		return nil, diags
	}

	if !found {
		diags.AddError(
			fmt.Sprintf("No %s found", d.def.Name),
			fmt.Sprintf("Nothing in %s has the id %q.", d.def.Read.Path, identity[idAttribute]),
		)

		return nil, diags
	}

	return d.def.Variant.Flatten(element), diags
}

// searchable reports whether this read is a lookup by filter rather than by id.
// The schema already guarantees exactly one of the two is set.
func (d *genericDataSource) searchable(config tftypes.Value) bool {
	if !d.def.Search.Supported() {
		return false
	}

	filter, err := stringMembers(config, []string{filterAttribute})

	return err == nil && filter[filterAttribute] != ""
}

// search finds the one object matching a filter. Anything other than one match is
// an error: a data source stands for a single object, and quietly picking the
// first of several would make the configuration depend on the API's ordering.
func (d *genericDataSource) search(ctx context.Context, client bwanclient.API, config tftypes.Value, query url.Values) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	searchPath, err := resolvePath(d.def.Search.Path, config)
	if err != nil {
		diags.AddError("Incomplete configuration", fmt.Sprintf("Cannot build the search path for %s: %s.", d.def.Name, err))

		return nil, diags
	}

	all, err := fetchAll(ctx, client, searchPath, query)
	if err != nil {
		diags.AddError("Could not search "+d.def.Name, err.Error())

		return nil, diags
	}

	filter := query.Get(filterAttribute)
	all.Data = d.def.Variant.Keep(all.Data)

	switch len(all.Data) {
	case 1:
		return d.def.Variant.Flatten(all.Data[0]), diags
	case 0:
		diags.AddError(
			fmt.Sprintf("No %s found", d.def.Name),
			fmt.Sprintf("No %s matches the filter %q.", d.def.Name, filter),
		)
	default:
		diags.AddError(
			fmt.Sprintf("Multiple %s found", d.def.Name),
			fmt.Sprintf("The filter %q matches %d objects; it has to match exactly one.", filter, len(all.Data)),
		)
	}

	return nil, diags
}

func (d *genericDataSource) fetch(ctx context.Context, client bwanclient.API, requestPath string, query url.Values) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	// A list data source stands for the whole collection: every page is walked, so
	// `data` holds the list rather than whatever the server's default page size
	// happens to be. There is no argument for asking for one page, which is why
	// the cursor parameters are not in the schema.
	if d.collection {
		all, err := fetchAll(ctx, client, requestPath, query)
		if err != nil {
			diags.AddError("Could not read "+d.def.Name, err.Error())

			return nil, diags
		}

		all.Data = d.def.Variant.FlattenAll(d.def.Variant.Keep(all.Data))

		return all.Document(), diags
	}

	document, found, err := fetch(ctx, client, bwanclient.Request{Method: d.def.Read.Method, Path: requestPath, Query: query})
	if err != nil {
		if errors.Is(err, bwanclient.ErrNotFound) {
			diags.AddError(
				fmt.Sprintf("No %s found", d.def.Name),
				fmt.Sprintf("The API has no %s matching this configuration: %s.", d.def.Name, err),
			)

			return nil, diags
		}

		diags.AddError("Could not read "+d.def.Name, err.Error())

		return nil, diags
	}

	if !found {
		diags.AddError(
			"Unexpected API response",
			fmt.Sprintf("%s %s returned no document.", d.def.Read.Method, d.def.Read.Path),
		)

		return nil, diags
	}

	if !d.def.Variant.Matches(document) {
		diags.AddError(
			fmt.Sprintf("No %s found", d.def.Name),
			d.def.Variant.Mismatch(d.meta.TypeName+"_"+d.def.Name, document),
		)

		return nil, diags
	}

	// Matches reads the kind off the document as the API shaped it, so reshaping is
	// the last thing that happens to it.
	return d.def.Variant.Flatten(document), diags
}
