package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"
	"github.com/netskopeoss/terraform-provider-netskopebwan/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _resourceGateway) resourceGatewayCreate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiSvc := m.(*swagger.APIClient)
	gwInput, err := ApplyBinderInputResourceData[swagger.Edge](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	addGwInput := swagger.AddEdgeInput{
		Name:           gwInput.Name,
		Role:           gwInput.Role,
		Model:          gwInput.Model,
		AssignedPolicy: gwInput.AssignedPolicy,
		Description:    gwInput.Description,
		Serialnumber:   gwInput.Serialnumber,
	}

	gateway, _, err := apiSvc.EdgesApi.AddEdge(ctx, addGwInput, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, gateway)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(gateway.Id)
	return diags

}

func (rt _resourceGateway) resourceGatewayRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	apiSvc := m.(*swagger.APIClient)
	gwInput, err := ApplyBinderInputResourceData[swagger.Edge](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}
	gateway, _, err := apiSvc.EdgesApi.GetEdgeById(ctx, gwInput.Id, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}
	err = ApplyBinderResourceData(rt.Binder, d, gateway)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(gateway.Id)
	return diags
}

func (rt _resourceGateway) resourceGatewayUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	if d.Get("id").(string) == "" {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)
	gwInput, err := ApplyBinderInputResourceData[swagger.Edge](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	addGwInput := swagger.UpdateEdgeInput{
		Name:                   gwInput.Name,
		Role:                   gwInput.Role,
		AssignedPolicy:         gwInput.AssignedPolicy,
		Description:            gwInput.Description,
		Serialnumber:           gwInput.Serialnumber,
		Swversion:              gwInput.Swversion,
		Swmanifest:             gwInput.Swmanifest,
		Psk:                    gwInput.Psk,
		BgpConfiguration:       gwInput.BgpConfiguration,
		StaticRoutes:           gwInput.StaticRoutes,
		One2OneNatRules:        gwInput.One2OneNatRules,
		PortForwardingNatRules: gwInput.PortForwardingNatRules,
		Interfaces:             &gwInput.Interfaces,
	}

	lock := utils.Mutex.Get(gwInput.Id)
	lock.Lock()
	defer lock.Unlock()
	gateway, _, err := apiSvc.EdgesApi.UpdateEdgeById(ctx, addGwInput, gwInput.Id, nil)

	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, gateway)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(gateway.Id)
	return diags

}

func (rt _resourceGateway) resourceGatewayDelete(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiSvc := m.(*swagger.APIClient)
	gwInput, err := ApplyBinderInputResourceData[swagger.Edge](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, err = apiSvc.EdgesApi.DeleteEdgeById(ctx, gwInput.Id, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

type _resourceGateway struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

func resourceGateway() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(swagger.Edge{}, Cfg{
		"name": {Schema: schema.Schema{Required: true}},
	})

	rt := _resourceGateway{Binder: binder, InputBinder: inputBinder}

	return &schema.Resource{
		CreateContext: rt.resourceGatewayCreate,
		ReadContext:   rt.resourceGatewayRead,
		UpdateContext: rt.resourceGatewayUpdate,
		DeleteContext: rt.resourceGatewayDelete,
		Schema:        swaggerSchema,
	}
}
