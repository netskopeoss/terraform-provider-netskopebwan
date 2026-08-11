// Package tfschema turns a generated terraform-plugin-framework schema into a
// walkable model, so whole objects move between Terraform and the API without
// per-resource code.
package tfschema

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Node describes one attribute of a schema.
type Node struct {
	Name string
	// Field is the name the API knows the attribute by. It differs from Name only
	// where Terraform reserves the API's name, e.g. a cloud account's `provider`.
	Field    string
	Required bool
	Optional bool
	Computed bool
	Type     tftypes.Type
	// Nested holds the attributes of a nested object, keyed by name, for single,
	// list, set and map nested attributes. It is nil for every other kind.
	Nested map[string]*Node
	// RawJSON marks a string attribute holding an embedded JSON document: the
	// API exchanges the document itself, Terraform stores it as a string.
	RawJSON bool
	// Variant marks a nested object standing for one of the forms its parent can
	// take. The API declares the forms as alternative shapes of the parent itself,
	// which Terraform has no type for, so each form becomes a block and exactly one
	// of a parent's blocks is set. A request splices the set block's fields back
	// into the parent; a response is reshaped the other way.
	Variant bool
}

// Configurable reports whether a practitioner can set the attribute. A purely
// computed attribute is server owned and never sent to the API.
func (n *Node) Configurable() bool {
	return n.Required || n.Optional
}

// Model is a schema prepared for moving values in and out of the API.
type Model struct {
	Type  tftypes.Type
	Attrs map[string]*Node
}

// FromResource builds the model for a generated resource schema. rawJSON holds
// dot-separated paths of string attributes carrying an embedded JSON document, and
// variants those of the blocks standing for the forms their parent can take.
func FromResource(ctx context.Context, schema rschema.Schema, rawJSON, variants []string, fields map[string]string) *Model {
	return &Model{
		Type:  schema.Type().TerraformType(ctx),
		Attrs: nodes(ctx, schema.Attributes, resourceChildren, parsePaths(rawJSON), parsePaths(variants), apiNames(fields)),
	}
}

// FromDataSource builds the model for a generated data source schema.
func FromDataSource(ctx context.Context, schema dschema.Schema, rawJSON, variants []string, fields map[string]string) *Model {
	return &Model{
		Type:  schema.Type().TerraformType(ctx),
		Attrs: nodes(ctx, schema.Attributes, dataSourceChildren, parsePaths(rawJSON), parsePaths(variants), apiNames(fields)),
	}
}

// apiNames inverts the rename map, which is written the way the configuration
// declares it: from the API's name to Terraform's.
func apiNames(fields map[string]string) map[string]string {
	out := make(map[string]string, len(fields))

	for from, to := range fields {
		out[to] = from
	}

	return out
}

// Validate reports the attributes whose shape this package cannot move between
// Terraform and JSON. Nested objects, lists of objects and maps of objects are
// supported; a nested set or tuple is not, because merging one by position would
// silently reorder it.
func (m *Model) Validate() error {
	var unsupported []string

	validateNodes(m.Attrs, "", &unsupported)

	if len(unsupported) > 0 {
		return fmt.Errorf("unsupported attribute(s): %s", strings.Join(unsupported, ", "))
	}

	return nil
}

func validateNodes(attrs map[string]*Node, prefix string, unsupported *[]string) {
	for _, name := range slices.Sorted(maps.Keys(attrs)) {
		node := attrs[name]

		path := name
		if prefix != "" {
			path = prefix + "." + name
		}

		if node.Nested != nil && !supportedNesting(node.Type) {
			*unsupported = append(*unsupported, fmt.Sprintf("%s (%s)", path, node.Type))
		}

		validateNodes(node.Nested, path, unsupported)
	}
}

func supportedNesting(typ tftypes.Type) bool {
	switch target := typ.(type) {
	case tftypes.Object:
		return true
	case tftypes.List:
		_, ok := target.ElementType.(tftypes.Object)

		return ok
	case tftypes.Map:
		_, ok := target.ElementType.(tftypes.Object)

		return ok
	default:
		return false
	}
}

// ConfigurableNames lists, in order, the attributes a practitioner can set.
func (m *Model) ConfigurableNames() []string {
	var out []string

	for name, node := range m.Attrs {
		if node.Configurable() {
			out = append(out, name)
		}
	}

	slices.Sort(out)

	return out
}

// attribute is the part of the framework's attribute interface the walk needs.
// Resource and data source attributes both satisfy it.
type attribute interface {
	GetType() attr.Type
	IsComputed() bool
	IsOptional() bool
	IsRequired() bool
}

func resourceChildren(a rschema.Attribute) map[string]rschema.Attribute {
	switch nested := a.(type) {
	case rschema.SingleNestedAttribute:
		return nested.Attributes
	case rschema.ListNestedAttribute:
		return nested.NestedObject.Attributes
	case rschema.SetNestedAttribute:
		return nested.NestedObject.Attributes
	case rschema.MapNestedAttribute:
		return nested.NestedObject.Attributes
	default:
		return nil
	}
}

func dataSourceChildren(a dschema.Attribute) map[string]dschema.Attribute {
	switch nested := a.(type) {
	case dschema.SingleNestedAttribute:
		return nested.Attributes
	case dschema.ListNestedAttribute:
		return nested.NestedObject.Attributes
	case dschema.SetNestedAttribute:
		return nested.NestedObject.Attributes
	case dschema.MapNestedAttribute:
		return nested.NestedObject.Attributes
	default:
		return nil
	}
}

func nodes[A attribute](
	ctx context.Context,
	attrs map[string]A,
	children func(A) map[string]A,
	rawJSON pathTree,
	variants pathTree,
	apiNames map[string]string,
) map[string]*Node {
	if len(attrs) == 0 {
		return nil
	}

	out := make(map[string]*Node, len(attrs))

	for name, a := range attrs {
		field := name
		if renamed, ok := apiNames[name]; ok {
			field = renamed
		}

		node := &Node{
			Name:     name,
			Field:    field,
			Required: a.IsRequired(),
			Optional: a.IsOptional(),
			Computed: a.IsComputed(),
			Type:     a.GetType().TerraformType(ctx),
			RawJSON:  rawJSON.isLeaf(name),
			Variant:  variants.isLeaf(name),
		}

		node.Nested = nodes(ctx, children(a), children, rawJSON[name], variants[name], apiNames)

		out[name] = node
	}

	return out
}

// pathTree indexes a list of dot-separated attribute paths, so a walk over the
// schema can ask whether the attribute it is on was named.
type pathTree map[string]pathTree

func parsePaths(paths []string) pathTree {
	root := pathTree{}

	for _, path := range paths {
		node := root

		for segment := range strings.SplitSeq(path, ".") {
			next, ok := node[segment]
			if !ok {
				next = pathTree{}
				node[segment] = next
			}

			node = next
		}
	}

	return root
}

func (p pathTree) isLeaf(name string) bool {
	child, ok := p[name]

	return ok && len(child) == 0
}
