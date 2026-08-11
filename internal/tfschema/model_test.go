package tfschema

import (
	"context"
	"encoding/json"
	"testing"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// testModel mirrors the shape the generators produce: an API-assigned id and
// timestamps that are computed only, required and optional fields, a nested
// object mixing configured and computed members, a nested list, and a string
// holding an embedded JSON document.
func testModel(t *testing.T) *Model {
	t.Helper()

	schema := rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id":         rschema.StringAttribute{Computed: true},
			"created_at": rschema.StringAttribute{Computed: true},
			"group_id":   rschema.StringAttribute{Required: true},
			"name":       rschema.StringAttribute{Required: true},
			// Optional and computed: the API fills it in when the configuration
			// does not.
			"description": rschema.StringAttribute{Optional: true, Computed: true},
			// Optional only: whatever the configuration says is final.
			"note":              rschema.StringAttribute{Optional: true},
			"device_config_raw": rschema.StringAttribute{Required: true},
			"config": rschema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]rschema.Attribute{
					"url":    rschema.StringAttribute{Required: true},
					"secret": rschema.StringAttribute{Optional: true, Computed: true},
					"key_id": rschema.StringAttribute{Computed: true},
				},
			},
			"rules": rschema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"name":    rschema.StringAttribute{Required: true},
						"rule_id": rschema.StringAttribute{Computed: true},
					},
				},
			},
			"labels": rschema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		},
	}

	return FromResource(context.Background(), schema, []string{"device_config_raw"}, nil, nil)
}

// object fills in every attribute the caller did not mention with null, the way
// Terraform always hands over a complete object.
func object(t *testing.T, typ tftypes.Type, members map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	asObject, ok := typ.(tftypes.Object)
	require.True(t, ok, "%s is not an object", typ)

	full := make(map[string]tftypes.Value, len(asObject.AttributeTypes))

	for name, attrType := range asObject.AttributeTypes {
		if member, ok := members[name]; ok {
			full[name] = member

			continue
		}

		full[name] = tftypes.NewValue(attrType, nil)
	}

	return tftypes.NewValue(asObject, full)
}

func attrType(t *testing.T, model *Model, name string) tftypes.Type {
	t.Helper()

	asObject, ok := model.Type.(tftypes.Object)
	require.True(t, ok)

	typ, ok := asObject.AttributeTypes[name]
	require.True(t, ok, "no attribute named %q", name)

	return typ
}

func members(t *testing.T, value tftypes.Value) map[string]tftypes.Value {
	t.Helper()

	var out map[string]tftypes.Value
	require.NoError(t, value.As(&out))

	return out
}

func str(value string) tftypes.Value {
	return tftypes.NewValue(tftypes.String, value)
}

func unknown(typ tftypes.Type) tftypes.Value {
	return tftypes.NewValue(typ, tftypes.UnknownValue)
}

func TestBodyCarriesOnlyWhatTheAPIAccepts(t *testing.T) {
	model := testModel(t)

	plan := object(t, model.Type, map[string]tftypes.Value{
		"id":                unknown(tftypes.String),
		"created_at":        unknown(tftypes.String),
		"group_id":          str("group-1"),
		"name":              str("segment"),
		"description":       unknown(tftypes.String),
		"device_config_raw": str(`{"mtu": 1500}`),
		"config": object(t, attrType(t, model, "config"), map[string]tftypes.Value{
			"url":    str("https://example.com"),
			"key_id": unknown(tftypes.String),
		}),
	})

	body, err := model.Body(plan, map[string]bool{"group_id": true})
	require.NoError(t, err)

	require.Equal(t, map[string]any{
		"name": "segment",
		// The embedded document is sent as JSON, not as a string.
		"device_config_raw": map[string]any{"mtu": json.Number("1500")},
		"config":            map[string]any{"url": "https://example.com"},
	}, body)
}

func TestApplyAfterWriteKeepsWhatTheConfigurationAsked(t *testing.T) {
	model := testModel(t)

	plan := object(t, model.Type, map[string]tftypes.Value{
		"id":                unknown(tftypes.String),
		"created_at":        unknown(tftypes.String),
		"group_id":          str("group-1"),
		"name":              str("segment"),
		"description":       str("mine"),
		"device_config_raw": str(`{"mtu": 1500}`),
		"config": object(t, attrType(t, model, "config"), map[string]tftypes.Value{
			"url":    str("https://example.com"),
			"secret": str("shh"),
			"key_id": unknown(tftypes.String),
		}),
	})

	document := map[string]any{
		"id":          "seg-1",
		"created_at":  "2026-08-11T00:00:00Z",
		"name":        "segment",
		"description": "renamed by the server",
		"note":        "added by the server",
		"config":      map[string]any{"url": "https://example.com", "key_id": "key-1"},
	}

	state, err := model.Apply(AfterWrite, document, plan)
	require.NoError(t, err)

	got := members(t, state)

	require.Equal(t, str("seg-1"), got["id"], "an unknown computed attribute takes the API's value")
	require.Equal(t, str("2026-08-11T00:00:00Z"), got["created_at"])
	require.Equal(t, str("mine"), got["description"], "a configured value must survive verbatim")
	require.True(t, got["note"].IsNull(), "an optional attribute left unset stays unset")
	require.Equal(t, str("group-1"), got["group_id"])
	require.Equal(t, str(`{"mtu": 1500}`), got["device_config_raw"], "the configured document is kept as written")

	config := members(t, got["config"])
	require.Equal(t, str("shh"), config["secret"], "the API not echoing a secret does not clear it")
	require.Equal(t, str("key-1"), config["key_id"], "a computed member of a configured object is filled in")
}

func TestApplyAfterWriteLeavesNoUnknownBehind(t *testing.T) {
	model := testModel(t)

	plan := object(t, model.Type, map[string]tftypes.Value{
		"id":                unknown(tftypes.String),
		"created_at":        unknown(tftypes.String),
		"group_id":          str("group-1"),
		"name":              str("segment"),
		"description":       unknown(tftypes.String),
		"labels":            unknown(attrType(t, model, "labels")),
		"rules":             unknown(attrType(t, model, "rules")),
		"device_config_raw": str("{}"),
		"config": object(t, attrType(t, model, "config"), map[string]tftypes.Value{
			"url":    str("https://example.com"),
			"secret": unknown(tftypes.String),
			"key_id": unknown(tftypes.String),
		}),
	})

	state, err := model.Apply(AfterWrite, map[string]any{"id": "seg-1"}, plan)
	require.NoError(t, err)

	// Terraform rejects a state that still holds an unknown after apply.
	require.True(t, state.IsFullyKnown())
}

func TestApplyAfterWriteAlignsNestedListWithTheResponse(t *testing.T) {
	model := testModel(t)
	rulesType := attrType(t, model, "rules")

	elementType, ok := rulesType.(tftypes.List)
	require.True(t, ok)

	rule := func(name string) tftypes.Value {
		return object(t, elementType.ElementType, map[string]tftypes.Value{
			"name":    str(name),
			"rule_id": unknown(tftypes.String),
		})
	}

	plan := object(t, model.Type, map[string]tftypes.Value{
		"id":                unknown(tftypes.String),
		"group_id":          str("group-1"),
		"name":              str("segment"),
		"device_config_raw": str("{}"),
		"config": object(t, attrType(t, model, "config"), map[string]tftypes.Value{
			"url": str("https://example.com"),
		}),
		"rules": tftypes.NewValue(rulesType, []tftypes.Value{rule("allow"), rule("deny")}),
	})

	document := map[string]any{
		"id": "seg-1",
		"rules": []any{
			map[string]any{"name": "allow", "rule_id": "r-1"},
			map[string]any{"name": "deny", "rule_id": "r-2"},
		},
	}

	state, err := model.Apply(AfterWrite, document, plan)
	require.NoError(t, err)

	var rules []tftypes.Value
	require.NoError(t, members(t, state)["rules"].As(&rules))
	require.Len(t, rules, 2)

	first := members(t, rules[0])
	require.Equal(t, str("allow"), first["name"])
	require.Equal(t, str("r-1"), first["rule_id"], "the id the server assigned to each element is kept")
}

func TestApplyAfterReadPrefersTheAPI(t *testing.T) {
	model := testModel(t)

	prior := object(t, model.Type, map[string]tftypes.Value{
		"id":                str("seg-1"),
		"group_id":          str("group-1"),
		"name":              str("segment"),
		"description":       str("stale"),
		"note":              str("mine"),
		"device_config_raw": str(`{"mtu":1500}`),
		"config": object(t, attrType(t, model, "config"), map[string]tftypes.Value{
			"url":    str("https://example.com"),
			"secret": str("shh"),
			"key_id": str("key-1"),
		}),
	})

	document := map[string]any{
		"id":                "seg-1",
		"name":              "renamed",
		"description":       "fresh",
		"device_config_raw": map[string]any{"mtu": 9000},
		"config":            map[string]any{"url": "https://example.com", "key_id": "key-2"},
	}

	state, err := model.Apply(AfterRead, document, prior)
	require.NoError(t, err)

	got := members(t, state)

	require.Equal(t, str("renamed"), got["name"], "drift has to be visible")
	require.Equal(t, str("fresh"), got["description"])
	require.Equal(t, str("mine"), got["note"], "a field the API never returns is not invented away")
	require.Equal(t, str(`{"mtu":9000}`), got["device_config_raw"])

	config := members(t, got["config"])
	require.Equal(t, str("key-2"), config["key_id"])
	require.Equal(t, str("shh"), config["secret"], "a write-only secret survives a refresh")
}

func TestApplyAfterReadKeepsPriorStateForAbsentObject(t *testing.T) {
	model := testModel(t)

	prior := object(t, model.Type, map[string]tftypes.Value{
		"id":       str("seg-1"),
		"group_id": str("group-1"),
		"config": object(t, attrType(t, model, "config"), map[string]tftypes.Value{
			"url": str("https://example.com"),
		}),
	})

	state, err := model.Apply(AfterRead, map[string]any{"id": "seg-1"}, prior)
	require.NoError(t, err)

	config := members(t, members(t, state)["config"])
	require.Equal(t, str("https://example.com"), config["url"])
}

func TestValidateAcceptsTheShapesTheGeneratorsProduce(t *testing.T) {
	require.NoError(t, testModel(t).Validate())
}

func TestValidateRejectsANestedSet(t *testing.T) {
	schema := rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"rules": rschema.SetNestedAttribute{
				Optional: true,
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"name": rschema.StringAttribute{Required: true},
					},
				},
			},
		},
	}

	err := FromResource(context.Background(), schema, nil, nil, nil).Validate()

	// Merging a set by position would silently reorder it, so it is refused
	// rather than mishandled.
	require.ErrorContains(t, err, "unsupported attribute(s): rules")
}

func TestConfigurableNamesAreSortedAndExcludeComputed(t *testing.T) {
	model := testModel(t)

	require.Equal(t, []string{
		"config",
		"description",
		"device_config_raw",
		"group_id",
		"labels",
		"name",
		"note",
		"rules",
	}, model.ConfigurableNames())
}
