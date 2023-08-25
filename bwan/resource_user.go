package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _resourceUser) resourceUserCreate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiSvc := m.(*swagger.APIClient)
	userInput, err := ApplyBinderInputResourceData[swagger.User](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}
	user, _, err := apiSvc.UsersApi.AddUser(ctx, userInput, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, user)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(user.Id)
	return diags

}

func (rt _resourceUser) resourceUserRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var user swagger.User
	var err error

	userInput, err := ApplyBinderInputResourceData[swagger.User](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)

	user, _, err = apiSvc.UsersApi.GetUserById(ctx, userInput.Id, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, user)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(user.Id)
	return diags
}

func (rt _resourceUser) resourceUserUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiSvc := m.(*swagger.APIClient)
	userInput, err := ApplyBinderInputResourceData[swagger.User](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}
	user, _, err := apiSvc.UsersApi.UpdateUserById(ctx, userInput, userInput.Id, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, user)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(user.Id)
	return diags
}

func (rt _resourceUser) resourceUserDelete(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiSvc := m.(*swagger.APIClient)
	userInput, err := ApplyBinderInputResourceData[swagger.User](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}
	user, _, err := apiSvc.UsersApi.DeleteUserById(ctx, userInput.Id, nil)
	if err != nil {
		if serr, ok := err.(swagger.GenericSwaggerError); ok {
			return diag.FromErr(fmt.Errorf("%s", serr.Body()))
		}
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, user)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

type _resourceUser struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

func resourceUser() *schema.Resource {
	swaggerSchema, binder, swaggerInputBinder := ReflectSchema(swagger.User{}, Cfg{
		"name": {
			Schema: &schema.Schema{
				Required: true,
			},
		},
	})

	rt := _resourceUser{Binder: binder, InputBinder: swaggerInputBinder}

	return &schema.Resource{
		CreateContext: rt.resourceUserCreate,
		ReadContext:   rt.resourceUserRead,
		UpdateContext: rt.resourceUserUpdate,
		DeleteContext: rt.resourceUserDelete,
		Schema:        swaggerSchema,
	}
}
