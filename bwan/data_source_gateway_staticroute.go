package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _dataSourceGatewayStaticRoute) getExistingStaticRoute(
	routes []swagger.StaticRoute, input swagger.StaticRoute) (index int) {
	for index, route := range routes {
		if route.Destination == input.Destination {
			return index
		}
	}
	return -1
}

func (rt _dataSourceGatewayStaticRoute) dataSourceGatewayStaticRouteRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	var routeConfig swagger.StaticRoute
	var err error

	edgeInput, err := ApplyBinderInputResourceData[dataSourceGatewayStaticRouteInput](rt.InputBinder, d)
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
		index := rt.getExistingStaticRoute(gateway.StaticRoutes, edgeInput.StaticRoute)
		if index >= 0 {
			routeConfig = gateway.StaticRoutes[index]
		}
	} else {
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, routeConfig)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(edgeInput.GatewayId)
	return diags
}

type _dataSourceGatewayStaticRoute struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

type dataSourceGatewayStaticRouteInput struct {
	GatewayId string
	swagger.StaticRoute
}

func dataSourceGatewayStaticRoute() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(
		dataSourceGatewayStaticRouteInput{}, Cfg{
			"gateway_id":  {Schema: schema.Schema{Required: true}},
			"destination": {Schema: schema.Schema{Required: true}},
		})

	rt := _dataSourceGatewayStaticRoute{Binder: binder, InputBinder: inputBinder}

	return &schema.Resource{
		ReadContext: rt.dataSourceGatewayStaticRouteRead,
		Schema:      swaggerSchema,
	}
}
