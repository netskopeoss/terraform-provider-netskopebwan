package bwan

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestObject struct {
	Id          string
	NilName     *string
	Name        *string
	NilStrings  []string
	Strings0    []string
	Strings0Nil []string
	Strings1    []string
	Child       TestObjectChild
}

type TestObjectChild struct {
	CNilName     *string
	CName        *string
	CNilStrings  []string
	CStrings0    []string
	CStrings0Nil []string
	CStrings1    []string
}

func resourceObject(t *testing.T) *schema.Resource {
	swaggerSchema, binder, inputBinder := ReflectSchema(TestObject{}, Cfg{
		"strings0":         {EmptyIsNotNull: true},
		"child.c_strings0": {EmptyIsNotNull: true},
	})

	return &schema.Resource{
		CreateContext: func(ctx context.Context, d *schema.ResourceData, i2 interface{}) diag.Diagnostics {
			objectInput, err := ApplyBinderInputResourceData[TestObject](inputBinder, d)
			if err != nil {
				return diag.FromErr(err)
			}

			assert.Nil(t, objectInput.Strings0Nil)
			assert.NotNil(t, objectInput.Strings0)
			assert.Nil(t, objectInput.Child.CStrings0Nil)
			assert.NotNil(t, objectInput.Child.CStrings0)

			err = ApplyBinderResourceData(binder, d, objectInput)
			if err != nil {
				return diag.FromErr(err)
			}
			d.SetId("newid")
			return nil
		},
		ReadContext: func(ctx context.Context, d *schema.ResourceData, i2 interface{}) diag.Diagnostics {
			objectInput, err := ApplyBinderInputResourceData[TestObject](inputBinder, d)
			if err != nil {
				return diag.FromErr(err)
			}

			err = ApplyBinderResourceData(binder, d, objectInput)
			if err != nil {
				return diag.FromErr(err)
			}
			d.SetId(objectInput.Id)
			return nil
		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, i2 interface{}) diag.Diagnostics {
			objectInput, err := ApplyBinderInputResourceData[TestObject](inputBinder, d)
			if err != nil {
				return diag.FromErr(err)
			}

			err = ApplyBinderResourceData(binder, d, objectInput)
			if err != nil {
				return diag.FromErr(err)
			}

			return nil
		},
		DeleteContext: func(ctx context.Context, d *schema.ResourceData, i2 interface{}) diag.Diagnostics {
			d.SetId("")
			return nil
		},
		Schema: swaggerSchema,
	}
}

func TestTF(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"tst": func() (*schema.Provider, error) {
				return &schema.Provider{
					ResourcesMap: map[string]*schema.Resource{
						"tst": resourceObject(t),
					},
				}, nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
				provider "tst" {}
				resource "tst" "blah" {
					#nil_name = null
					name = "blah"
					#nil_strings = null
					strings0 = []
					strings0_nil = []
					strings1 = ["heyy"]

					child {
						#c_nil_name = null
						c_name = "blah"
						#c_nil_strings = null
						c_strings0 = []
						c_strings0_nil = []
						c_strings1 = ["heyy"]
					}
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tst.blah", "name", "blah"),

					// Check that the key is actually not present
					resource.TestCheckNoResourceAttr("tst.blah", "nil_strings"),
					// Check that the key exist and is an empty list
					resource.TestCheckResourceAttr("tst.blah", "strings0.#", "0"),
					// Check that the key exist and is a list with 1 item
					resource.TestCheckResourceAttr("tst.blah", "strings1.#", "1"),

					// Check that the key is actually present
					resource.TestCheckResourceAttr("tst.blah", "child.0.c_nil_strings.#", "0"),
					// Check that the key exist and is an empty list
					resource.TestCheckResourceAttr("tst.blah", "child.0.c_strings0.#", "0"),
					// Check that the key exist and is a list with 1 item
					resource.TestCheckResourceAttr("tst.blah", "child.0.c_strings1.#", "1"),

					// Here to add a breakpoint and peek into the state
					func(state *terraform.State) error {
						p := state.RootModule().Resources["tst.blah"].Primary
						_ = p
						return nil
					},
				),
			},
		},
	})
}
