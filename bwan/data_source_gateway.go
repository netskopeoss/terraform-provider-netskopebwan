package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _dataSourceGateway) dataSourceGatewayRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	var gateway swagger.Edge
	var err error

	gwInput, err := ApplyBinderInputResourceData[swagger.Edge](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)

	if len(gwInput.Id) > 0 {
		gateway, _, err = apiSvc.EdgesApi.GetEdgeById(ctx, gwInput.Id, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
	} else if len(gwInput.Name) > 0 {
		gatewayList, _, err := apiSvc.EdgesApi.GetAllEdges(ctx, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
		for _, gw := range gatewayList.Data {
			if gw.Name == gwInput.Name {
				gateway = gw
				break
			}
		}
		if len(gateway.Name) == 0 {
			return diag.FromErr(err)
		}
	}
	err = ApplyBinderResourceData(rt.Binder, d, gateway)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(gateway.Id)
	return diags

}

type _dataSourceGateway struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

func dataSourceGateway() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(swagger.Edge{}, Cfg{})

	rt := _dataSourceGateway{Binder: binder, InputBinder: inputBinder}
	return &schema.Resource{
		ReadContext: rt.dataSourceGatewayRead,
		Schema:      swaggerSchema,
	}
}
