// Package provider wires the generated resources and data sources into a
// terraform-plugin-framework provider.
package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"infiot.com/infiot/mgmt/tf-provider/internal/bwanclient"
	"infiot.com/infiot/mgmt/tf-provider/internal/genresource"
	"infiot.com/infiot/mgmt/tf-provider/internal/registry"
)

// Environment variables standing in for the provider arguments of the same name,
// so a token never has to live in a configuration file.
const (
	endpointEnv = "BWAN_ENDPOINT"
	tokenEnv    = "BWAN_TOKEN"
)

// TypeName prefixes every resource and data source this provider exposes.
const TypeName = "bwan"

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
		Description: "Manages Netskope Borderless WAN through its v2 API.",
		Attributes:  attributes,
	}
}

func (p *bwanProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
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
		UserAgent:          "terraform-provider-bwan/" + p.version,
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
