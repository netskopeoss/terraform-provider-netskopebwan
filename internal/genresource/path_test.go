package genresource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestPlaceholders(t *testing.T) {
	require.Equal(t, []string{"group_id", "id"}, placeholders("/address-groups/{group_id}/address-objects/{id}"))
	require.Nil(t, placeholders("/segments"))
}

func TestParentPlaceholdersSkipsTheResourceIDAndDeduplicates(t *testing.T) {
	names := parentPlaceholders(
		"/address-groups/{group_id}/address-objects",
		"/address-groups/{group_id}/address-objects/{id}",
	)

	require.Equal(t, []string{"group_id"}, names)
	require.Empty(t, parentPlaceholders("/segments", "/segments/{id}"))
}

func TestResolvePathEscapesInterpolatedValues(t *testing.T) {
	typ := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"group_id": tftypes.String,
		"id":       tftypes.String,
	}}

	value := tftypes.NewValue(typ, map[string]tftypes.Value{
		"group_id": tftypes.NewValue(tftypes.String, "group/one"),
		"id":       tftypes.NewValue(tftypes.String, "obj 1"),
	})

	resolved, err := resolvePath("/address-groups/{group_id}/address-objects/{id}", value)

	require.NoError(t, err)
	require.Equal(t, "/address-groups/group%2Fone/address-objects/obj%201", resolved)
}

func TestResolvePathReportsMissingValues(t *testing.T) {
	typ := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.String}}

	value := tftypes.NewValue(typ, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, nil),
	})

	_, err := resolvePath("/segments/{id}", value)
	require.ErrorContains(t, err, "id is not set")

	_, err = resolvePath("/segments/{missing}", value)
	require.ErrorContains(t, err, "missing is not an attribute of this resource")
}
