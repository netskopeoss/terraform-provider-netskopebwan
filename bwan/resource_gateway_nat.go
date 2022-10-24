package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"
	"github.com/netskopeoss/terraform-provider-netskopebwan/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _resourceGatewayNat) resourceGatewayNatRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var natConfig swagger.InboundNatRule
	var err error

	edgeInput, err := ApplyBinderInputResourceData[resourceGatewayNatInput](rt.InputBinder, d)
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

	d.SetId(utils.Hash(natConfig))
	return diags
}

func (rt _resourceGatewayNat) resourceGatewayNatUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var natConfig swagger.InboundNatRule
	var err error

	edgeInput, err := ApplyBinderInputResourceData[resourceGatewayNatInput](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)
	lock := utils.Mutex.Get(edgeInput.GatewayId)
	lock.Lock()
	defer lock.Unlock()
	gateway, _, err := apiSvc.EdgesApi.GetEdgeById(ctx, edgeInput.GatewayId, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	rt.AddConfig(&gateway, edgeInput)

	addGwInput := swagger.UpdateEdgeInput{
		One2OneNatRules:        gateway.One2OneNatRules,
		PortForwardingNatRules: gateway.PortForwardingNatRules,
	}

	if len(edgeInput.GatewayId) > 0 {
		gateway, _, err := apiSvc.EdgesApi.UpdateEdgeById(ctx, addGwInput, edgeInput.GatewayId, nil)
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
	d.SetId(utils.Hash(natConfig))
	return diags
}

func (rt _resourceGatewayNat) resourceGatewayNatDelete(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	var err error

	edgeInput, err := ApplyBinderInputResourceData[resourceGatewayNatInput](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)
	lock := utils.Mutex.Get(edgeInput.GatewayId)
	lock.Lock()
	defer lock.Unlock()
	gateway, _, err := apiSvc.EdgesApi.GetEdgeById(ctx, edgeInput.GatewayId, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}
	rt.DeleteConfig(&gateway, edgeInput)
	addGwInput := swagger.UpdateEdgeInput{
		One2OneNatRules:        gateway.One2OneNatRules,
		PortForwardingNatRules: gateway.PortForwardingNatRules,
	}

	if len(edgeInput.GatewayId) > 0 {
		_, _, err := apiSvc.EdgesApi.UpdateEdgeById(ctx, addGwInput, edgeInput.GatewayId, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
	} else {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

type _resourceGatewayNat struct {
	Binder       []FieldBinder
	InputBinder  []FieldBinder
	DeleteConfig func(*swagger.Edge, resourceGatewayNatInput)
	GetConfig    func(*swagger.Edge, resourceGatewayNatInput) swagger.InboundNatRule
	AddConfig    func(*swagger.Edge, resourceGatewayNatInput)
}

type resourceGatewayNatInput struct {
	GatewayId string
	swagger.InboundNatRule
}

func resourceGatewayNat() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(resourceGatewayNatInput{}, Cfg{
		"gateway_id":      {Schema: schema.Schema{Required: true}},
		"public_ip":       {Schema: schema.Schema{Required: true}},
		"up_link_if_name": {Schema: schema.Schema{Required: true}},
		"lan_ip":          {Schema: schema.Schema{Required: true}},
		"bi_directional":  {Schema: schema.Schema{Required: true}},
	})

	rt := _resourceGatewayNat{
		Binder:      binder,
		InputBinder: inputBinder,
		DeleteConfig: func(gateway *swagger.Edge, edgeInput resourceGatewayNatInput) {
			index := utils.GetExistingNat(gateway.One2OneNatRules, edgeInput.InboundNatRule)
			if index >= 0 {
				gateway.One2OneNatRules = append(
					gateway.One2OneNatRules[:index],
					gateway.One2OneNatRules[index+1:]...,
				)
			}
		},
		GetConfig: func(gateway *swagger.Edge,
			edgeInput resourceGatewayNatInput) (
			natConfig swagger.InboundNatRule) {
			index := utils.GetExistingNat(gateway.One2OneNatRules, edgeInput.InboundNatRule)
			if index >= 0 {
				natConfig = gateway.One2OneNatRules[index]
			}
			return natConfig
		},
		AddConfig: func(gateway *swagger.Edge, edgeInput resourceGatewayNatInput) {
			index := utils.GetExistingNat(gateway.One2OneNatRules, edgeInput.InboundNatRule)
			if index >= 0 {
				gateway.One2OneNatRules[index] = edgeInput.InboundNatRule
			} else {
				gateway.One2OneNatRules = append(
					gateway.One2OneNatRules,
					edgeInput.InboundNatRule,
				)
			}
		},
	}

	return &schema.Resource{
		CreateContext: rt.resourceGatewayNatUpdate,
		ReadContext:   rt.resourceGatewayNatRead,
		UpdateContext: rt.resourceGatewayNatUpdate,
		DeleteContext: rt.resourceGatewayNatDelete,
		Schema:        swaggerSchema,
	}
}

func resourceGatewayPortForward() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(resourceGatewayNatInput{}, Cfg{
		"gateway_id":      {Schema: schema.Schema{Required: true}},
		"public_ip":       {Schema: schema.Schema{Required: true}},
		"up_link_if_name": {Schema: schema.Schema{Required: true}},
		"lan_ip":          {Schema: schema.Schema{Required: true}},
		"bi_directional":  {Schema: schema.Schema{Required: true}},
		"lan_port":        {Schema: schema.Schema{Required: true}},
		"public_port":     {Schema: schema.Schema{Required: true}},
	})

	rt := _resourceGatewayNat{
		Binder:      binder,
		InputBinder: inputBinder,
		DeleteConfig: func(gateway *swagger.Edge, edgeInput resourceGatewayNatInput) {
			index := utils.GetExistingNat(gateway.PortForwardingNatRules, edgeInput.InboundNatRule)
			if index >= 0 {
				gateway.One2OneNatRules = append(
					gateway.PortForwardingNatRules[:index],
					gateway.PortForwardingNatRules[index+1:]...,
				)
			}
		},
		GetConfig: func(gateway *swagger.Edge,
			edgeInput resourceGatewayNatInput) (
			natConfig swagger.InboundNatRule) {
			index := utils.GetExistingNat(gateway.PortForwardingNatRules, edgeInput.InboundNatRule)
			if index >= 0 {
				natConfig = gateway.PortForwardingNatRules[index]
			}
			return natConfig
		},
		AddConfig: func(gateway *swagger.Edge, edgeInput resourceGatewayNatInput) {
			index := utils.GetExistingNat(gateway.PortForwardingNatRules, edgeInput.InboundNatRule)
			if index >= 0 {
				gateway.PortForwardingNatRules[index] = edgeInput.InboundNatRule
			} else {
				gateway.PortForwardingNatRules = append(
					gateway.PortForwardingNatRules,
					edgeInput.InboundNatRule,
				)
			}
		},
	}

	return &schema.Resource{
		CreateContext: rt.resourceGatewayNatUpdate,
		ReadContext:   rt.resourceGatewayNatRead,
		UpdateContext: rt.resourceGatewayNatUpdate,
		DeleteContext: rt.resourceGatewayNatDelete,
		Schema:        swaggerSchema,
	}
}
