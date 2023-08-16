package bwan

import (
	"context"
	"fmt"

	swagger "github.com/infiotinc/netskopebwan-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (rt _dataSourceUser) dataSourceUserRead(
	ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	var user swagger.User
	var err error

	userInput, err := ApplyBinderInputResourceData[swagger.User](rt.InputBinder, d)
	if err != nil {
		return diag.FromErr(err)
	}

	apiSvc := m.(*swagger.APIClient)

	if len(userInput.Id) > 0 {
		user, _, err = apiSvc.UsersApi.GetUserById(ctx, userInput.Id, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
	} else if len(userInput.Name) > 0 {
		userList, _, err := apiSvc.UsersApi.GetAllUsers(ctx, nil)
		if err != nil {
			if serr, ok := err.(swagger.GenericSwaggerError); ok {
				return diag.FromErr(fmt.Errorf("%s", serr.Body()))
			}
			return diag.FromErr(err)
		}
		for _, u := range userList {
			if user.Name == userInput.Name {
				user = u
				break
			}
		}
		if len(user.Name) == 0 {
			return diag.FromErr(err)
		}
	} else {
		return diag.FromErr(err)
	}

	err = ApplyBinderResourceData(rt.Binder, d, user)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(user.Id)
	return diags
}

type _dataSourceUser struct {
	Binder      []FieldBinder
	InputBinder []FieldBinder
}

func dataSourceUser() *schema.Resource {
	swaggerSchema, binder, swaggerInputBinder := ReflectSchema(swagger.User{}, Cfg{
		"name": {
			Schema: &schema.Schema{
				Optional: true,
			},
		},
	})

	rt := _dataSourceUser{Binder: binder, InputBinder: swaggerInputBinder}

	return &schema.Resource{
		ReadContext: rt.dataSourceUserRead,
		Schema:      swaggerSchema,
	}
}
