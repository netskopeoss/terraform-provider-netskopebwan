package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _resourcePolicy) resourcePolicyCreate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiSvc := m.(*swagger.APIClient)
	policyInput, err := ApplyBinderInputResourceData[swagger.Policy](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	addPolicyInput := swagger.AddPolicyInput{
		Name:   policyInput.Name,
		Hubs:   policyInput.Hubs,
		Config: policyInput.Config,
	}

	policy, _, err := apiSvc.PoliciesApi.AddPolicy(ctx, addPolicyInput, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, policy)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Id)
	return diags

}

func (rt _resourcePolicy) resourcePolicyRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	apiSvc := m.(*swagger.APIClient)
	policyInput, err := ApplyBinderInputResourceData[swagger.Policy](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	policy, _, err := apiSvc.PoliciesApi.GetPolicyById(ctx, policyInput.Id, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}
	err = ApplyBinderResourceData(rt.Binder, d, policy)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(policy.Id)
	return diags
}

func (rt _resourcePolicy) resourcePolicyUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	policyInput, err := ApplyBinderInputResourceData[swagger.Policy](rt.InputBinder, d)

	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)
	policy, _, err := apiSvc.PoliciesApi.UpdatePolicyById(ctx, policyInput, policyInput.Id, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, policy)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(policy.Id)
	return diags
}

func (rt _resourcePolicy) resourcePolicyDelete(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiSvc := m.(*swagger.APIClient)
	policyInput, err := ApplyBinderInputResourceData[swagger.Policy](rt.InputBinder, d)

	if err != nil {
		return diag.FromErr(err)
	}
	_, _, err = apiSvc.PoliciesApi.DeletePolicyById(ctx, policyInput.Id, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

type _resourcePolicy struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

func resourcePolicy() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(swagger.Policy{}, Cfg{
		"name": {Schema: schema.Schema{Required: true}},
	})

	rt := _resourcePolicy{Binder: binder, InputBinder: inputBinder}

	return &schema.Resource{
		CreateContext: rt.resourcePolicyCreate,
		ReadContext:   rt.resourcePolicyRead,
		UpdateContext: rt.resourcePolicyUpdate,
		DeleteContext: rt.resourcePolicyDelete,
		Schema:        swaggerSchema,
	}
}
