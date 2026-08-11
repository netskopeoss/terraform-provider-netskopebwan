package tfschema

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// VariantGroup is a set of sibling blocks of which exactly one stands for the
// form the object takes.
type VariantGroup struct {
	// Path is the dotted path of the object holding the group, empty at the root.
	Path string
	// Names are the sibling blocks, sorted.
	Names []string
}

// Attribute renders the dotted path of one block in the group.
func (g VariantGroup) Attribute(name string) string {
	if g.Path == "" {
		return name
	}

	return g.Path + "." + name
}

// VariantGroups lists every set of alternative blocks in the schema, so a caller
// can check that exactly one of each is set before anything reaches the API.
func (m *Model) VariantGroups() []VariantGroup {
	var out []VariantGroup

	collectGroups(m.Attrs, "", &out)

	return out
}

func collectGroups(attrs map[string]*Node, prefix string, out *[]VariantGroup) {
	var names []string

	for _, name := range slices.Sorted(maps.Keys(attrs)) {
		node := attrs[name]

		if node.Variant {
			names = append(names, name)
		}

		path := name
		if prefix != "" {
			path = prefix + "." + name
		}

		collectGroups(node.Nested, path, out)
	}

	if len(names) > 0 {
		*out = append(*out, VariantGroup{Path: prefix, Names: names})
	}
}

// Reshape rewrites an API document into the shape the schema declares, moving the
// fields of whichever form the object took into the block standing for that form.
//
// The API describes those forms as alternative shapes of the object itself, so it
// sends a link monitor as {"id": ..., "fqdn": ...} where the schema has an fqdn
// block. Which form arrived has to be worked out from the fields present, because
// a form the API declares without a discriminator has nothing else to go on.
func (m *Model) Reshape(document any) any {
	return reshape(m.Attrs, document)
}

func reshape(attrs map[string]*Node, document any) any {
	switch typed := document.(type) {
	case map[string]any:
		return reshapeObject(attrs, typed)
	case []any:
		out := make([]any, 0, len(typed))

		for _, element := range typed {
			out = append(out, reshape(attrs, element))
		}

		return out
	default:
		return document
	}
}

func reshapeObject(attrs map[string]*Node, document map[string]any) map[string]any {
	out := make(map[string]any, len(document))
	maps.Copy(out, document)

	if group := variantNodes(attrs); len(group) > 0 {
		out = nestVariant(group, out)
	}

	for name, node := range attrs {
		if node.Nested == nil {
			continue
		}

		if value, ok := out[name]; ok {
			out[name] = reshape(node.Nested, value)
		}
	}

	return out
}

func variantNodes(attrs map[string]*Node) map[string]*Node {
	var group map[string]*Node

	for name, node := range attrs {
		if !node.Variant {
			continue
		}

		if group == nil {
			group = map[string]*Node{}
		}

		group[name] = node
	}

	return group
}

// nestVariant moves the fields of the best-matching form into its block. Nothing
// moves when no form matches: an object the API sent without any of their fields
// is one this schema cannot place, and inventing a form for it would put values
// into state that never came back.
func nestVariant(group map[string]*Node, document map[string]any) map[string]any {
	best, fields := bestForm(group, document)
	if best == "" {
		return document
	}

	nested := make(map[string]any, len(fields))
	out := make(map[string]any, len(document))
	maps.Copy(out, document)

	for _, name := range fields {
		nested[name] = out[name]
		delete(out, name)
	}

	out[best] = nested

	return out
}

// bestForm picks the block whose fields the document matches most closely. A block
// every one of whose fields is present beats one that only overlaps, so a form
// sharing a field with a longer form does not shadow it.
func bestForm(group map[string]*Node, document map[string]any) (string, []string) {
	var (
		best     string
		bestKeys []string
		complete bool
	)

	for _, name := range slices.Sorted(maps.Keys(group)) {
		var present []string

		for _, child := range group[name].Nested {
			if value, ok := document[fieldName(child, child.Name)]; ok && value != nil {
				present = append(present, fieldName(child, child.Name))
			}
		}

		if len(present) == 0 {
			continue
		}

		slices.Sort(present)

		whole := len(present) == len(group[name].Nested)

		if best == "" || (whole && !complete) || (whole == complete && len(present) > len(bestKeys)) {
			best, bestKeys, complete = name, present, whole
		}
	}

	return best, bestKeys
}

// spliceVariants moves the fields of whichever block is set back out into the
// object, which is the shape the API declared. It reports an error when more than
// one form is set: the API accepts exactly one, and picking for the practitioner
// would silently drop the other.
func spliceVariants(out, blocks map[string]any) error {
	if len(blocks) > 1 {
		return fmt.Errorf("%s are alternatives, so only one of them can be set", strings.Join(slices.Sorted(maps.Keys(blocks)), " and "))
	}

	for name, encoded := range blocks {
		fields, ok := encoded.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected an object, got %T", name, encoded)
		}

		for _, field := range slices.Sorted(maps.Keys(fields)) {
			if _, clash := out[field]; clash {
				return fmt.Errorf("%s.%s collides with a field of the same name on its parent", name, field)
			}

			out[field] = fields[field]
		}
	}

	return nil
}
