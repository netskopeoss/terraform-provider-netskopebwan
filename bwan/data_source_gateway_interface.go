package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _dataSourceGatewayInterface) dataSourceGatewayInterfaceRead(
	ctx context.Context, d *schema.ResourceData,
	m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	var intf swagger.InterfaceSettings
	var err error

	intfInput, err := ApplyBinderInputResourceData[dataSourceGatewayInterfaceInput](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)

	if len(intfInput.GatewayId) > 0 && len(intfInput.Name) > 0 {
		intf, _, err = apiSvc.EdgesApi.GetEdgeIfByName(ctx, intfInput.GatewayId, intfInput.Name, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
	} else {
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, intf)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(intf.Name)
	return diags

}

type _dataSourceGatewayInterface struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

type dataSourceGatewayInterfaceInput struct {
	GatewayId string
	swagger.InterfaceSettings
}

func dataSourceGatewayInterface() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(dataSourceGatewayInterfaceInput{}, Cfg{
		"name":       {Schema: schema.Schema{Required: true}},
		"gateway_id": {Schema: schema.Schema{Required: true}},
	})

	rt := _dataSourceGatewayInterface{Binder: binder, InputBinder: inputBinder}

	return &schema.Resource{
		ReadContext: rt.dataSourceGatewayInterfaceRead,
		Schema:      swaggerSchema,
	}
}
