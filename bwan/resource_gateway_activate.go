package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"
	"github.com/netskopeoss/terraform-provider-netskopebwan/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _resourceGatewayActivate) resourceGatewayActivateCreate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiSvc := m.(*swagger.APIClient)
	gwActivationInput, err := ApplyBinderInputResourceData[dataSourceGatewayActivationInput](
		rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiInput := swagger.EdgeGenerateActivationCodeInput{
		EmailAddresses:   gwActivationInput.EmailAddresses,
		TimeoutInSeconds: gwActivationInput.TimeoutInSeconds,
	}

	gwActivationData, _, err := apiSvc.EdgesApi.ActivateEdgeById(
		ctx, apiInput, gwActivationInput.GatewayId, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	gwActivationInput.Token = gwActivationData.Token

	err = ApplyBinderResourceData(rt.Binder, d, gwActivationInput)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.Hash(gwActivationData))
	return diags

}

func (rt _resourceGatewayActivate) resourceGatewayActivateRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func (rt _resourceGatewayActivate) resourceGatewayActivateUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func (rt _resourceGatewayActivate) resourceGatewayActivateDelete(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

type _resourceGatewayActivate struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

type dataSourceGatewayActivationInput struct {
	GatewayId string `json:"gateway_id,omitempty"`
	swagger.EdgeGenerateActivationCodeInput
	swagger.EdgeActivationCode
}

func resourceGatewayActivate() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(dataSourceGatewayActivationInput{}, Cfg{
		"gateway_id": {Schema: schema.Schema{Required: true}},
	})
	rt := _resourceGatewayActivate{Binder: binder, InputBinder: inputBinder}

	return &schema.Resource{
		CreateContext: rt.resourceGatewayActivateCreate,
		ReadContext:   rt.resourceGatewayActivateRead,
		UpdateContext: rt.resourceGatewayActivateUpdate,
		DeleteContext: rt.resourceGatewayActivateDelete,
		Schema:        swaggerSchema,
	}
}
