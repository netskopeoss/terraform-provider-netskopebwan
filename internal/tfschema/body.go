package tfschema

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/tfjson"
)

// Body builds the JSON document for a write request from a Terraform value.
//
// Three kinds of attribute are left out: purely computed ones, which the server
// owns; the ones named in skip, which travel in the URL rather than the body;
// and the ones the configuration did not set, so that a PATCH only ever carries
// what the practitioner actually declared.
func (m *Model) Body(value tftypes.Value, skip map[string]bool) (map[string]any, error) {
	body, _, err := encodeObject(m.Attrs, value, skip)

	return body, err
}

// Query renders the configurable attributes of a value as query parameters,
// which is how a data source passes its filters to the API.
func (m *Model) Query(value tftypes.Value, skip map[string]bool) (map[string][]string, error) {
	document, _, err := encodeObject(m.Attrs, value, skip)
	if err != nil {
		return nil, err
	}

	out := map[string][]string{}

	for name, encoded := range document {
		values, err := queryValues(encoded)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}

		out[name] = values
	}

	return out, nil
}

func queryValues(encoded any) ([]string, error) {
	switch typed := encoded.(type) {
	case []any:
		out := make([]string, 0, len(typed))

		for _, item := range typed {
			values, err := queryValues(item)
			if err != nil {
				return nil, err
			}

			out = append(out, values...)
		}

		return out, nil
	case map[string]any:
		return nil, errors.New("an object cannot be a query parameter")
	default:
		return []string{fmt.Sprintf("%v", typed)}, nil
	}
}

func encodeObject(attrs map[string]*Node, value tftypes.Value, skip map[string]bool) (map[string]any, bool, error) {
	if !value.IsKnown() || value.IsNull() {
		return nil, false, nil
	}

	var members map[string]tftypes.Value
	if err := value.As(&members); err != nil {
		return nil, false, err
	}

	out := make(map[string]any, len(members))

	// A block standing for one of the object's forms is held back: the API declared
	// the form as a shape of the object itself, so the block's fields belong beside
	// its siblings rather than under its name.
	var blocks map[string]any

	for name, member := range members {
		if skip[name] {
			continue
		}

		node, ok := attrs[name]
		if !ok || !node.Configurable() {
			continue
		}

		encoded, include, err := encodeNode(node, member)
		if err != nil {
			return nil, false, fmt.Errorf("%s: %w", name, err)
		}

		if !include {
			continue
		}

		if node.Variant {
			if blocks == nil {
				blocks = map[string]any{}
			}

			blocks[name] = encoded

			continue
		}

		out[node.Field] = encoded
	}

	if err := spliceVariants(out, blocks); err != nil {
		return nil, false, err
	}

	return out, true, nil
}

func encodeNode(node *Node, value tftypes.Value) (any, bool, error) {
	if !value.IsKnown() || value.IsNull() {
		return nil, false, nil
	}

	if node.RawJSON {
		var text string
		if err := value.As(&text); err != nil {
			return nil, false, err
		}

		document, err := tfjson.Unmarshal([]byte(text))
		if err != nil {
			return nil, false, err
		}

		return document, true, nil
	}

	if node.Nested == nil {
		return tfjson.Encode(value)
	}

	switch value.Type().(type) {
	case tftypes.Object:
		return encodeObject(node.Nested, value, nil)
	case tftypes.List, tftypes.Set, tftypes.Tuple:
		return encodeNestedSequence(node, value)
	case tftypes.Map:
		return encodeNestedMapping(node, value)
	default:
		return nil, false, fmt.Errorf("cannot encode nested attribute of type %s", value.Type())
	}
}

func encodeNestedSequence(node *Node, value tftypes.Value) (any, bool, error) {
	var elements []tftypes.Value
	if err := value.As(&elements); err != nil {
		return nil, false, err
	}

	out := make([]any, 0, len(elements))

	for i, element := range elements {
		encoded, include, err := encodeObject(node.Nested, element, nil)
		if err != nil {
			return nil, false, fmt.Errorf("[%d]: %w", i, err)
		}

		if !include {
			// Collections are positional, so an unset element still has to
			// occupy its slot.
			out = append(out, nil)

			continue
		}

		out = append(out, encoded)
	}

	return out, true, nil
}

func encodeNestedMapping(node *Node, value tftypes.Value) (any, bool, error) {
	var members map[string]tftypes.Value
	if err := value.As(&members); err != nil {
		return nil, false, err
	}

	out := make(map[string]any, len(members))

	for name, member := range members {
		encoded, include, err := encodeObject(node.Nested, member, nil)
		if err != nil {
			return nil, false, fmt.Errorf("%s: %w", name, err)
		}

		if include {
			out[name] = encoded
		}
	}

	return out, true, nil
}
