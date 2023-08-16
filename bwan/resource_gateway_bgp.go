package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"
	"github.com/netskopeoss/terraform-provider-netskopebwan/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _resourceGatewayBgp) fixupBgpConfig(bgp *swagger.EdgeBgpConfiguration) {
	if bgp.LocalAS == 0 {
		bgp.LocalAS = 400
	}
}

func (rt _resourceGatewayBgp) getExistingBgpPeer(bgpPeers []swagger.EdgeBgpConfiguration, peer swagger.EdgeBgpConfiguration) (index int) {
	for index, bgp := range bgpPeers {
		if bgp.Neighbor == peer.Neighbor {
			return index
		}
	}
	return -1
}

func (rt _resourceGatewayBgp) resourceGatewayBgpRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var bgpConfig swagger.EdgeBgpConfiguration
	var err error

	bgpInput, err := ApplyBinderInputResourceData[resourceGatewayBgpInput](rt.InputBinder, d)
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

	d.SetId(utils.Hash(bgpConfig))
	return diags
}

func (rt _resourceGatewayBgp) resourceGatewayBgpUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var bgpConfig swagger.EdgeBgpConfiguration
	var err error

	bgpInput, err := ApplyBinderInputResourceData[resourceGatewayBgpInput](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	lock := utils.Mutex.Get(bgpInput.GatewayId)
	rt.fixupBgpConfig(&bgpInput.EdgeBgpConfiguration)
	lock.Lock()
	defer lock.Unlock()
	apiSvc := m.(*swagger.APIClient)
	gateway, _, err := apiSvc.EdgesApi.GetEdgeById(ctx, bgpInput.GatewayId, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}
	existBgpConfig := gateway.BgpConfiguration
	index := rt.getExistingBgpPeer(existBgpConfig, bgpInput.EdgeBgpConfiguration)
	if index >= 0 {
		existBgpConfig[index] = bgpInput.EdgeBgpConfiguration
	} else {
		existBgpConfig = append(existBgpConfig, bgpInput.EdgeBgpConfiguration)
	}

	addGwInput := swagger.UpdateEdgeInput{
		BgpConfiguration: existBgpConfig,
	}

	if len(bgpInput.GatewayId) > 0 {
		gateway, _, err := apiSvc.EdgesApi.UpdateEdgeById(ctx, addGwInput, bgpInput.GatewayId, nil)
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
	d.SetId(utils.Hash(bgpConfig))
	return diags
}

func (rt _resourceGatewayBgp) resourceGatewayBgpDelete(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	bgpInput, err := ApplyBinderInputResourceData[resourceGatewayBgpInput](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)
	lock := utils.Mutex.Get(bgpInput.GatewayId)
	lock.Lock()
	defer lock.Unlock()
	gateway, _, err := apiSvc.EdgesApi.GetEdgeById(ctx, bgpInput.GatewayId, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}
	existBgpConfig := gateway.BgpConfiguration
	index := rt.getExistingBgpPeer(existBgpConfig, bgpInput.EdgeBgpConfiguration)

	if index >= 0 {
		existBgpConfig = append(existBgpConfig[:index], existBgpConfig[index+1:]...)
		addGwInput := swagger.UpdateEdgeInput{
			BgpConfiguration: existBgpConfig,
		}
		if len(bgpInput.GatewayId) > 0 {
			_, _, err := apiSvc.EdgesApi.UpdateEdgeById(ctx, addGwInput, bgpInput.GatewayId, nil)
			if err != nil {
				if serr, ok := err.(swagger.GenericSwaggerError); ok {
					return diag.FromErr(fmt.Errorf("%s", serr.Body()))
				}
				return diag.FromErr(err)
			}
			d.SetId("")
		} else {
			return diag.FromErr(err)
		}
	}
	return diags
}

type _resourceGatewayBgp struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

type resourceGatewayBgpInput struct {
	GatewayId string
	swagger.EdgeBgpConfiguration
}

func resourceGatewayBgp() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(resourceGatewayBgpInput{}, Cfg{
		"name":       {Schema: &schema.Schema{Required: true}},
		"gateway_id": {Schema: &schema.Schema{Required: true}},
		"neighbor":   {Schema: &schema.Schema{Required: true}},
		"remote_as":  {Schema: &schema.Schema{Required: true}},
	})

	rt := _resourceGatewayBgp{Binder: binder, InputBinder: inputBinder}

	return &schema.Resource{
		CreateContext: rt.resourceGatewayBgpUpdate,
		ReadContext:   rt.resourceGatewayBgpRead,
		UpdateContext: rt.resourceGatewayBgpUpdate,
		DeleteContext: rt.resourceGatewayBgpDelete,
		Schema:        swaggerSchema,
	}
}
