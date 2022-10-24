package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _resourceTenant) resourceTenantCreate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	apiSvc := m.(*swagger.APIClient)

	tenantInput, err := ApplyBinderInputResourceData[swagger.Tenant](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	tenant, _, err := apiSvc.TenantsApi.AddTenant(ctx, tenantInput, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, tenant)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tenant.Id)
	return diags
}

func (rt _resourceTenant) resourceTenantRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	apiSvc := m.(*swagger.APIClient)

	tenantInput, err := ApplyBinderInputResourceData[swagger.Tenant](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	tenant, _, err := apiSvc.TenantsApi.GetTenantById(ctx, tenantInput.Id, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}
	err = ApplyBinderResourceData(rt.Binder, d, tenant)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tenant.Id)
	return diags
}

func (rt _resourceTenant) resourceTenantUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	apiSvc := m.(*swagger.APIClient)
	tenantInput, err := ApplyBinderInputResourceData[swagger.Tenant](rt.InputBinder, d)
	if err != nil || tenantInput.Id == "" {
		return diag.FromErr(err)
	}

	tenant, _, err := apiSvc.TenantsApi.UpdateTenantById(ctx, tenantInput, tenantInput.Id, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, tenant)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tenant.Id)
	return diags
}

func (rt _resourceTenant) resourceTenantDelete(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiSvc := m.(*swagger.APIClient)

	tenantInput, err := ApplyBinderInputResourceData[swagger.Tenant](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, err = apiSvc.TenantsApi.DeleteTenantById(ctx, tenantInput.Id, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

type _resourceTenant struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

func resourceTenant() *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(swagger.Tenant{}, Cfg{
		"name": {Schema: schema.Schema{Required: true}},
	})

	rt := _resourceTenant{Binder: binder, InputBinder: inputBinder}

	return &schema.Resource{
		CreateContext: rt.resourceTenantCreate,
		ReadContext:   rt.resourceTenantRead,
		UpdateContext: rt.resourceTenantUpdate,
		DeleteContext: rt.resourceTenantDelete,
		Schema:        swaggerSchema,
	}
}
