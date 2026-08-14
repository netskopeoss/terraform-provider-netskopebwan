package genresource

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/bwanclient"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/tenanturl"
)

// OperatingTenantAttribute names the escape hatch every resource and data source
// gains when it is switched on: the tenant that one object is managed in,
// whatever tenant the provider itself is configured for.
const OperatingTenantAttribute = "operating_tenant"

// OperatingTenantEnv gates that attribute, and gates it in the schema rather
// than at apply time.
//
// The gate is an environment variable and not a provider argument because the
// point of it is to be invisible. Everything that describes the provider —
// tfplugindocs, the language server's completion and validation, the registry —
// reads the schema the provider serves over GetProviderSchema, so an attribute
// that is not in the schema is in none of them. That also means the variable has
// to be set wherever Terraform runs, not just where the configuration is
// written: with it unset, a configuration that names operating_tenant fails in
// Terraform core as an unexpected argument, which is what an opt-in should do.
//
// State survives the gate being switched off: the framework unmarshals prior
// state with IgnoreUndefinedAttributes, so an attribute the schema no longer has
// is dropped rather than raising an error.
const OperatingTenantEnv = "BWAN_ENABLE_OPERATING_TENANT"

// operatingTenantPattern is what may be spliced into a hostname. It is a guard,
// not a format: the value becomes the leading label of the endpoint's host, so
// anything with a dot or a slash in it could point the provider's bearer token
// at a host the practitioner never configured.
var operatingTenantPattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

const operatingTenantDescription = "Identifier of the tenant to manage this object in, " +
	"instead of the tenant the provider's `endpoint` names. The endpoint's tenant domain is " +
	"replaced with `tid-<operating_tenant>`. Changing it moves the object to another tenant, " +
	"which means replacing it. Undocumented and unsupported: it exists for tooling that " +
	"administers many tenants through one provider configuration, and it is only in the schema " +
	"at all when $" + OperatingTenantEnv + " is set."

// operatingTenantEnabled reports whether the escape hatch is switched on.
func operatingTenantEnabled() bool {
	switch os.Getenv(OperatingTenantEnv) {
	case "", "0", "false":
		return false
	default:
		return true
	}
}

// operatingTenantResourceAttribute is the attribute as a resource carries it.
// Pointing an existing object at another tenant is not an update the API can
// make; the object has to be created there and destroyed here.
func operatingTenantResourceAttribute() rschema.Attribute {
	return rschema.StringAttribute{
		Optional:            true,
		Description:         operatingTenantDescription,
		MarkdownDescription: operatingTenantDescription,
		Validators:          []validator.String{stringvalidator.RegexMatches(operatingTenantPattern, operatingTenantIDError(""))},
		PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
}

func operatingTenantDataSourceAttribute() dschema.Attribute {
	return dschema.StringAttribute{
		Optional:            true,
		Description:         operatingTenantDescription,
		MarkdownDescription: operatingTenantDescription,
		Validators:          []validator.String{stringvalidator.RegexMatches(operatingTenantPattern, operatingTenantIDError(""))},
	}
}

func operatingTenantIDError(id string) string {
	if id == "" {
		return "An operating tenant is a bare identifier: letters, digits and dashes."
	}

	return fmt.Sprintf("%q is not an operating tenant identifier; expected letters, digits and dashes.", id)
}

// operatingTenantOf reads the attribute out of a plan, state or configuration
// value. Everything about it is optional — the gate may be off, so the schema
// may not have the attribute at all — and an absent one means the provider's own
// tenant.
func operatingTenantOf(value tftypes.Value) (string, error) {
	if value.IsNull() || !value.IsKnown() {
		return "", nil
	}

	var members map[string]tftypes.Value
	if err := value.As(&members); err != nil {
		return "", err
	}

	member, ok := members[OperatingTenantAttribute]
	if !ok || member.IsNull() || !member.IsKnown() {
		return "", nil
	}

	var text string
	if err := member.As(&text); err != nil {
		return "", err
	}

	return text, nil
}

// clientFor returns the client addressing the operating tenant named in value,
// which for all but a caller using the escape hatch is the provider's own.
func (m *Meta) clientFor(value tftypes.Value) (bwanclient.API, diag.Diagnostics) {
	var diags diag.Diagnostics

	id, err := operatingTenantOf(value)
	if err != nil {
		diags.AddError("Unexpected value", fmt.Sprintf("Cannot read %s: %s.", OperatingTenantAttribute, err))

		return nil, diags
	}

	if id == "" {
		return m.Client, diags
	}

	if m.Tenant == nil {
		diags.AddError(
			"Operating tenant not enabled",
			fmt.Sprintf("This object sets %s, which needs $%s set wherever Terraform runs.", OperatingTenantAttribute, OperatingTenantEnv),
		)

		return nil, diags
	}

	client, err := m.Tenant(id)
	if err != nil {
		diags.AddError("Could not address the operating tenant", err.Error())

		return nil, diags
	}

	return client, diags
}

// TenantClients builds the factory Meta hands to every resource: one client per
// operating tenant, each the provider's own configuration with the endpoint's
// tenant domain replaced.
//
// It is nil where the escape hatch is switched off, so that a provider without
// it cannot be made to talk to a tenant other than its own — by state left over
// from a run that had it on, or by anything else.
//
// The clients are cached because every operation resolves its own, and a client
// is a connection pool: building one per resource per apply would leave a large
// configuration opening a connection it uses once.
func TenantClients(cfg bwanclient.Config) func(string) (bwanclient.API, error) {
	if !operatingTenantEnabled() {
		return nil
	}

	var (
		mu     sync.Mutex
		cached = map[string]bwanclient.API{}
	)

	return func(id string) (bwanclient.API, error) {
		if !operatingTenantPattern.MatchString(id) {
			return nil, errors.New(operatingTenantIDError(id))
		}

		mu.Lock()
		defer mu.Unlock()

		if client, ok := cached[id]; ok {
			return client, nil
		}

		endpoint, err := tenanturl.ReplaceTenantDomainInURL(cfg.Endpoint, "tid-"+id)
		if err != nil {
			return nil, fmt.Errorf("pointing %s at tenant %s: %w", cfg.Endpoint, id, err)
		}

		scoped := cfg
		scoped.Endpoint = endpoint

		client, err := bwanclient.New(scoped)
		if err != nil {
			return nil, err
		}

		cached[id] = client

		return client, nil
	}
}
