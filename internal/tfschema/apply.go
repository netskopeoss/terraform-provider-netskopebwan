package tfschema

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/tfjson"
)

// Mode selects how an API document combines with the value Terraform already
// holds.
type Mode int

const (
	// AfterWrite keeps every value the configuration planned, so Terraform sees
	// the result it asked for; the API only fills in what was left open.
	// Terraform fails an apply whose result differs from its plan, so a create
	// or update response can never overrule the configuration.
	AfterWrite Mode = iota
	// AfterRead trusts the API so that drift becomes visible, and falls back to
	// what is already in state only for fields the API does not return at all —
	// write-only credentials, for instance, which would otherwise look removed
	// on every refresh.
	AfterRead
)

// Apply combines an API document with the value Terraform holds — the plan for
// AfterWrite, prior state for AfterRead — into a complete state value.
func (m *Model) Apply(mode Mode, document any, current tftypes.Value) (tftypes.Value, error) {
	object, ok := m.Type.(tftypes.Object)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("schema type %s is not an object", m.Type)
	}

	// The API sends the fields of whichever form an object took beside its other
	// fields; the schema holds them in a block per form. Reshaping first keeps the
	// merge below working in terms of the schema alone.
	return applyObject(m.Attrs, object, mode, m.Reshape(document), current)
}

// fieldName is the key a value arrives under in an API document.
func fieldName(node *Node, name string) string {
	if node == nil || node.Field == "" {
		return name
	}

	return node.Field
}

func applyObject(attrs map[string]*Node, typ tftypes.Object, mode Mode, document any, current tftypes.Value) (tftypes.Value, error) {
	fields, _ := document.(map[string]any)
	members := objectMembers(current)

	out := make(map[string]tftypes.Value, len(typ.AttributeTypes))

	for name, attrType := range typ.AttributeTypes {
		value, err := applyAttribute(attrs[name], attrType, mode, fields[fieldName(attrs[name], name)], members[name])
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("%s: %w", name, err)
		}

		out[name] = value
	}

	return newValue(typ, out)
}

// applyAttribute resolves one attribute. A JSON null and an absent field are
// treated alike: both mean the API said nothing about the attribute.
func applyAttribute(node *Node, typ tftypes.Type, mode Mode, raw any, current tftypes.Value) (tftypes.Value, error) {
	if node == nil {
		// An attribute the model does not describe cannot be merged, so the API
		// is the only source for it.
		return tfjson.Decode(raw, typ)
	}

	if node.Nested != nil {
		return applyNested(node, typ, mode, raw, current)
	}

	if mode == AfterRead {
		if raw == nil {
			return reuse(current, typ)
		}

		return decodeLeaf(node, typ, raw)
	}

	// A value the practitioner declared has to survive verbatim, including an
	// explicit null: Terraform rejects a result that differs from its plan.
	if node.Configurable() && !node.Computed {
		return reuse(current, typ)
	}

	if held(current) && !current.IsNull() {
		return reuse(current, typ)
	}

	return decodeLeaf(node, typ, raw)
}

func applyNested(node *Node, typ tftypes.Type, mode Mode, raw any, current tftypes.Value) (tftypes.Value, error) {
	switch target := typ.(type) {
	case tftypes.Object:
		return applyNestedObject(node.Nested, target, mode, raw, current)
	case tftypes.List:
		return applyNestedList(node, typ, target.ElementType, mode, raw, current)
	case tftypes.Map:
		return applyNestedMap(node, typ, target.ElementType, mode, raw, current)
	default:
		// No schema generated from the BWAN spec uses a nested set or tuple, and
		// merging one by position would silently reorder it.
		return tftypes.Value{}, fmt.Errorf("nested attribute of type %s is not supported", typ)
	}
}

func applyNestedObject(attrs map[string]*Node, typ tftypes.Object, mode Mode, raw any, current tftypes.Value) (tftypes.Value, error) {
	if _, ok := raw.(map[string]any); ok {
		return applyObject(attrs, typ, mode, raw, current)
	}

	if mode == AfterRead {
		return reuse(current, typ)
	}

	// The API said nothing, so only the planned object remains; recursing fills
	// its computed members with null rather than leaving them unknown.
	if held(current) && !current.IsNull() {
		return applyObject(attrs, typ, mode, nil, current)
	}

	return nullOf(typ)
}

func applyNestedList(node *Node, typ, elementType tftypes.Type, mode Mode, raw any, current tftypes.Value) (tftypes.Value, error) {
	elementObject, ok := elementType.(tftypes.Object)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("nested list of %s is not supported", elementType)
	}

	items, hasItems := raw.([]any)
	currentElements, hasCurrent := sequenceElements(current)

	// Positional alignment only makes sense when both sides agree on the length;
	// otherwise the side that is not driving the merge contributes nothing.
	if mode == AfterRead {
		if !hasItems {
			return reuse(current, typ)
		}

		return mergeList(node.Nested, typ, elementObject, mode, items, sameLength(currentElements, len(items)))
	}

	// On write the configuration decides how many elements there are.
	if !hasCurrent {
		if !hasItems {
			return nullOf(typ)
		}

		return mergeList(node.Nested, typ, elementObject, mode, items, nil)
	}

	return mergeList(node.Nested, typ, elementObject, mode, sameLength(items, len(currentElements)), currentElements)
}

// mergeList merges two positionally aligned slices. Either side may be nil where
// it has nothing to contribute; the longer one sets the result length.
func mergeList(
	nested map[string]*Node,
	typ tftypes.Type,
	elementType tftypes.Object,
	mode Mode,
	items []any,
	currents []tftypes.Value,
) (tftypes.Value, error) {
	out := make([]tftypes.Value, 0, max(len(items), len(currents)))

	for i := range max(len(items), len(currents)) {
		var item any
		if i < len(items) {
			item = items[i]
		}

		var current tftypes.Value
		if i < len(currents) {
			current = currents[i]
		}

		value, err := applyNestedObject(nested, elementType, mode, item, current)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("[%d]: %w", i, err)
		}

		out = append(out, value)
	}

	return newValue(typ, out)
}

func sameLength[T any](values []T, length int) []T {
	if len(values) != length {
		return nil
	}

	return values
}

func keySet[T any](values map[string]T) map[string]struct{} {
	out := make(map[string]struct{}, len(values))

	for key := range values {
		out[key] = struct{}{}
	}

	return out
}

func applyNestedMap(node *Node, typ, elementType tftypes.Type, mode Mode, raw any, current tftypes.Value) (tftypes.Value, error) {
	elementObject, ok := elementType.(tftypes.Object)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("nested map of %s is not supported", elementType)
	}

	fields, hasFields := raw.(map[string]any)
	currentMembers := objectMembers(current)

	// The keys of the result come from whichever side is authoritative: the
	// configuration on write, the API on read.
	var keys map[string]struct{}

	switch {
	case mode == AfterWrite && len(currentMembers) > 0:
		keys = keySet(currentMembers)
	case hasFields:
		keys = keySet(fields)
	case mode == AfterRead:
		return reuse(current, typ)
	default:
		return nullOf(typ)
	}

	out := make(map[string]tftypes.Value, len(keys))

	for key := range keys {
		value, err := applyNestedObject(node.Nested, elementObject, mode, fields[key], currentMembers[key])
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("%s: %w", key, err)
		}

		out[key] = value
	}

	return newValue(typ, out)
}

func decodeLeaf(node *Node, typ tftypes.Type, raw any) (tftypes.Value, error) {
	if !node.RawJSON {
		return tfjson.Decode(raw, typ)
	}

	if raw == nil {
		return nullOf(typ)
	}

	// The API exchanges the document itself; state keeps it as a string. The
	// attribute uses jsontypes.Normalized, so re-encoding it here cannot make
	// Terraform see a difference that is only formatting.
	encoded, err := json.Marshal(raw)
	if err != nil {
		return tftypes.Value{}, fmt.Errorf("encoding embedded JSON document: %w", err)
	}

	return newValue(typ, string(encoded))
}

// held reports whether Terraform has a usable value: a member read out of a null
// object carries no type and cannot be reused.
func held(value tftypes.Value) bool {
	return value.Type() != nil && value.IsKnown()
}

// reuse returns the value Terraform already holds, or null when there is none
// that fits typ.
func reuse(value tftypes.Value, typ tftypes.Type) (tftypes.Value, error) {
	if held(value) && value.Type().Equal(typ) {
		return value, nil
	}

	return nullOf(typ)
}

func nullOf(typ tftypes.Type) (tftypes.Value, error) {
	return newValue(typ, nil)
}

func objectMembers(value tftypes.Value) map[string]tftypes.Value {
	if !held(value) || value.IsNull() {
		return nil
	}

	var members map[string]tftypes.Value
	if err := value.As(&members); err != nil {
		return nil
	}

	return members
}

func sequenceElements(value tftypes.Value) ([]tftypes.Value, bool) {
	if !held(value) || value.IsNull() {
		return nil, false
	}

	var elements []tftypes.Value
	if err := value.As(&elements); err != nil {
		return nil, false
	}

	return elements, true
}

func newValue(typ tftypes.Type, value any) (tftypes.Value, error) {
	if err := tftypes.ValidateValue(typ, value); err != nil {
		return tftypes.Value{}, err
	}

	return tftypes.NewValue(typ, value), nil
}
