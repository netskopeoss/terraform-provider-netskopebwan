package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"
	"github.com/netskopeoss/terraform-provider-netskopebwan/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _resourceGatewayStaticRoute) fixupStaticRouteConfig(
	route *swagger.StaticRoute) {
	if route.Cost == 0 {
		route.Cost = 1
	}
}

func (rt _resourceGatewayStaticRoute) getExistingStaticRoute(
	bgpPeers []swagger.StaticRoute, peer swagger.StaticRoute) (index int) {
	for index, bgp := range bgpPeers {
		if bgp.Destination == peer.Destination {
			return index
		}
	}
	return -1
}

func (rt _resourceGatewayStaticRoute) resourceGatewayStaticRouteRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var routeConfig swagger.StaticRoute
	var err error

	edgeInput, err := ApplyBinderInputResourceData[resourceGatewayStaticRouteInput](rt.InputBinder, d)
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

	d.SetId(utils.Hash(routeConfig))
	return diags
}

func (rt _resourceGatewayStaticRoute) resourceGatewayStaticRouteUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var routeConfig swagger.StaticRoute
	var err error

	edgeInput, err := ApplyBinderInputResourceData[resourceGatewayStaticRouteInput](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}
	rt.fixupStaticRouteConfig(&edgeInput.StaticRoute)
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
	existRoutes := gateway.StaticRoutes
	index := rt.getExistingStaticRoute(existRoutes, edgeInput.StaticRoute)
	if index >= 0 {
		existRoutes[index] = edgeInput.StaticRoute
	} else {
		existRoutes = append(existRoutes, edgeInput.StaticRoute)
	}
	addGwInput := swagger.UpdateEdgeInput{
		StaticRoutes: existRoutes,
	}

	if len(edgeInput.GatewayId) > 0 {
		gateway, _, err := apiSvc.EdgesApi.UpdateEdgeById(ctx, addGwInput, edgeInput.GatewayId, nil)
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
	d.SetId(utils.Hash(routeConfig))
	return diags
}

func (rt _resourceGatewayStaticRoute) resourceGatewayStaticRouteDelete(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var routeConfig swagger.StaticRoute
	var err error

	edgeInput, err := ApplyBinderInputResourceData[resourceGatewayStaticRouteInput](rt.InputBinder, d)
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
	existRoutes := gateway.StaticRoutes
	index := rt.getExistingStaticRoute(existRoutes, edgeInput.StaticRoute)
	if index >= 0 {
		existRoutes = append(existRoutes[:index], existRoutes[index+1:]...)

		addGwInput := swagger.UpdateEdgeInput{
			StaticRoutes: existRoutes,
		}

		if len(edgeInput.GatewayId) > 0 {
			gateway, _, err := apiSvc.EdgesApi.UpdateEdgeById(ctx, addGwInput, edgeInput.GatewayId, nil)
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
	}

	err = ApplyBinderResourceData(rt.Binder, d, routeConfig)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

type _resourceGatewayStaticRoute struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

type resourceGatewayStaticRouteInput struct {
	GatewayId string
	swagger.StaticRoute
}

func resourceGatewayStaticRoute() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(resourceGatewayStaticRouteInput{}, Cfg{
		"gateway_id":  {Schema: schema.Schema{Required: true}},
		"destination": {Schema: schema.Schema{Required: true}},
		"device":      {Schema: schema.Schema{Required: true}},
		"nhop":        {Schema: schema.Schema{Required: true}},
	})

	rt := _resourceGatewayStaticRoute{Binder: binder, InputBinder: inputBinder}

	return &schema.Resource{
		CreateContext: rt.resourceGatewayStaticRouteUpdate,
		ReadContext:   rt.resourceGatewayStaticRouteRead,
		UpdateContext: rt.resourceGatewayStaticRouteUpdate,
		DeleteContext: rt.resourceGatewayStaticRouteDelete,
		Schema:        swaggerSchema,
	}
}
