package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"
	"github.com/netskopeoss/terraform-provider-netskopebwan/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _resourceGatewayInterface) fixupInterfaceConfig(
	intf *swagger.InterfaceSettings) {
	if len(intf.Addresses) == 0 {
		intf.Addresses = []swagger.InterfaceSettingsAddresses{
			{
				AddressAssignment: "dhcp",
				AddressFamily:     "ipv4",
				DnsPrimary:        "8.8.8.8",
				DnsSecondary:      "8.8.4.4",
			},
		}
	}
	if intf.Type_ == "" {
		intf.Type_ = "ethernet"
	}
	if intf.Mode == "" {
		intf.Mode = "routed"
	}
	if intf.Zone == "" {
		intf.Zone = "trusted"
	}
	if intf.Mtu == 0 {
		intf.Mtu = 1500
	}
}

func (rt _resourceGatewayInterface) resourceGatewayInterfaceRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var intf swagger.InterfaceSettings

	intfInput, err := ApplyBinderInputResourceData[resourceGatewayInterfaceInput](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)

	if len(intfInput.GatewayId) > 0 && len(intfInput.InterfaceSettings.Name) > 0 {
		intf, _, err = apiSvc.EdgesApi.GetEdgeIfByName(
			ctx, intfInput.GatewayId, intfInput.InterfaceSettings.Name, nil)
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

	d.SetId(utils.Hash(intf))
	return diags
}

func (rt _resourceGatewayInterface) resourceGatewayInterfaceUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var intf swagger.InterfaceSettings

	intfInput, err := ApplyBinderInputResourceData[resourceGatewayInterfaceInput](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)

	if len(intfInput.GatewayId) > 0 && len(intfInput.InterfaceSettings.Name) > 0 {
		lock := utils.Mutex.Get(intfInput.GatewayId)
		lock.Lock()
		defer lock.Unlock()
		rt.fixupInterfaceConfig(&intfInput.InterfaceSettings)
		gateway, _, err := apiSvc.EdgesApi.UpdateEdgeIfByName(
			ctx, intfInput.InterfaceSettings, intfInput.GatewayId, intfInput.InterfaceSettings.Name, nil)

		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
		for _, i := range gateway.Interfaces {
			if i.Name == intfInput.InterfaceSettings.Name {
				intf = i
			}
		}

		d.SetId(utils.Hash(intf))
		err = ApplyBinderResourceData(rt.Binder, d, intf)
		if err != nil {
			return diag.FromErr(err)
		}
		return diags
	} else {
		return diag.FromErr(err)
	}
}

func (rt _resourceGatewayInterface) resourceGatewayInterfaceDelete(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var intf swagger.InterfaceSettings
	var err error

	intfInput, err := ApplyBinderInputResourceData[resourceGatewayInterfaceInput](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)

	if len(intfInput.GatewayId) > 0 && len(intfInput.InterfaceSettings.Name) > 0 {
		lock := utils.Mutex.Get(intfInput.GatewayId)
		lock.Lock()
		defer lock.Unlock()
		intf, _, err = apiSvc.EdgesApi.GetEdgeIfByName(
			ctx, intfInput.GatewayId, intfInput.InterfaceSettings.Name, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}

		// We cant delete the interface. So we are disabling it.
		intfInput.InterfaceSettings.IsDisabled = true

		gateway, _, err := apiSvc.EdgesApi.UpdateEdgeIfByName(ctx,
			intf,
			intfInput.GatewayId,
			intf.Name,
			nil,
		)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
		for _, i := range gateway.Interfaces {
			if i.Name == intfInput.InterfaceSettings.Name {
				intf = i
			}
		}
	} else {
		return diag.FromErr(err)
	}
	err = ApplyBinderResourceData(rt.Binder, d, intf)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}

type _resourceGatewayInterface struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

type resourceGatewayInterfaceInput struct {
	GatewayId string
	swagger.InterfaceSettings
}

func resourceGatewayInterface() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(resourceGatewayInterfaceInput{}, Cfg{
		"name":        {Schema: &schema.Schema{Required: true}},
		"gateway_id":  {Schema: &schema.Schema{Required: true}},
		"is_disabled": {Schema: &schema.Schema{Required: true}},
	})

	rt := _resourceGatewayInterface{Binder: binder, InputBinder: inputBinder}

	return &schema.Resource{
		CreateContext: rt.resourceGatewayInterfaceUpdate,
		ReadContext:   rt.resourceGatewayInterfaceRead,
		UpdateContext: rt.resourceGatewayInterfaceUpdate,
		DeleteContext: rt.resourceGatewayInterfaceDelete,
		Schema:        swaggerSchema,
	}
}
