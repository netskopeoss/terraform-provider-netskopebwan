package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"
	"github.com/netskopeoss/terraform-provider-netskopebwan/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _dataSourceGatewayNat) dataSourceGatewayNatRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	var natConfig swagger.InboundNatRule
	var err error

	edgeInput, err := ApplyBinderInputResourceData[dataSourceGatewayNatInput](
		rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)

	if len(edgeInput.GatewayId) > 0 {
		gateway, _, err := apiSvc.EdgesApi.GetEdgeById(ctx, edgeInput.GatewayId, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}

		natConfig = rt.GetConfig(&gateway, edgeInput)
	} else {
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, natConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(edgeInput.GatewayId + "-" + natConfig.Name)
	return diags
}

type _dataSourceGatewayNat struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
	GetConfig   func(*swagger.Edge, dataSourceGatewayNatInput) swagger.InboundNatRule
}

type dataSourceGatewayNatInput struct {
	GatewayId string
	swagger.InboundNatRule
}

func dataSourceGatewayNat() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(dataSourceGatewayNatInput{}, Cfg{
		"gateway_id": {Schema: &schema.Schema{Required: true}},
	})

	rt := _dataSourceGatewayNat{Binder: binder,
		InputBinder: inputBinder,
		GetConfig: func(gateway *swagger.Edge,
			edgeInput dataSourceGatewayNatInput) (
			natConfig swagger.InboundNatRule) {
			index := utils.GetExistingNat(gateway.One2OneNatRules, edgeInput.InboundNatRule)
			if index >= 0 {
				natConfig = gateway.One2OneNatRules[index]
			}
			return natConfig
		},
	}

	return &schema.Resource{
		ReadContext: rt.dataSourceGatewayNatRead,
		Schema:      swaggerSchema,
	}
}

func dataSourceGatewayPortForward() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(dataSourceGatewayNatInput{}, Cfg{
		"gateway_id": {Schema: &schema.Schema{Required: true}},
		"name":       {Schema: &schema.Schema{Required: true}},
	})

	rt := _dataSourceGatewayNat{Binder: binder,
		InputBinder: inputBinder,
		GetConfig: func(gateway *swagger.Edge,
			edgeInput dataSourceGatewayNatInput) (
			natConfig swagger.InboundNatRule) {
			index := utils.GetExistingNat(gateway.PortForwardingNatRules, edgeInput.InboundNatRule)
			if index >= 0 {
				natConfig = gateway.PortForwardingNatRules[index]
			}
			return natConfig
		},
	}

	return &schema.Resource{
		ReadContext: rt.dataSourceGatewayNatRead,
		Schema:      swaggerSchema,
	}
}
