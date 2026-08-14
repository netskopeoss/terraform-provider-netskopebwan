package genresource

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/bwanclient"
)

// Meta is what the provider hands to every resource and data source once it has
// been configured.
type Meta struct {
	Client bwanclient.API
	// TypeName prefixes every resource and data source, so a diagnostic can name
	// the type the way a practitioner writes it.
	TypeName string
	// RawEnabled records which opaque-configuration features the practitioner
	// opted in to.
	RawEnabled map[string]bool
	// Tenant returns a client addressing another operating tenant, and is nil
	// unless the operating_tenant escape hatch is switched on. See
	// OperatingTenantEnv.
	Tenant func(id string) (bwanclient.API, error)
}

// RawOptInArgument names the provider argument guarding a feature.
func RawOptInArgument(feature string) string {
	return "enable_raw_" + feature
}

// RawOptInDescription documents a raw opt-in. The provider argument and the
// error a gated object raises share it, so a practitioner reads the same terms
// wherever they meet the feature.
//
// The wording is deliberately blunt: these objects exist because part of the API
// is not modelled yet, and anyone relying on them has to know it is temporary.
func RawOptInDescription(feature string) string {
	return fmt.Sprintf(
		"Manage %s objects whose configuration the API only exposes as an opaque JSON document. "+
			"Enabling this is an explicit opt-out of compatibility: these resources and data sources are NOT covered by "+
			"the provider's backward-compatibility guarantees, their schemas WILL change without a major release, and they "+
			"WILL be removed once typed replacements ship. Terraform cannot validate or diff a single field of an opaque "+
			"document, so a change to it is all-or-nothing.",
		featureLabel(feature),
	)
}

// rawGateError explains why a gated object refused to do anything, and what to
// set to change that.
func rawGateError(kind, typeName, feature string) (string, string) {
	argument := RawOptInArgument(feature)

	summary := fmt.Sprintf("%s %s is not enabled", kind, typeName)

	detail := fmt.Sprintf(
		"%s carries its configuration as an opaque JSON document, which Terraform cannot validate or diff field by field.\n\n"+
			"To use it, set `%s = true` on the provider:\n\n"+
			"    provider \"%s\" {\n      %s = true\n    }\n\n"+
			"Understand what that opts you in to: %s objects are NOT covered by the provider's backward-compatibility "+
			"guarantees, their schemas WILL change without a major release, and they WILL be removed once a typed "+
			"replacement ships. The typed replacement will take the name without the \"_raw\" suffix.",
		typeName, argument, providerNameOf(typeName), argument, featureLabel(feature),
	)

	return summary, detail
}

// providerNameOf recovers the provider block's name from a resource type. Every
// type is prefixed with it, so taking the prefix keeps the example in the
// diagnostic correct without threading Meta.TypeName through the gate.
func providerNameOf(typeName string) string {
	name, _, found := strings.Cut(typeName, "_")
	if !found {
		return typeName
	}

	return name
}

func featureLabel(feature string) string {
	return strings.ReplaceAll(feature, "_", " ")
}

// gate reports whether a gated object may run, adding the explanation when it
// may not.
func gate(diags *diag.Diagnostics, kind, typeName, feature string, enabled map[string]bool) bool {
	if feature == "" || enabled[feature] {
		return true
	}

	summary, detail := rawGateError(kind, typeName, feature)
	diags.AddError(summary, detail)

	return false
}
