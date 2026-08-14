package genresource

import (
	"context"
	"maps"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	fwpath "github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// decorate adjusts a generated resource schema for the three things an OpenAPI
// document cannot express about a Terraform resource: which attributes identify
// the parent an object lives under, that an id is stable, and which strings are
// really JSON documents.
func decorate(ctx context.Context, def Definition) rschema.Schema {
	schema := def.Schema(ctx)
	attributes := maps.Clone(schema.Attributes)

	if attributes == nil {
		attributes = map[string]rschema.Attribute{}
	}

	// A path placeholder other than "id" names the object this one lives under.
	// The generator only sees a path parameter and maps it as optional; it is
	// really a required reference, and pointing a resource at a different parent
	// means creating it somewhere else.
	for _, name := range parentPlaceholders(def.Create.Path, def.Read.Path, def.Update.Path, def.Delete.Path) {
		text := descriptionOf(attributes[name])
		if text == "" {
			text = "Identifier of the object this " + strings.ReplaceAll(def.Name, "_", " ") + " belongs to."
		}

		attributes[name] = rschema.StringAttribute{
			Required:            true,
			Description:         text,
			MarkdownDescription: text,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		}
	}

	attributes[idAttribute] = idAttributeSchema(attributes[idAttribute])

	attributes[OperatingTenantAttribute] = operatingTenantResourceAttribute()

	for _, path := range def.RawJSONAttributes {
		// Only a top-level attribute is given the JSON custom type: a nested one
		// would no longer match the attribute types its generated parent object
		// declares. Nested JSON documents still round-trip, they are just
		// compared as text.
		if !strings.Contains(path, ".") {
			markJSONDocument(attributes, path)
		}
	}

	schema.Attributes = attributes

	return schema
}

// idAttributeSchema keeps the API-assigned id out of every plan: it is set once
// at create and never changes, so re-reading it as "known after apply" would add
// noise to every diff.
func idAttributeSchema(existing rschema.Attribute) rschema.Attribute {
	if attribute, ok := existing.(rschema.StringAttribute); ok {
		attribute.PlanModifiers = append(attribute.PlanModifiers, stringplanmodifier.UseStateForUnknown())

		return attribute
	}

	const text = "Identifier assigned by the API."

	return rschema.StringAttribute{
		Computed:            true,
		Description:         text,
		MarkdownDescription: text,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func markJSONDocument(attributes map[string]rschema.Attribute, name string) {
	attribute, ok := attributes[name].(rschema.StringAttribute)
	if !ok {
		return
	}

	attribute.CustomType = jsontypes.NormalizedType{}
	attributes[name] = attribute
}

func descriptionOf(attribute rschema.Attribute) string {
	if attribute == nil {
		return ""
	}

	return attribute.GetDescription()
}

// decorateDataSource lets a data source that addresses one object find it by
// filter as well as by id. The OpenAPI document has no way to say that two
// different endpoints reach the same object, so the generated schema only knows
// about the id in the path.
func decorateDataSource(ctx context.Context, def DataSourceDefinition) dschema.Schema {
	schema := def.Schema(ctx)

	attributes := maps.Clone(schema.Attributes)
	if attributes == nil {
		attributes = map[string]dschema.Attribute{}
	}

	attributes[OperatingTenantAttribute] = operatingTenantDataSourceAttribute()

	if !def.Search.Supported() {
		schema.Attributes = attributes

		return schema
	}

	const idDescription = "Identifier of the object to read. Exactly one of `id` or `filter` is required."

	// The id stops being required, but stays computed so it is filled in when the
	// object was found by filter instead.
	attributes[idAttribute] = dschema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Description:         idDescription,
		MarkdownDescription: idDescription,
		Validators: []validator.String{
			stringvalidator.ExactlyOneOf(fwpath.MatchRoot(idAttribute), fwpath.MatchRoot(filterAttribute)),
		},
	}

	const filterDescription = "Filter selecting the object to read, in the API's filter syntax, " +
		"e.g. `name eq \"corporate\"`. It has to match exactly one object. " +
		"Exactly one of `id` or `filter` is required."

	attributes[filterAttribute] = dschema.StringAttribute{
		Optional:            true,
		Description:         filterDescription,
		MarkdownDescription: filterDescription,
	}

	schema.Attributes = attributes

	return schema
}
