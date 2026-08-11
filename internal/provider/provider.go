// Package provider wires the generated resources and data sources into a
// terraform-plugin-framework provider.
package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/bwanclient"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/genresource"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/registry"
)

// Environment variables standing in for the provider arguments of the same name,
// so a token never has to live in a configuration file.
const (
	endpointEnv = "BWAN_ENDPOINT"
	tokenEnv    = "BWAN_TOKEN"
)

// TypeName prefixes every resource and data source this provider exposes. It
// has to match the name the registry derives from the repository
// (terraform-provider-netskopebwan), because that is what decides which
// resource a docs/ page belongs to.
const TypeName = "netskopebwan"

// providerSummary, AlphaHeadline and AlphaDetail are what the provider says
// about itself. They travel through the provider schema into docs/index.md,
// which is the page the Terraform registry shows first, so a practitioner reads
// how finished this provider is before installing it rather than after.
//
// The README says the same thing for anyone who arrives at the repository
// instead, and a test holds the two together.
const (
	providerSummary = "Manages Netskope Borderless WAN through its v2 API."

	// AlphaHeadline is the warning itself, short enough to be read, and says what
	// to do rather than only what to fear. The registry treats a prerelease as
	// the latest version even though Terraform will not install one, so this page
	// is what a practitioner sees first and it has to send them to the 0.x line.
	AlphaHeadline = "The 1.x line is an alpha; use the released 0.x line for now."

	// AlphaDetail is why the 1.x schema cannot be promised — it is generated from
	// a specification that is still moving — and how to ask for either line.
	AlphaDetail = "1.x resources, data sources and attributes are generated from the BWAN v2 " +
		"OpenAPI specification, which is itself still changing, so any of them may be renamed, " +
		"reshaped or removed from one prerelease to the next, and a configuration or state " +
		"written against one may need reworking for the next. Until 1.0.0 is released, ask for " +
		`version = "~> 0.0". Terraform installs a prerelease only for an exact version, and a ` +
		"prerelease then refuses to configure until that same version is acknowledged with the " +
		PreReleaseArgument + " argument, so nobody runs the 1.x line by accident. Setting it " +
		"accepts that instability: a prerelease is NOT covered by the provider's " +
		"backward-compatibility guarantees."
)

// PreReleaseArgument is the provider argument a prerelease build has to be
// acknowledged through before it will configure.
//
// It takes the version rather than a boolean, so the acknowledgement is of one
// build and not of prereleases in general: the next one may have moved the
// schema again, and that is worth being asked about twice.
const PreReleaseArgument = "enable_pre_release"

// PreReleaseDescription tells a practitioner reading the registry's home page
// what the gate is for before they meet it as an error.
//
// It names no version, because this text is generated into docs/ before the tag
// it will be published under exists: an example version here would be stale by
// the time anyone read it. The version a build actually wants is in the error
// the build itself raises.
const PreReleaseDescription = "Acknowledge running a prerelease build of this provider by naming " +
	"the exact version installed, which is the version Terraform records in .terraform.lock.hcl. " +
	"A prerelease refuses to configure without it. Setting it is an explicit acceptance of the " +
	"instability that comes with one: a prerelease is NOT covered by the provider's " +
	"backward-compatibility guarantees, its resources, data sources and attributes MAY be " +
	"renamed, reshaped or removed in the next prerelease, and a configuration or state written " +
	"against this version MAY need reworking to move to it. The version is named rather than " +
	"switched on so that the next prerelease asks again. This has no effect on a released version."

// New returns the provider factory the plugin server is started with.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &bwanProvider{version: version}
	}
}

var _ provider.Provider = (*bwanProvider)(nil)

type bwanProvider struct {
	version string
}

func (p *bwanProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = TypeName
	resp.Version = p.version
}

func (p *bwanProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"endpoint": schema.StringAttribute{
			Optional: true,
			Description: "Base URL of the BWAN API, without the version prefix, " +
				"e.g. https://foo.api.infiot.net. Defaults to $" + endpointEnv + ".",
		},
		"token": schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "Bearer token used to authenticate. Defaults to $" + tokenEnv + ".",
		},
		"request_timeout": schema.StringAttribute{
			Optional:    true,
			Description: `How long to wait for a single API call, as a Go duration such as "30s". Defaults to "60s".`,
		},
		"insecure_skip_verify": schema.BoolAttribute{
			Optional:    true,
			Description: "Skip TLS certificate verification. Only useful against a tenant serving a self-signed certificate.",
		},
		PreReleaseArgument: schema.StringAttribute{
			Optional:    true,
			Description: PreReleaseDescription,
		},
	}

	// Objects the API only exposes through an opaque configuration document are
	// off by default and have to be asked for by name. See RawOptInDescription
	// for what enabling one means.
	for _, feature := range registry.RawFeatures() {
		attributes[genresource.RawOptInArgument(feature)] = schema.BoolAttribute{
			Optional:            true,
			Description:         genresource.RawOptInDescription(feature),
			MarkdownDescription: genresource.RawOptInDescription(feature),
		}
	}

	resp.Schema = schema.Schema{
		Description:         providerSummary + " " + AlphaHeadline + " " + AlphaDetail,
		MarkdownDescription: providerSummary + "\n\n~> **" + AlphaHeadline + "** " + AlphaDetail,
		Attributes:          attributes,
	}
}

func (p *bwanProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Nothing else is worth reporting to someone running a prerelease they have
	// not asked for: the answer is to pin a released version, not to fix the
	// arguments.
	if !p.acknowledgedPreRelease(ctx, req, resp) {
		return
	}

	endpoint := p.stringArgument(ctx, req, resp, "endpoint", endpointEnv)
	token := p.stringArgument(ctx, req, resp, "token", tokenEnv)
	timeoutText := p.stringArgument(ctx, req, resp, "request_timeout", "")

	var insecure types.Bool

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("insecure_skip_verify"), &insecure)...)

	rawEnabled := map[string]bool{}

	for _, feature := range registry.RawFeatures() {
		var enabled types.Bool

		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(genresource.RawOptInArgument(feature)), &enabled)...)
		rawEnabled[feature] = enabled.ValueBool()
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing endpoint",
			"Set the provider's `endpoint` argument or the $"+endpointEnv+" environment variable.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing token",
			"Set the provider's `token` argument or the $"+tokenEnv+" environment variable.",
		)
	}

	var timeout time.Duration

	if timeoutText != "" {
		parsed, err := time.ParseDuration(timeoutText)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid request_timeout",
				fmt.Sprintf("%q is not a Go duration: %s.", timeoutText, err),
			)
		}

		timeout = parsed
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := bwanclient.New(bwanclient.Config{
		Endpoint:           endpoint,
		Token:              token,
		Timeout:            timeout,
		InsecureSkipVerify: insecure.ValueBool(),
		UserAgent:          "terraform-provider-netskopebwan/" + p.version,
	})
	if err != nil {
		resp.Diagnostics.AddError("Could not configure the BWAN client", err.Error())

		return
	}

	meta := &genresource.Meta{
		Client:     client,
		TypeName:   TypeName,
		RawEnabled: rawEnabled,
	}

	resp.ResourceData = meta
	resp.DataSourceData = meta
}

func (p *bwanProvider) Resources(_ context.Context) []func() resource.Resource {
	definitions := registry.Resources()
	out := make([]func() resource.Resource, 0, len(definitions))

	for _, definition := range definitions {
		out = append(out, genresource.NewResource(definition))
	}

	return out
}

func (p *bwanProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	definitions := registry.DataSources()
	out := make([]func() datasource.DataSource, 0, len(definitions))

	for _, definition := range definitions {
		out = append(out, genresource.NewDataSource(definition))
	}

	return out
}

// acknowledgedPreRelease reports whether configuration may go on, refusing a
// prerelease build the configuration has not named.
//
// The version is only known here, at runtime, so this is where a practitioner is
// told which one to acknowledge: the argument's own documentation is generated
// before any tag exists and cannot name it.
func (p *bwanProvider) acknowledgedPreRelease(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) bool {
	var argument types.String

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(PreReleaseArgument), &argument)...)

	if resp.Diagnostics.HasError() {
		return false
	}

	acknowledged := strings.TrimPrefix(strings.TrimSpace(argument.ValueString()), "v")
	release, prerelease, stamped := splitVersion(p.version)

	if !stamped || prerelease == "" {
		// A released build has nothing to acknowledge. Saying so beats leaving a
		// line in the configuration that looks like it is still doing something,
		// but an unstamped development build is not a release either and has no
		// business complaining about it.
		if acknowledged != "" && stamped {
			resp.Diagnostics.AddWarning(
				"Acknowledging a prerelease that is not being run",
				fmt.Sprintf(
					"This provider is %s, a released version, so %s does nothing here and can be removed.",
					release, PreReleaseArgument),
			)
		}

		return true
	}

	switch acknowledged {
	case p.version:
		return true
	case "":
		resp.Diagnostics.AddError(
			"This provider is a prerelease",
			fmt.Sprintf(
				"Version %s is a prerelease: its resources, data sources and attributes are "+
					"generated from a specification that is still changing, so a configuration or "+
					"state written against it may need reworking for the next prerelease.\n\n"+
					"Use a released version instead — %s — or acknowledge this one:\n\n"+
					"  provider %q {\n    %s = %q\n  }\n\n"+
					"Setting that accepts the instability: this version is not covered by the "+
					"provider's backward-compatibility guarantees, and the next prerelease may "+
					"rename, reshape or remove any part of the schema it exposes. The "+
					"acknowledgement names one version on purpose, so that next prerelease asks "+
					"again rather than inheriting this answer.",
				p.version, `version = "~> 0.0"`, TypeName, PreReleaseArgument, p.version),
		)
	default:
		resp.Diagnostics.AddError(
			"A different prerelease is acknowledged",
			fmt.Sprintf(
				"%s acknowledges %q, but the provider being run is %s. Terraform records which "+
					"version it installed in .terraform.lock.hcl.\n\n"+
					"Acknowledge the version being run, or pin the one meant with version = %q "+
					"and run terraform init -upgrade.",
				PreReleaseArgument, acknowledged, p.version, "= "+acknowledged),
		)
	}

	return false
}

// splitVersion breaks a stamped version into its release and prerelease halves,
// reporting whether it was stamped at all: a build from a working tree carries
// "dev", which is not a version and makes no promises either way.
func splitVersion(version string) (release, prerelease string, stamped bool) {
	release, prerelease, _ = strings.Cut(strings.TrimPrefix(version, "v"), "-")

	if release == "" || release[0] < '0' || release[0] > '9' {
		return "", "", false
	}

	return release, prerelease, true
}

// stringArgument reads a string argument, falling back to an environment
// variable so credentials can stay out of the configuration.
func (p *bwanProvider) stringArgument(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
	name string,
	env string,
) string {
	var value types.String

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(name), &value)...)

	if !value.IsNull() && !value.IsUnknown() && value.ValueString() != "" {
		return value.ValueString()
	}

	if env == "" {
		return ""
	}

	return os.Getenv(env)
}
