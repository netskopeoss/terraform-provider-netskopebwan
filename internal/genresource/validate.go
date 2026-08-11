package genresource

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"infiot.com/infiot/mgmt/tf-provider/internal/tfschema"
)

// validateVariants reports a set of alternative blocks that does not have exactly
// one member set.
//
// The API declares those forms as alternative shapes of one object, which has no
// Terraform type, so each form is a block and the rule that one of them applies
// cannot be expressed in the schema. Checking it here means a practitioner hears
// about it from `terraform validate` rather than from a rejected request.
func validateVariants(model *tfschema.Model, config tftypes.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if config.IsNull() || !config.IsKnown() {
		return diags
	}

	for _, group := range model.VariantGroups() {
		object, ok := objectAt(config, group.Path)
		if !ok {
			continue
		}

		var set []string

		for _, name := range group.Names {
			if member, ok := object[name]; ok && member.IsKnown() && !member.IsNull() {
				set = append(set, name)
			}
		}

		if len(set) == 1 {
			continue
		}

		diags.Append(variantDiagnostic(group, set))
	}

	return diags
}

func variantDiagnostic(group tfschema.VariantGroup, set []string) diag.Diagnostic {
	attribute := fwpath.Root(group.Names[0])
	if group.Path != "" {
		attribute = fwpath.Root(strings.SplitN(group.Path, ".", 2)[0])
	}

	quoted := make([]string, 0, len(group.Names))

	for _, name := range group.Names {
		quoted = append(quoted, group.Attribute(name))
	}

	options := strings.Join(quoted, ", ")

	if len(set) == 0 {
		return diag.NewAttributeErrorDiagnostic(
			attribute,
			"Missing one of "+options,
			fmt.Sprintf(
				"The API takes one of these forms, so exactly one of %s has to be set. None of them is.",
				options),
		)
	}

	return diag.NewAttributeErrorDiagnostic(
		attribute,
		"Too many of "+options,
		fmt.Sprintf(
			"The API takes one of these forms, so exactly one of %s has to be set. These are set: %s.",
			options, strings.Join(set, ", ")),
	)
}

// objectAt walks to the object holding a group. A path that runs through something
// unset has no group to check, which is reported by the second result.
func objectAt(value tftypes.Value, path string) (map[string]tftypes.Value, bool) {
	current := value

	if path != "" {
		for segment := range strings.SplitSeq(path, ".") {
			members, ok := objectMembers(current)
			if !ok {
				return nil, false
			}

			next, ok := members[segment]
			if !ok || next.IsNull() || !next.IsKnown() {
				return nil, false
			}

			current = next
		}
	}

	return objectMembers(current)
}

func objectMembers(value tftypes.Value) (map[string]tftypes.Value, bool) {
	if value.IsNull() || !value.IsKnown() {
		return nil, false
	}

	if _, ok := value.Type().(tftypes.Object); !ok {
		return nil, false
	}

	var members map[string]tftypes.Value
	if err := value.As(&members); err != nil {
		return nil, false
	}

	return members, true
}
