package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/genresource"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/provider"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/registry"
)

// Examples are generated because there is a documentation page per object, and a
// page describing every attribute without showing one line of configuration
// answers nobody's first question. Writing 84 of them by hand would put a file per
// object back into a repository whose point is that there is no file per object,
// and they would rot the next time the spec moved.
//
// They are generated from the schema the provider itself reports — the decorated
// one, not the generated schema underneath it, because that is where a path
// placeholder has become a required argument and a collection has gained its
// filter. An example is therefore made of arguments that exist, of the types
// Terraform will check.
//
// It is deliberately minimal: what the object cannot be created without, and
// nothing else. Guessing at optional arguments produces an example a practitioner
// has to check against the schema below it, which is the situation this is meant
// to fix.
//
// A file already on disk is never overwritten, so anything worth explaining
// properly can be, and the generator fills in the long tail.
const (
	// exampleLabel names the resource or data source in every generated example.
	exampleLabel = "example"

	// idExample stands in for an object id. The API's ids are Mongo object ids, so
	// one that looks like an id reads better than a word does.
	idExample = "6501f0c2e4b0a1b2c3d4e5f6"

	// filterExample is the API's filter syntax, which is how a single object is
	// addressed by something other than its id and how a collection is narrowed.
	filterExample = `name eq \"example\"`

	idAttribute     = "id"
	filterAttribute = "filter"
)

func runExamples(args []string) error {
	flags := flag.NewFlagSet("examples", flag.ExitOnError)
	out := flags.String("out", "", "directory to write examples to")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *out == "" {
		return errors.New("examples: -out is required")
	}

	ctx := context.Background()
	written, kept := 0, 0

	for _, definition := range registry.Resources() {
		typeName := provider.TypeName + "_" + definition.Name
		dir := filepath.Join(*out, "resources", typeName)

		schema, err := resourceSchema(ctx, definition)
		if err != nil {
			return err
		}

		for base, content := range map[string]string{
			"resource.tf": resourceExample(ctx, typeName, definition, schema),
			"import.sh":   importExample(typeName, definition),
		} {
			w, err := writeExample(filepath.Join(dir, base), content)
			if err != nil {
				return err
			}

			written, kept = written+w, kept+1-w
		}
	}

	for _, definition := range registry.DataSources() {
		typeName := provider.TypeName + "_" + definition.Name

		schema, err := dataSourceSchema(ctx, definition)
		if err != nil {
			return err
		}

		w, err := writeExample(
			filepath.Join(*out, "data-sources", typeName, "data-source.tf"),
			dataSourceExample(ctx, typeName, definition, schema),
		)
		if err != nil {
			return err
		}

		written, kept = written+w, kept+1-w
	}

	fmt.Fprintf(os.Stderr, "tfdocs: examples: wrote %d, kept %d already present\n", written, kept)

	return nil
}

// writeExample writes an example unless there is one there already, and reports
// whether it wrote. A file on disk is either hand-written or a previous run's, and
// both are left alone: regenerating never overwrites an explanation.
func writeExample(path, content string) (int, error) {
	switch _, err := os.Stat(path); {
	case err == nil:
		return 0, nil
	case !os.IsNotExist(err):
		return 0, fmt.Errorf("checking %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, fmt.Errorf("creating %s: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return 0, fmt.Errorf("writing %s: %w", path, err)
	}

	return 1, nil
}

// resourceSchema asks a resource for the schema it exposes, which is the generated
// one after the runtime has decorated it.
func resourceSchema(ctx context.Context, definition genresource.Definition) (map[string]exampleAttribute, error) {
	resp := &resource.SchemaResponse{}
	genresource.NewResource(definition)().Schema(ctx, resource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		return nil, fmt.Errorf("schema of resource %s: %v", definition.Name, resp.Diagnostics.Errors())
	}

	out := map[string]exampleAttribute{}

	for name, attribute := range resp.Schema.Attributes {
		out[name] = exampleAttribute{name: name, attribute: attribute}
	}

	return out, nil
}

func dataSourceSchema(ctx context.Context, definition genresource.DataSourceDefinition) (map[string]exampleAttribute, error) {
	resp := &datasource.SchemaResponse{}
	genresource.NewDataSource(definition)().Schema(ctx, datasource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		return nil, fmt.Errorf("schema of data source %s: %v", definition.Name, resp.Diagnostics.Errors())
	}

	out := map[string]exampleAttribute{}

	for name, attribute := range resp.Schema.Attributes {
		out[name] = exampleAttribute{name: name, attribute: attribute}
	}

	return out, nil
}

// exampleAttribute is one attribute of either kind of schema. A resource attribute
// and a data source attribute are different Go types with the same shape, so what
// an example needs of them is asked for through an interface both satisfy.
type exampleAttribute struct {
	name      string
	attribute any
}

type settable interface {
	IsRequired() bool
	IsOptional() bool
	GetType() attr.Type
}

type stringValidators interface {
	StringValidators() []validator.String
}

func (a exampleAttribute) required() bool {
	value, ok := a.attribute.(settable)

	return ok && value.IsRequired()
}

func (a exampleAttribute) valueType() attr.Type {
	value, ok := a.attribute.(settable)
	if !ok {
		return nil
	}

	return value.GetType()
}

// enum lists the values a validator restricts the attribute to. The framework
// keeps validators to itself apart from the sentence each one describes itself
// with, so the values come back out of that sentence.
func (a exampleAttribute) enum(ctx context.Context) []string {
	value, ok := a.attribute.(stringValidators)
	if !ok {
		return nil
	}

	for _, entry := range value.StringValidators() {
		description := entry.Description(ctx)

		open := strings.Index(description, "[")
		if !strings.Contains(description, "must be one of") || open < 0 {
			continue
		}

		var values []string

		for _, field := range strings.Fields(strings.Trim(description[open:], "[]")) {
			values = append(values, strings.Trim(field, `"`))
		}

		if len(values) > 0 {
			return values
		}
	}

	return nil
}

// nested returns the attributes of a nested object, and whether the attribute is
// one. The framework's own interface for this is internal, so the shapes the
// provider actually produces are matched instead.
func (a exampleAttribute) nested() (map[string]exampleAttribute, bool) {
	var attributes map[string]exampleAttribute

	switch typed := a.attribute.(type) {
	case rschema.SingleNestedAttribute:
		attributes = resourceAttributes(typed.Attributes)
	case rschema.ListNestedAttribute:
		attributes = resourceAttributes(typed.NestedObject.Attributes)
	case rschema.SetNestedAttribute:
		attributes = resourceAttributes(typed.NestedObject.Attributes)
	case dschema.SingleNestedAttribute:
		attributes = dataSourceAttributes(typed.Attributes)
	case dschema.ListNestedAttribute:
		attributes = dataSourceAttributes(typed.NestedObject.Attributes)
	case dschema.SetNestedAttribute:
		attributes = dataSourceAttributes(typed.NestedObject.Attributes)
	default:
		return nil, false
	}

	return attributes, true
}

// list reports whether a nested attribute holds a collection of objects rather
// than one, which is the difference between a block and a list of them.
func (a exampleAttribute) list() bool {
	switch a.attribute.(type) {
	case rschema.ListNestedAttribute, rschema.SetNestedAttribute,
		dschema.ListNestedAttribute, dschema.SetNestedAttribute:
		return true
	default:
		return false
	}
}

func resourceAttributes(in map[string]rschema.Attribute) map[string]exampleAttribute {
	out := make(map[string]exampleAttribute, len(in))

	for name, attribute := range in {
		out[name] = exampleAttribute{name: name, attribute: attribute}
	}

	return out
}

func dataSourceAttributes(in map[string]dschema.Attribute) map[string]exampleAttribute {
	out := make(map[string]exampleAttribute, len(in))

	for name, attribute := range in {
		out[name] = exampleAttribute{name: name, attribute: attribute}
	}

	return out
}

// resourceExample renders a resource block holding what the object cannot be
// created without.
func resourceExample(ctx context.Context, typeName string, definition genresource.Definition, attributes map[string]exampleAttribute) string {
	var out strings.Builder

	if definition.RawFeature != "" {
		fmt.Fprintf(&out,
			"# This object's whole configuration is a JSON document the API declares no\n"+
				"# shape for, so the resource is off until %s is set on the\n"+
				"# provider. That opts out of the provider's compatibility guarantees.\n",
			genresource.RawOptInArgument(definition.RawFeature))
	}

	if !definition.Update.Supported() {
		out.WriteString("# The API cannot update one of these, so any change replaces it.\n")
	}

	fmt.Fprintf(&out, "resource %q %q {\n", typeName, exampleLabel)
	out.WriteString(indent(arguments(ctx, attributes, "", definition.RawJSONAttributes, definition.VariantBlocks), "  "))
	out.WriteString("}\n")

	return out.String()
}

// importExample renders the import command, taking the shape of the id from the
// definition the runtime parses it with.
func importExample(typeName string, definition genresource.Definition) string {
	names := definition.IdentityAttributes()
	values := make([]string, 0, len(names))

	for _, name := range names {
		values = append(values, "<"+name+">")
	}

	var out strings.Builder

	if len(names) > 1 {
		fmt.Fprintf(&out,
			"# This object lives under another, so importing it takes the ids of both, in\n"+
				"# the order the API's own path has them: %s.\n",
			strings.Join(names, ", "))
	}

	fmt.Fprintf(&out, "terraform import %s.%s %s\n", typeName, exampleLabel, strings.Join(values, "/"))

	return out.String()
}

// dataSourceExample renders however many data blocks it takes to show the ways the
// object can be addressed: by id or by filter for a single object, whole or
// narrowed for a collection.
func dataSourceExample(ctx context.Context, typeName string, definition genresource.DataSourceDefinition, attributes map[string]exampleAttribute) string {
	var out strings.Builder

	if definition.RawFeature != "" {
		fmt.Fprintf(&out,
			"# This object's whole configuration is a JSON document the API declares no\n"+
				"# shape for, so the data source is off until %s is set\n"+
				"# on the provider.\n",
			genresource.RawOptInArgument(definition.RawFeature))
	}

	required := requiredArguments(ctx, attributes, "", definition.RawJSONAttributes, definition.VariantBlocks)

	_, byID := attributes[idAttribute]
	_, byFilter := attributes[filterAttribute]

	switch {
	case required != "":
		// Something has to be given — the id of the object or of its parent — so
		// there is only one way to write this.
		fmt.Fprintf(&out, "data %q %q {\n", typeName, exampleLabel)
		out.WriteString(indent(required, "  "))
		out.WriteString("}\n")

	case byID && byFilter:
		out.WriteString("# A single object is addressed either by its id...\n")
		fmt.Fprintf(&out, "data %q \"by_id\" {\n  id = %q\n}\n\n", typeName, idExample)
		out.WriteString("# ...or by a filter, which has to match exactly one object.\n")
		fmt.Fprintf(&out, "data %q \"by_filter\" {\n  filter = \"%s\"\n}\n", typeName, filterExample)

	case byID:
		fmt.Fprintf(&out, "data %q %q {\n  id = %q\n}\n", typeName, exampleLabel, idExample)

	case byFilter:
		out.WriteString("# Every page is walked, so data holds the whole collection. A filter narrows\n" +
			"# it; first or after ask for one page instead.\n")
		fmt.Fprintf(&out, "data %q %q {\n  filter = \"%s\"\n}\n", typeName, exampleLabel, filterExample)

	default:
		fmt.Fprintf(&out, "data %q %q {}\n", typeName, exampleLabel)
	}

	return out.String()
}

// requiredArguments is arguments, for callers that only want to know whether there
// were any.
func requiredArguments(ctx context.Context, attributes map[string]exampleAttribute, prefix string, rawJSON, variants []string) string {
	return arguments(ctx, attributes, prefix, rawJSON, variants)
}

// arguments renders a line per argument that has to be set: every required one,
// plus the first of any set of alternative blocks, since "exactly one of these" is
// not satisfied by leaving all of them out.
//
// Lines are aligned the way terraform fmt aligns them — a run of single-line
// arguments shares a column, and a block interrupts the run — so that a generated
// example passes the formatting check.
func arguments(ctx context.Context, attributes map[string]exampleAttribute, prefix string, rawJSON, variants []string) string {
	names := slices.Sorted(maps.Keys(attributes))
	chosen := make([]exampleAttribute, 0, len(names))
	variantTaken := false

	for _, name := range names {
		attribute := attributes[name]
		path := join(prefix, name)

		switch {
		case attribute.required():
			chosen = append(chosen, attribute)
		case slices.Contains(variants, path) && !variantTaken:
			chosen = append(chosen, attribute)
			variantTaken = true
		}
	}

	rendered := make([]string, 0, len(chosen))
	widths := make([]int, len(chosen))

	for i, attribute := range chosen {
		rendered = append(rendered, value(ctx, attribute, join(prefix, attribute.name), rawJSON, variants))
		widths[i] = len(attribute.name)
	}

	// A block spans lines, so it ends the run of arguments sharing a column.
	for i := range chosen {
		if strings.Contains(rendered[i], "\n") {
			widths[i] = 0

			continue
		}

		for j := i - 1; j >= 0 && widths[j] > 0 && !strings.Contains(rendered[j], "\n"); j-- {
			widths[i] = max(widths[i], len(chosen[j].name))
		}

		for j := i + 1; j < len(chosen) && !strings.Contains(rendered[j], "\n"); j++ {
			widths[i] = max(widths[i], len(chosen[j].name))
		}
	}

	var out strings.Builder

	for i, attribute := range chosen {
		fmt.Fprintf(&out, "%-*s = %s\n", widths[i], attribute.name, rendered[i])
	}

	return out.String()
}

// value renders something of the attribute's type: the first value the schema
// allows where it restricts them, the required arguments again for a nested
// object, and otherwise a placeholder named after the attribute — a value nobody
// could mistake for a working one, which is the point.
func value(ctx context.Context, attribute exampleAttribute, path string, rawJSON, variants []string) string {
	if nested, ok := attribute.nested(); ok {
		body := "{\n" + indent(arguments(ctx, nested, path, rawJSON, variants), "  ") + "}"

		if attribute.list() {
			return "[" + body + "]"
		}

		return body
	}

	if slices.Contains(rawJSON, path) {
		// The attribute carries a JSON document as a string. jsonencode is how a
		// configuration writes one without escaping it by hand.
		return "jsonencode({})"
	}

	if allowed := attribute.enum(ctx); len(allowed) > 0 {
		return fmt.Sprintf("%q", allowed[0])
	}

	switch typed := attribute.valueType().(type) {
	case basetypes.BoolType:
		return "false"

	case basetypes.Int64Type, basetypes.Float64Type, basetypes.NumberType:
		return "1"

	case basetypes.ListType:
		return "[" + scalar(singular(attribute.name), typed.ElemType) + "]"

	case basetypes.SetType:
		return "[" + scalar(singular(attribute.name), typed.ElemType) + "]"

	case basetypes.MapType:
		return "{}"

	default:
		return scalar(attribute.name, attribute.valueType())
	}
}

// scalar renders one value of an element type, named after the attribute holding
// it.
func scalar(name string, elem attr.Type) string {
	switch elem.(type) {
	case basetypes.BoolType:
		return "false"
	case basetypes.Int64Type, basetypes.Float64Type, basetypes.NumberType:
		return "1"
	}

	switch {
	case name == idAttribute, strings.HasSuffix(name, "_id"):
		return fmt.Sprintf("%q", idExample)

	case name == "name", name == "display_name", name == "group_name":
		return fmt.Sprintf("%q", exampleLabel)

	case name == "description":
		return `"Managed by Terraform"`

	default:
		return fmt.Sprintf("%q", "<"+name+">")
	}
}

// singular trims a plural attribute name, so that an element of a collection reads
// as one of whatever the collection holds: servers holds a server.
func singular(name string) string {
	if len(name) > 3 && strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss") {
		return strings.TrimSuffix(name, "s")
	}

	return name
}

func join(prefix, name string) string {
	if prefix == "" {
		return name
	}

	return prefix + "." + name
}

func indent(block, with string) string {
	if block == "" {
		return ""
	}

	lines := strings.Split(strings.TrimSuffix(block, "\n"), "\n")

	for i, line := range lines {
		if line != "" {
			lines[i] = with + line
		}
	}

	return strings.Join(lines, "\n") + "\n"
}
