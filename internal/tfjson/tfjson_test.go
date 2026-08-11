package tfjson

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestEncodeOmitsNullAndUnknown(t *testing.T) {
	for name, value := range map[string]tftypes.Value{
		"null":    tftypes.NewValue(tftypes.String, nil),
		"unknown": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	} {
		t.Run(name, func(t *testing.T) {
			encoded, include, err := Encode(value)

			require.NoError(t, err)
			require.False(t, include, "an attribute the configuration left out must not reach the API")
			require.Nil(t, encoded)
		})
	}
}

func TestEncodeObjectDropsUnsetMembers(t *testing.T) {
	typ := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name":        tftypes.String,
		"description": tftypes.String,
		"enabled":     tftypes.Bool,
	}}

	encoded, include, err := Encode(tftypes.NewValue(typ, map[string]tftypes.Value{
		"name":        tftypes.NewValue(tftypes.String, "segment"),
		"description": tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, true),
	}))

	require.NoError(t, err)
	require.True(t, include)
	require.Equal(t, map[string]any{"name": "segment", "enabled": true}, encoded)
}

func TestEncodeKeepsWholeNumbersIntegral(t *testing.T) {
	encoded, _, err := Encode(tftypes.NewValue(tftypes.Number, big.NewFloat(1e9)))
	require.NoError(t, err)

	serialised, err := json.Marshal(encoded)
	require.NoError(t, err)
	require.JSONEq(t, "1000000000", string(serialised))

	encoded, _, err = Encode(tftypes.NewValue(tftypes.Number, big.NewFloat(1.5)))
	require.NoError(t, err)

	serialised, err = json.Marshal(encoded)
	require.NoError(t, err)
	require.JSONEq(t, "1.5", string(serialised))
}

func TestEncodeSequenceKeepsPositions(t *testing.T) {
	typ := tftypes.List{ElementType: tftypes.String}

	encoded, include, err := Encode(tftypes.NewValue(typ, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "a"),
		tftypes.NewValue(tftypes.String, nil),
		tftypes.NewValue(tftypes.String, "c"),
	}))

	require.NoError(t, err)
	require.True(t, include)
	require.Equal(t, []any{"a", nil, "c"}, encoded)
}

func TestDecodeAbsentFieldsBecomeNull(t *testing.T) {
	typ := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name":        tftypes.String,
		"description": tftypes.String,
	}}

	value, err := Decode(map[string]any{"name": "segment"}, typ)
	require.NoError(t, err)

	var members map[string]tftypes.Value
	require.NoError(t, value.As(&members))
	require.Equal(t, tftypes.NewValue(tftypes.String, "segment"), members["name"])
	require.True(t, members["description"].IsNull())
}

func TestDecodeIgnoresFieldsTheSchemaDoesNotDescribe(t *testing.T) {
	typ := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}

	value, err := Decode(map[string]any{"name": "segment", "added_by_a_newer_api": "surprise"}, typ)
	require.NoError(t, err)

	var members map[string]tftypes.Value
	require.NoError(t, value.As(&members))
	require.Len(t, members, 1)
}

func TestDecodeRejectsTypeMismatch(t *testing.T) {
	_, err := Decode([]any{"a"}, tftypes.String)

	require.ErrorContains(t, err, "expected tftypes.String, got array")
}

func TestDecodeKeepsLargeIntegersExact(t *testing.T) {
	document, err := Unmarshal([]byte(`{"overlay_id": 9007199254740993}`))
	require.NoError(t, err)

	typ := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"overlay_id": tftypes.Number}}

	value, err := Decode(document, typ)
	require.NoError(t, err)

	encoded, _, err := Encode(value)
	require.NoError(t, err)

	serialised, err := json.Marshal(encoded)
	require.NoError(t, err)
	require.JSONEq(t, `{"overlay_id": 9007199254740993}`, string(serialised))
}

func TestRoundTripNestedDocument(t *testing.T) {
	typ := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name": tftypes.String,
		"rules": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"port":    tftypes.Number,
			"enabled": tftypes.Bool,
		}}},
		"labels": tftypes.List{ElementType: tftypes.String},
	}}

	document, err := Unmarshal([]byte(`{
		"name": "policy",
		"rules": [{"port": 443, "enabled": true}, {"port": 80, "enabled": false}],
		"labels": ["a", "b"]
	}`))
	require.NoError(t, err)

	value, err := Decode(document, typ)
	require.NoError(t, err)

	encoded, include, err := Encode(value)
	require.NoError(t, err)
	require.True(t, include)

	serialised, err := json.Marshal(encoded)
	require.NoError(t, err)

	require.JSONEq(t, `{
		"name": "policy",
		"rules": [{"port": 443, "enabled": true}, {"port": 80, "enabled": false}],
		"labels": ["a", "b"]
	}`, string(serialised))
}
