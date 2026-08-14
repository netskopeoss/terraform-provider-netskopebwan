package genresource

import (
	"errors"
	"fmt"
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
// carries: the tenant that one object is managed in, whatever tenant the
// provider itself is configured for.
//
// It is undocumented, and it is undocumented by omission from the pages rather
// than by omission from the schema. Terraform core validates a configuration
// against the schema the provider serves, so an attribute that is not in the
// schema is not an argument that can be written at all — being usable and being
// in the schema are the same thing. tools/tfdocs is what generates docs/ and
// examples/, and it leaves this attribute out of both; see `documented` there.
//
// What that does not hide it from is an editor. The language server reads the
// same schema Terraform does, so it completes and validates operating_tenant
// like any other argument. Nothing in the plugin protocol marks an attribute as
// internal — SchemaAttribute has Sensitive, Deprecated and WriteOnly, and no
// third state between present and absent.
const OperatingTenantAttribute = "operating_tenant"

// operatingTenantPattern is what may be spliced into a hostname. It is a guard,
// not a format: the value becomes the leading label of the endpoint's host, so
// anything with a dot or a slash in it could point the provider's bearer token
// at a host the practitioner never configured.
var operatingTenantPattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

const operatingTenantDescription = "Identifier of the tenant to manage this object in, " +
	"instead of the tenant the provider's `endpoint` names. The endpoint's tenant domain is " +
	"replaced with `tid-<operating_tenant>`. Changing it moves the object to another tenant, " +
	"which means replacing it. Undocumented and unsupported: it exists for tooling that " +
	"administers many tenants through one provider configuration, and it appears in no " +
	"published documentation."

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
// value. An absent one means the provider's own tenant, which is what almost
// every object says.
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
			"Provider not configured",
			fmt.Sprintf("This object sets %s, but the provider has no way to reach another tenant. This is a bug in the provider.", OperatingTenantAttribute),
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
// The clients are cached because every operation resolves its own, and a client
// is a connection pool: building one per resource per apply would leave a large
// configuration opening a connection it uses once.
func TenantClients(cfg bwanclient.Config) func(string) (bwanclient.API, error) {
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
