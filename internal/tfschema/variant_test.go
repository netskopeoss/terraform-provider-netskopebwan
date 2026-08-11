package tfschema

import (
	"context"
	"testing"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func block(attributes map[string]rschema.Attribute) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{Optional: true, Attributes: attributes}
}

func requiredString() rschema.StringAttribute {
	return rschema.StringAttribute{Required: true}
}

// monitorModel mirrors a link monitor: an id of its own, and three forms of which
// one is set. Each form carries a single field named after it, which is how the
// API declares them.
func monitorModel(t *testing.T) *Model {
	t.Helper()

	schema := rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id":   rschema.StringAttribute{Computed: true},
			"fqdn": block(map[string]rschema.Attribute{"fqdn": requiredString()}),
			"ipv4": block(map[string]rschema.Attribute{"ipv4": requiredString()}),
			"ipv6": block(map[string]rschema.Attribute{"ipv6": requiredString()}),
		},
	}

	return FromResource(context.Background(), schema, nil, []string{"fqdn", "ipv4", "ipv6"}, nil)
}

// accountModel mirrors a cloud account: the forms sit under a nested object rather
// than at the root, and two of them share a field.
func accountModel(t *testing.T) *Model {
	t.Helper()

	schema := rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id":   rschema.StringAttribute{Computed: true},
			"name": requiredString(),
			"config": rschema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]rschema.Attribute{
					"netskope": block(map[string]rschema.Attribute{"url": requiredString()}),
					"device_security": block(map[string]rschema.Attribute{
						"url":   requiredString(),
						"email": requiredString(),
					}),
				},
			},
		},
	}

	return FromResource(context.Background(), schema, nil,
		[]string{"config.netskope", "config.device_security"}, nil)
}

// value builds the object Terraform would hand over, filling anything the caller
// left out with null.
func value(t *testing.T, model *Model, members map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	object, ok := model.Type.(tftypes.Object)
	require.True(t, ok)

	full := make(map[string]tftypes.Value, len(object.AttributeTypes))

	for name, typ := range object.AttributeTypes {
		if member, ok := members[name]; ok {
			full[name] = member

			continue
		}

		full[name] = tftypes.NewValue(typ, nil)
	}

	return tftypes.NewValue(object, full)
}

func nested(t *testing.T, model *Model, name string, members map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	object, ok := model.Type.(tftypes.Object)
	require.True(t, ok)

	typ, ok := object.AttributeTypes[name].(tftypes.Object)
	require.True(t, ok, "%s is not a nested object", name)

	return tftypes.NewValue(typ, members)
}

func TestVariantGroupsFindsAlternativeBlocksAtEveryDepth(t *testing.T) {
	require.Equal(t,
		[]VariantGroup{{Path: "", Names: []string{"fqdn", "ipv4", "ipv6"}}},
		monitorModel(t).VariantGroups())

	require.Equal(t,
		[]VariantGroup{{Path: "config", Names: []string{"device_security", "netskope"}}},
		accountModel(t).VariantGroups())
}

// TestBodySplicesTheSetFormIntoTheObject is the point of the whole mechanism: the
// schema nests a form, the API declared it flat, and a request has to carry what
// the API declared.
func TestBodySplicesTheSetFormIntoTheObject(t *testing.T) {
	model := monitorModel(t)

	config := value(t, model, map[string]tftypes.Value{
		"fqdn": nested(t, model, "fqdn", map[string]tftypes.Value{
			"fqdn": tftypes.NewValue(tftypes.String, "probe.example.com"),
		}),
	})

	body, err := model.Body(config, nil)

	require.NoError(t, err)
	require.Equal(t, map[string]any{"fqdn": "probe.example.com"}, body,
		"the form's field belongs beside its siblings, not under the form's name")
}

func TestBodySplicesAFormNestedUnderAnObject(t *testing.T) {
	model := accountModel(t)

	object, _ := model.Type.(tftypes.Object)
	configType, _ := object.AttributeTypes["config"].(tftypes.Object)
	netskopeType, _ := configType.AttributeTypes["netskope"].(tftypes.Object)

	config := value(t, model, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "primary"),
		"config": tftypes.NewValue(configType, map[string]tftypes.Value{
			"netskope": tftypes.NewValue(netskopeType, map[string]tftypes.Value{
				"url": tftypes.NewValue(tftypes.String, "https://tenant.example.com"),
			}),
			"device_security": tftypes.NewValue(configType.AttributeTypes["device_security"], nil),
		}),
	})

	body, err := model.Body(config, nil)

	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"name":   "primary",
		"config": map[string]any{"url": "https://tenant.example.com"},
	}, body)
}

func TestBodyRefusesTwoFormsAtOnce(t *testing.T) {
	model := monitorModel(t)

	config := value(t, model, map[string]tftypes.Value{
		"fqdn": nested(t, model, "fqdn", map[string]tftypes.Value{
			"fqdn": tftypes.NewValue(tftypes.String, "probe.example.com"),
		}),
		"ipv4": nested(t, model, "ipv4", map[string]tftypes.Value{
			"ipv4": tftypes.NewValue(tftypes.String, "10.0.0.1"),
		}),
	})

	_, err := model.Body(config, nil)

	require.ErrorContains(t, err, "fqdn and ipv4 are alternatives")
}

func TestReshapeNestsTheFormTheAPISent(t *testing.T) {
	reshaped := monitorModel(t).Reshape(map[string]any{
		"id":   "monitor-1",
		"fqdn": "probe.example.com",
	})

	require.Equal(t, map[string]any{
		"id":   "monitor-1",
		"fqdn": map[string]any{"fqdn": "probe.example.com"},
	}, reshaped)
}

// TestReshapePrefersTheFormItMatchesWholly covers two forms sharing a field: the
// shorter one must not swallow an object that really belongs to the longer one, and
// the longer one must not claim an object carrying only the shared field.
func TestReshapePrefersTheFormItMatchesWholly(t *testing.T) {
	model := accountModel(t)

	onlyURL := model.Reshape(map[string]any{
		"config": map[string]any{"url": "https://tenant.example.com"},
	})
	require.Equal(t, map[string]any{
		"config": map[string]any{"netskope": map[string]any{"url": "https://tenant.example.com"}},
	}, onlyURL)

	withEmail := model.Reshape(map[string]any{
		"config": map[string]any{"url": "https://tenant.example.com", "email": "ops@example.com"},
	})
	require.Equal(t, map[string]any{
		"config": map[string]any{"device_security": map[string]any{
			"url":   "https://tenant.example.com",
			"email": "ops@example.com",
		}},
	}, withEmail)
}

// TestReshapeLeavesAnUnplaceableObjectAlone matters because inventing a form would
// put values into state the API never sent back.
func TestReshapeLeavesAnUnplaceableObjectAlone(t *testing.T) {
	reshaped := monitorModel(t).Reshape(map[string]any{"id": "monitor-1"})

	require.Equal(t, map[string]any{"id": "monitor-1"}, reshaped)
}

// TestApplyRoundTripsAForm walks the whole path: a request carries the form flat,
// the response comes back flat, and state holds it nested again.
func TestApplyRoundTripsAForm(t *testing.T) {
	model := monitorModel(t)

	config := value(t, model, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"ipv4": nested(t, model, "ipv4", map[string]tftypes.Value{
			"ipv4": tftypes.NewValue(tftypes.String, "10.0.0.1"),
		}),
	})

	body, err := model.Body(config, nil)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"ipv4": "10.0.0.1"}, body)

	state, err := model.Apply(AfterRead, map[string]any{"id": "monitor-1", "ipv4": "10.0.0.1"}, config)
	require.NoError(t, err)

	var members map[string]tftypes.Value
	require.NoError(t, state.As(&members))

	var id string
	require.NoError(t, members["id"].As(&id))
	require.Equal(t, "monitor-1", id)

	var form map[string]tftypes.Value
	require.NoError(t, members["ipv4"].As(&form))

	var address string
	require.NoError(t, form["ipv4"].As(&address))
	require.Equal(t, "10.0.0.1", address)

	require.True(t, members["fqdn"].IsNull(), "the forms that did not arrive stay unset")
	require.True(t, members["ipv6"].IsNull())
}
