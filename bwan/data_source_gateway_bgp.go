package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _dataSourceGatewayBgp) getExistingBgpPeer(
	bgpPeers []swagger.EdgeBgpConfiguration,
	peer swagger.EdgeBgpConfiguration) (index int) {
	for index, bgp := range bgpPeers {
		if bgp.Neighbor == peer.Neighbor {
			return index
		}
	}
	return -1
}

func (rt _dataSourceGatewayBgp) dataSourceGatewayBgpRead(
	ctx context.Context, d *schema.ResourceData,
	m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	var bgpConfig swagger.EdgeBgpConfiguration
	var err error

	bgpInput, err := ApplyBinderInputResourceData[dataSourceGatewayBgpInput](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)

	if len(bgpInput.GatewayId) > 0 {
		gateway, _, err := apiSvc.EdgesApi.GetEdgeById(ctx, bgpInput.GatewayId, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
		index := rt.getExistingBgpPeer(gateway.BgpConfiguration, bgpInput.EdgeBgpConfiguration)
		if index >= 0 {
			bgpConfig = gateway.BgpConfiguration[index]
		}
	} else {
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, bgpConfig)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(bgpInput.GatewayId + "-" + bgpInput.EdgeBgpConfiguration.Name)
	return diags
}

type _dataSourceGatewayBgp struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

type dataSourceGatewayBgpInput struct {
	GatewayId string
	swagger.EdgeBgpConfiguration
}

func dataSourceGatewayBgp() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(dataSourceGatewayBgpInput{}, Cfg{
		"name":       {Schema: &schema.Schema{Required: true}},
		"gateway_id": {Schema: &schema.Schema{Required: true}},
	})

	rt := _dataSourceGatewayBgp{Binder: binder, InputBinder: inputBinder}

	return &schema.Resource{
		ReadContext: rt.dataSourceGatewayBgpRead,
		Schema:      swaggerSchema,
	}
}
