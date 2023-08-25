package bwan

import (
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Sanity struct {
	String string

	Bool bool

	Int   int
	Int16 int16
	Int64 int64

	UInt   uint
	UInt16 uint16
	UInt64 uint64

	Float32 float32
	Float64 float64
}

var SanitySchema = ms{
	"string":  &schema.Schema{Type: schema.TypeString, Optional: true, Computed: true},
	"bool":    &schema.Schema{Type: schema.TypeBool, Optional: true, Computed: true},
	"int":     &schema.Schema{Type: schema.TypeInt, Optional: true, Computed: true},
	"int16":   &schema.Schema{Type: schema.TypeInt, Optional: true, Computed: true},
	"int64":   &schema.Schema{Type: schema.TypeInt, Optional: true, Computed: true},
	"u_int":   &schema.Schema{Type: schema.TypeInt, Optional: true, Computed: true},
	"u_int16": &schema.Schema{Type: schema.TypeInt, Optional: true, Computed: true},
	"u_int64": &schema.Schema{Type: schema.TypeInt, Optional: true, Computed: true},
	"float32": &schema.Schema{Type: schema.TypeFloat, Optional: true, Computed: true},
	"float64": &schema.Schema{Type: schema.TypeFloat, Optional: true, Computed: true},
}

type NestedObject struct {
	Child NestedObjectChild
}

var NestedObjectSchema = ms{
	"child": &schema.Schema{
		Type:     schema.TypeSet,
		Elem:     &schema.Resource{Schema: NestedObjectChildSchema},
		MaxItems: 1,
		Optional: true,
		Computed: true,
	},
}

type ArrayObject struct {
	Children []NestedObjectChild
}

var ArrayObjectSchema = ms{
	"children": &schema.Schema{
		Type:     schema.TypeList,
		Elem:     &schema.Resource{Schema: NestedObjectChildSchema},
		Optional: true,
		Computed: false,
	},
}

type NestedObjectChild struct {
	Id string
}

var NestedObjectChildSchema = ms{
	"id": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
}

type ArrayPrimitive struct {
	Strings []string
}

var ArrayPrimitiveSchema = ms{
	"strings": &schema.Schema{
		Type:     schema.TypeList,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Optional: true,
		Computed: false,
	},
}

type PointerObject struct {
	Child *NestedObjectChild
}

var PointerObjectSchema = ms{
	"child": &schema.Schema{
		Type:     schema.TypeSet,
		Elem:     &schema.Resource{Schema: NestedObjectChildSchema},
		MaxItems: 1,
		Optional: true,
		Computed: true,
	},
}

type EmbedObject struct {
	ParentId string
	NestedObjectChild
}

var EmbedObjectSchema = ms{
	"parent_id": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"id": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
}

type ComplexObject struct {
	Sub []ArrayObject
}

type i = interface{}
type m = map[string]i
type ms = map[string]*schema.Schema

func TestSchema(t *testing.T) {
	tests := []struct {
		name string
		t    interface{}
		s    ms
	}{
		{"sanity", Sanity{}, SanitySchema},
		{"array primitive", ArrayPrimitive{}, ArrayPrimitiveSchema},
		{"nested objects", NestedObject{}, NestedObjectSchema},
		{"array object", ArrayObject{}, ArrayObjectSchema},
		{"pointer object", PointerObject{}, PointerObjectSchema},
		{"embed object", EmbedObject{}, EmbedObjectSchema},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sch, _, _ := ReflectSchema(test.t, Cfg{})

			assert.Equal(t, test.s, sch)
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		t    interface{}
		out  map[string]interface{}
	}{
		{"sanity", Sanity{}, m{
			"string":  "",
			"bool":    false,
			"int":     int64(0),
			"int16":   int64(0),
			"int64":   int64(0),
			"u_int":   uint64(0),
			"u_int16": uint64(0),
			"u_int64": uint64(0),
			"float32": float64(0),
			"float64": float64(0),
		}},
		{"sanity", Sanity{
			String: "str",
			Bool:   true,
			Int:    -1,
			Int16:  -2,
			Int64:  -3,
			UInt:   4,
			UInt16: 5,
			UInt64: 6,
			//Float32: -1.1, # float imprecision
			Float64: -1.2,
		}, m{
			"string":  "str",
			"bool":    true,
			"int":     int64(-1),
			"int16":   int64(-2),
			"int64":   int64(-3),
			"u_int":   uint64(4),
			"u_int16": uint64(5),
			"u_int64": uint64(6),
			"float32": float64(0),
			"float64": float64(-1.2),
		}},
		{"sanity", Sanity{
			String: "str",
			Bool:   true,
			Int:    1,
			Int16:  2,
			Int64:  3,
		}, m{
			"string":  "str",
			"bool":    true,
			"int":     int64(1),
			"int16":   int64(2),
			"int64":   int64(3),
			"u_int":   uint64(0),
			"u_int16": uint64(0),
			"u_int64": uint64(0),
			"float32": float64(0),
			"float64": float64(0),
		}},
		{"array primitives", ArrayPrimitive{
			Strings: nil,
		}, m{
			"strings": nil,
		}},
		{"array primitives EINN", ArrayPrimitive{
			Strings: []string{},
		}, m{
			"strings": []i{},
		}},
		{"array primitives NOTSYMMETRICAL", ArrayPrimitive{
			Strings: []string{},
		}, m{
			"strings": nil,
		}},
		{"array primitives", ArrayPrimitive{
			Strings: []string{""},
		}, m{
			"strings": []i{""},
		}},
		{"array primitives", ArrayPrimitive{
			Strings: []string{"1"},
		}, m{
			"strings": []i{"1"},
		}},
		{"array primitives", ArrayPrimitive{
			Strings: []string{"1", "2", "3"},
		}, m{
			"strings": []i{"1", "2", "3"},
		}},
		{"array object", ArrayObject{
			Children: nil,
		}, m{
			"children": nil,
		}},
		{"array object NOTSYMMETRICAL", ArrayObject{
			Children: []NestedObjectChild{},
		}, m{
			"children": nil,
		}},
		{"array object EINN", ArrayObject{
			Children: []NestedObjectChild{},
		}, m{
			"children": []i{},
		}},
		{"array object", ArrayObject{
			Children: []NestedObjectChild{{}},
		}, m{
			"children": []i{m{"id": ""}},
		}},
		{"array object", ArrayObject{
			Children: []NestedObjectChild{{Id: "id"}},
		}, m{
			"children": []i{m{"id": "id"}},
		}},
		{"nested object", NestedObject{}, m{
			"child": []i{m{"id": ""}},
		}},
		{"nested object", NestedObject{
			Child: NestedObjectChild{Id: "id"},
		}, m{
			"child": []i{m{"id": "id"}},
		}},
		{"pointer object", PointerObject{}, m{
			"child": nil,
		}},
		{"pointer object", PointerObject{
			Child: &NestedObjectChild{},
		}, m{
			"child": []i{m{"id": ""}},
		}},
		{"pointer object", PointerObject{
			Child: &NestedObjectChild{Id: "id"},
		}, m{
			"child": []i{m{"id": "id"}},
		}},
		{"embed object", EmbedObject{}, m{
			"parent_id": "",
			"id":        "",
		}},
		{"embed object", EmbedObject{NestedObjectChild: NestedObjectChild{Id: "id"}}, m{
			"parent_id": "",
			"id":        "id",
		}},
		{"complex nesting", ComplexObject{
			Sub: []ArrayObject{
				{
					Children: []NestedObjectChild{{Id: "1"}, {Id: "2"}},
				},
				{
					Children: nil,
				},
			},
		}, m{"sub": []i{
			m{"children": []i{m{"id": "1"}, m{"id": "2"}}},
			m{"children": nil},
		}}},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			isEmptyNotNil := strings.Contains(test.name, "EINN")
			isNotSymetrical := strings.Contains(test.name, "NOTSYMMETRICAL")

			cfg := Cfg{}
			if isEmptyNotNil {
				c := cfg["*"]
				c.EmptyIsNotNull = true
				cfg["*"] = c
			}
			_, bm, ibm := ReflectSchema(test.t, cfg)

			ma, err := ApplyBinderMap(bm, reflect.ValueOf(test.t))
			require.NoError(t, err)

			assert.EqualValues(t, test.out, ma)

			v, err := applyBinderInput(reflect.TypeOf(test.t), ibm, func(k string) (interface{}, bool) {
				v, ok := ma[k]
				return v, ok
			})
			require.NoError(t, err)

			if isNotSymetrical {
				return
			}

			av := v.Interface()
			assert.EqualValues(t, test.t, av)
		})
	}
}
