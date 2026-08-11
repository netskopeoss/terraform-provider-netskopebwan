// Package tfjson converts between Terraform values and JSON documents.
//
// The provider is generated from an OpenAPI document, so every attribute name
// is also its JSON field name and the conversion needs no per-resource code.
package tfjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Encode converts a Terraform value into a value ready for json.Marshal.
//
// Null and unknown values report include=false rather than encoding to JSON
// null: the API is told nothing about attributes the configuration left out,
// which is what makes a PATCH of a partially configured object safe.
func Encode(value tftypes.Value) (any, bool, error) {
	if !value.IsKnown() || value.IsNull() {
		return nil, false, nil
	}

	typ := value.Type()

	switch {
	case typ.Is(tftypes.String):
		var out string
		if err := value.As(&out); err != nil {
			return nil, false, err
		}

		return out, true, nil
	case typ.Is(tftypes.Bool):
		var out bool
		if err := value.As(&out); err != nil {
			return nil, false, err
		}

		return out, true, nil
	case typ.Is(tftypes.Number):
		var out big.Float
		if err := value.As(&out); err != nil {
			return nil, false, err
		}

		return encodeNumber(&out), true, nil
	}

	switch typ.(type) {
	case tftypes.Object, tftypes.Map:
		return encodeMapping(value)
	case tftypes.List, tftypes.Set, tftypes.Tuple:
		return encodeSequence(value)
	default:
		return nil, false, fmt.Errorf("cannot encode %s to JSON", typ)
	}
}

func encodeMapping(value tftypes.Value) (any, bool, error) {
	var members map[string]tftypes.Value
	if err := value.As(&members); err != nil {
		return nil, false, err
	}

	out := make(map[string]any, len(members))

	for name, member := range members {
		encoded, include, err := Encode(member)
		if err != nil {
			return nil, false, fmt.Errorf("%s: %w", name, err)
		}

		if include {
			out[name] = encoded
		}
	}

	return out, true, nil
}

func encodeSequence(value tftypes.Value) (any, bool, error) {
	var elements []tftypes.Value
	if err := value.As(&elements); err != nil {
		return nil, false, err
	}

	out := make([]any, 0, len(elements))

	for i, element := range elements {
		// A collection is positional, so an element that would be omitted has
		// to be sent as JSON null to keep the remaining elements in place.
		encoded, _, err := Encode(element)
		if err != nil {
			return nil, false, fmt.Errorf("[%d]: %w", i, err)
		}

		out = append(out, encoded)
	}

	return out, true, nil
}

// encodeNumber keeps whole numbers integral so an int64 field does not reach the
// API as "1e+09".
func encodeNumber(value *big.Float) any {
	if value.IsInt() {
		if asInt, accuracy := value.Int64(); accuracy == big.Exact {
			return asInt
		}
	}

	return json.Number(value.Text('f', -1))
}

// Decode converts a decoded JSON document into a Terraform value of type typ.
// A JSON null, or a document that simply does not carry the field, becomes a
// null Terraform value; fields the type does not describe are dropped.
func Decode(document any, typ tftypes.Type) (tftypes.Value, error) {
	if document == nil {
		return newValue(typ, nil)
	}

	switch {
	case typ.Is(tftypes.String):
		text, ok := document.(string)
		if !ok {
			return tftypes.Value{}, typeMismatch(document, typ)
		}

		return newValue(typ, text)
	case typ.Is(tftypes.Bool):
		flag, ok := document.(bool)
		if !ok {
			return tftypes.Value{}, typeMismatch(document, typ)
		}

		return newValue(typ, flag)
	case typ.Is(tftypes.Number):
		number, err := decodeNumber(document)
		if err != nil {
			return tftypes.Value{}, err
		}

		return newValue(typ, number)
	}

	switch target := typ.(type) {
	case tftypes.Object:
		return decodeObject(document, target)
	case tftypes.Map:
		return decodeMap(document, target)
	case tftypes.List:
		return decodeSequence(document, target, target.ElementType)
	case tftypes.Set:
		return decodeSequence(document, target, target.ElementType)
	case tftypes.Tuple:
		return decodeTuple(document, target)
	default:
		return tftypes.Value{}, fmt.Errorf("cannot decode JSON into %s", typ)
	}
}

func decodeObject(document any, typ tftypes.Object) (tftypes.Value, error) {
	fields, ok := document.(map[string]any)
	if !ok {
		return tftypes.Value{}, typeMismatch(document, typ)
	}

	members := make(map[string]tftypes.Value, len(typ.AttributeTypes))

	for name, attrType := range typ.AttributeTypes {
		member, err := Decode(fields[name], attrType)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("%s: %w", name, err)
		}

		members[name] = member
	}

	return newValue(typ, members)
}

func decodeMap(document any, typ tftypes.Map) (tftypes.Value, error) {
	fields, ok := document.(map[string]any)
	if !ok {
		return tftypes.Value{}, typeMismatch(document, typ)
	}

	members := make(map[string]tftypes.Value, len(fields))

	for name, field := range fields {
		member, err := Decode(field, typ.ElementType)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("%s: %w", name, err)
		}

		members[name] = member
	}

	return newValue(typ, members)
}

func decodeSequence(document any, typ tftypes.Type, elementType tftypes.Type) (tftypes.Value, error) {
	items, ok := document.([]any)
	if !ok {
		return tftypes.Value{}, typeMismatch(document, typ)
	}

	elements := make([]tftypes.Value, 0, len(items))

	for i, item := range items {
		element, err := Decode(item, elementType)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("[%d]: %w", i, err)
		}

		elements = append(elements, element)
	}

	return newValue(typ, elements)
}

func decodeTuple(document any, typ tftypes.Tuple) (tftypes.Value, error) {
	items, ok := document.([]any)
	if !ok {
		return tftypes.Value{}, typeMismatch(document, typ)
	}

	if len(items) != len(typ.ElementTypes) {
		return tftypes.Value{}, fmt.Errorf("expected %d elements for %s, got %d", len(typ.ElementTypes), typ, len(items))
	}

	elements := make([]tftypes.Value, 0, len(items))

	for i, item := range items {
		element, err := Decode(item, typ.ElementTypes[i])
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("[%d]: %w", i, err)
		}

		elements = append(elements, element)
	}

	return newValue(typ, elements)
}

func decodeNumber(document any) (*big.Float, error) {
	switch number := document.(type) {
	case json.Number:
		parsed, _, err := big.ParseFloat(number.String(), 10, big.MaxPrec, big.ToNearestEven)
		if err != nil {
			return nil, fmt.Errorf("parsing number %q: %w", number, err)
		}

		return parsed, nil
	case float64:
		return big.NewFloat(number), nil
	case int64:
		return new(big.Float).SetInt64(number), nil
	default:
		return nil, typeMismatch(document, tftypes.Number)
	}
}

// newValue builds a value only after checking it against typ, so a malformed
// API response surfaces as a diagnostic rather than a panic.
func newValue(typ tftypes.Type, value any) (tftypes.Value, error) {
	if err := tftypes.ValidateValue(typ, value); err != nil {
		return tftypes.Value{}, err
	}

	return tftypes.NewValue(typ, value), nil
}

func typeMismatch(document any, typ tftypes.Type) error {
	return fmt.Errorf("expected %s, got %s", typ, jsonKind(document))
}

func jsonKind(document any) string {
	switch document.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64, int64, json.Number:
		return "number"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%T", document)
	}
}

// Unmarshal decodes an API document, keeping numbers exact so an int64 field
// does not lose precision on the way into Terraform state.
func Unmarshal(raw []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var document any
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return document, nil
}
