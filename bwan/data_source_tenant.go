package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"

	"github.com/antihax/optional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _dataSourceTenant) dataSourceTenantsRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	var err error
	var tenant swagger.Tenant

	apiSvc := m.(*swagger.APIClient)

	tenantInput, err := ApplyBinderInputResourceData[swagger.Tenant](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	tenantQueryOpts := swagger.TenantsApiGetAllTenantsOpts{
		MaxItems: optional.NewInt32(10000),
	}

	if len(tenantInput.Id) > 0 {
		tenant, _, err = apiSvc.TenantsApi.GetTenantById(ctx, tenantInput.Id, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
	} else if len(tenantInput.Name) > 0 {
		tenantList, _, err := apiSvc.TenantsApi.GetAllTenants(ctx, &tenantQueryOpts)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
		for _, t := range tenantList.Data {
			if t.Name == tenantInput.Name {
				tenant = t
				break
			}
		}
	}
	err = ApplyBinderResourceData(rt.Binder, d, tenant)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tenant.Id)
	return diags
}

type _dataSourceTenant struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

func dataSourceTenant() *schema.Resource {
	swaggerSchema, binder, swaggerInputBinder := ReflectSchema(swagger.Tenant{}, Cfg{
		"name": {
			Schema: schema.Schema{
				Optional: true,
			},
		},
	})

	rt := _dataSourceTenant{Binder: binder, InputBinder: swaggerInputBinder}

	return &schema.Resource{
		ReadContext: rt.dataSourceTenantsRead,
		Schema:      swaggerSchema,
	}
}
