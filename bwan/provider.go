package bwan

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	swagger "github.com/infiotinc/netskopebwan-go-client"
)

// Provider - Netskope APIv2 Provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"baseurl": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NS_SDWAN_MGMT_URL", nil),
			},
			"apitoken": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("NS_SDWAN_MGMT_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"netskopebwan_tenant":               resourceTenant(),
			"netskopebwan_user":                 resourceUser(),
			"netskopebwan_gateway":              resourceGateway(),
			"netskopebwan_gateway_interface":    resourceGatewayInterface(),
			"netskopebwan_gateway_bgpconfig":    resourceGatewayBgp(),
			"netskopebwan_gateway_nat":          resourceGatewayNat(),
			"netskopebwan_gateway_port_forward": resourceGatewayPortForward(),
			"netskopebwan_gateway_staticroute":  resourceGatewayStaticRoute(),
			"netskopebwan_policy":               resourcePolicy(),
			"netskopebwan_gateway_activate":     resourceGatewayActivate(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"netskopebwan_tenant":               dataSourceTenant(),
			"netskopebwan_user":                 dataSourceUser(),
			"netskopebwan_gateway":              dataSourceGateway(),
			"netskopebwan_gateway_interface":    dataSourceGatewayInterface(),
			"netskopebwan_gateway_bgpconfig":    dataSourceGatewayBgp(),
			"netskopebwan_gateway_nat":          dataSourceGatewayNat(),
			"netskopebwan_gateway_port_forward": dataSourceGatewayPortForward(),
			"netskopebwan_gateway_staticroute":  dataSourceGatewayStaticRoute(),
			"netskopebwan_policy":               dataSourcePolicy(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	nsclient := swagger.NewAPIClient(
		&swagger.Configuration{
			BasePath: d.Get("baseurl").(string),
			DefaultHeader: map[string]string{
				"Authorization": "Bearer " + d.Get("apitoken").(string),
			},
		},
	)
	return nsclient, nil
}
