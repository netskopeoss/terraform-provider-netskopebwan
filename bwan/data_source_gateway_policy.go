package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _dataSourcePolicy) dataSourcePolicyRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	var policy swagger.Policy
	var err error

	policyInput, err := ApplyBinderInputResourceData[swagger.Policy](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)

	if len(policyInput.Id) > 0 {
		policy, _, err = apiSvc.PoliciesApi.GetPolicyById(ctx, policyInput.Id, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
	} else if len(policyInput.Name) > 0 {
		policyList, _, err := apiSvc.PoliciesApi.GetAllPolicies(ctx, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
		for _, pol := range policyList {
			if pol.Name == policyInput.Name {
				policy = pol
				break
			}
		}
		if len(policy.Name) == 0 {
			return diag.FromErr(err)
		}
	} else {
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, policy)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(policy.Id)
	return diags

}

type _dataSourcePolicy struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

func dataSourcePolicy() *schema.Resource {
	swaggerSchema, binder, swaggerInputBinder := ReflectSchema(swagger.Policy{}, Cfg{
		"name": {Schema: &schema.Schema{Required: true}},
	})

	rt := _dataSourcePolicy{Binder: binder, InputBinder: swaggerInputBinder}

	return &schema.Resource{
		ReadContext: rt.dataSourcePolicyRead,
		Schema:      swaggerSchema,
	}
}
