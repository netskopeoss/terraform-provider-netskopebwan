package bwan

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

const SchemaStruct = schema.ValueType(-1)

func ToSnakeCase(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	var str string
	if tag == "" || tag == "-" {
		str = field.Name
	} else {
		str = strings.Split(tag, ",")[0]
	}
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

type Cfg map[string]struct {
	schema.Schema
}

func ReflectSchema(v interface{}, cfg Cfg) (map[string]*schema.Schema, []FieldBinder, []FieldBinder) {
	t := reflect.TypeOf(v)
	return reflectSchemaType("", t, cfg)
}

type FieldBinder struct {
	MapKey    string
	FieldName string
	Func      BinderFunc
}

func ApplyBinderMap(bm []FieldBinder, v reflect.Value) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	err := ApplyBinder(bm, v, func(key string, v interface{}) error {
		m[key] = v

		return nil
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func ApplyBinderInputResourceData[T any](bm []FieldBinder, d *schema.ResourceData) (T, error) {
	return ApplyBinderInput[T](bm, d.GetOk)
}

func ApplyBinderInput[T any](bm []FieldBinder, get func(k string) (interface{}, bool)) (T, error) {
	var tzero T
	t := reflect.TypeOf(tzero)

	to, err := applyBinderInput(t, bm, get)
	if err != nil {
		return tzero, err
	}

	return to.Interface().(T), nil
}

func applyBinderInput(t reflect.Type, bm []FieldBinder, get func(k string) (interface{}, bool)) (reflect.Value, error) {
	nvp := reflect.New(t)
	nv := nvp.Elem()

	for _, b := range bm {
		f := nv.FieldByName(b.FieldName)

		if !f.IsValid() {
			panic(fmt.Sprintf("field %v does not exist on %v", b.FieldName, nv.Type().String()))
		}

		mv, ok := get(b.MapKey)
		if !ok || mv == nil {
			continue
		}

		iv, err := b.Func(reflect.ValueOf(mv))
		if err != nil {
			return reflect.Value{}, err
		}
		if iv == nil {
			continue
		}
		v := reflect.ValueOf(iv)

		switch f.Type().Kind() {
		case reflect.Array, reflect.Slice:
			nv := reflect.MakeSlice(reflect.SliceOf(f.Type().Elem()), 0, 0)

			for i := 0; i < v.Len(); i++ {
				nv = reflect.Append(nv, v.Index(i).Convert(f.Type().Elem()))
			}

			v = nv
		}

		f.Set(v.Convert(f.Type()))
	}

	return nv, nil
}

func safeResourceDataSet(d *schema.ResourceData, k string, v interface{}) (rerr error) {
	defer func() {
		if r := recover(); r != nil {
			rerr = fmt.Errorf("panic: %v", r)
		}
	}()

	return d.Set(k, v)
}

func ApplyBinderResourceData(bm []FieldBinder, d *schema.ResourceData, v interface{}) error {
	return ApplyBinder(bm, reflect.ValueOf(v), func(key string, v interface{}) error {
		if err := safeResourceDataSet(d, key, v); err != nil {
			return fmt.Errorf("%v: %w", key, err)
		}

		return nil
	})
}

func ApplyBinder(bm []FieldBinder, v reflect.Value, do func(key string, v interface{}) error) error {
	for _, b := range bm {
		f := v.FieldByName(b.FieldName)

		if !f.IsValid() {
			//panic(fmt.Sprintf("field %v does not exist on %v", b.FieldName, v.Type().String()))
			continue
		}

		fv, err := b.Func(f)
		if err != nil {
			return err
		}

		err = do(b.MapKey, fv)
		if err != nil {
			return err
		}
	}

	return nil
}

func reflectSchemaType(path string, t reflect.Type, cfg Cfg) (map[string]*schema.Schema, []FieldBinder, []FieldBinder) {
	if t.Kind() == reflect.Ptr {
		panic("ptr not supported")
	}

	s := map[string]*schema.Schema{}
	var bm []FieldBinder
	var ibm []FieldBinder

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Anonymous {
			t := field.Type
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}

			if t.Kind() == reflect.Struct {
				es, ebm, eibm := reflectSchemaType(path, t, cfg)

				for k, v := range es {
					s[k] = v
				}
				bm = append(bm, ebm...)
				ibm = append(ibm, eibm...)

				continue
			}

			panic("unhandled anonymous field " + t.Kind().String())
		}

		name := ToSnakeCase(field)

		var fpath string
		if path == "" || field.Anonymous {
			fpath = name
		} else {
			fpath = path + "." + name
		}

		fs, b, ib := reflectSchemaField(fpath, cfg, field.Type, true, false)
		s[name] = fs
		bm = append(bm, FieldBinder{
			MapKey:    name,
			FieldName: field.Name,
			Func:      b,
		})

		if ib != nil {
			ibm = append(ibm, FieldBinder{
				MapKey:    name,
				FieldName: field.Name,
				Func:      ib,
			})
		}
	}

	return s, bm, ibm
}

func reflectSchemaField(path string, cfg Cfg, t reflect.Type, extra, allowDirectObject bool) (*schema.Schema, BinderFunc, BinderFunc) {
	fcfg, ok := cfg[path]

	s := fcfg.Schema
	var b, ib BinderFunc
	s.Type, s.Elem, b, ib = reflectSchemaFieldType(path, t, cfg, allowDirectObject)
	if extra && !ok {
		s.Optional = true
		s.Computed = true
	}

	if s.Type == schema.TypeSet {
		s.MaxItems = 1
	}

	return &s, b, ib
}

type BinderFunc func(v reflect.Value) (interface{}, error)

func defaultBinder(t reflect.Type) BinderFunc {
	return func(v reflect.Value) (interface{}, error) {
		return v.Convert(t).Interface(), nil
	}
}

func reflectSchemaFieldType(path string, t reflect.Type, cfg Cfg, allowDirectObject bool) (schema.ValueType, interface{}, BinderFunc, BinderFunc) {
	switch t.Kind() {
	case reflect.Ptr:
		st, elem, b, ib := reflectSchemaFieldType(path, t.Elem(), cfg, allowDirectObject)
		return st, elem, func(v reflect.Value) (interface{}, error) {
				if v.IsNil() {
					return nil, nil
				}

				vp, err := b(v.Elem())
				if err != nil {
					return nil, err
				}

				return vp, nil
			}, func(v reflect.Value) (interface{}, error) {
				iv, err := ib(v)
				if err != nil {
					return nil, err
				}

				if iv == nil {
					return nil, nil
				}

				p := reflect.New(t.Elem())
				p.Elem().Set(reflect.ValueOf(iv).Convert(t.Elem()))

				return p.Interface(), nil
			}
	case reflect.String:
		return schema.TypeString, nil, func(v reflect.Value) (interface{}, error) {
			return v.String(), nil
		}, defaultBinder(t)
	case reflect.Bool:
		return schema.TypeBool, nil, func(v reflect.Value) (interface{}, error) {
			return v.Bool(), nil
		}, defaultBinder(t)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return schema.TypeInt, nil, func(v reflect.Value) (interface{}, error) {
			return v.Int(), nil
		}, defaultBinder(t)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return schema.TypeInt, nil, func(v reflect.Value) (interface{}, error) {
			return v.Uint(), nil
		}, defaultBinder(t)
	case reflect.Float64, reflect.Float32:
		return schema.TypeFloat, nil, func(v reflect.Value) (interface{}, error) {
			return v.Float(), nil
		}, defaultBinder(t)
	case reflect.Slice, reflect.Array:
		s, b, ib := reflectSchemaField(path, cfg, t.Elem(), false, true)

		var innerType interface{} = s
		if s.Type == SchemaStruct {
			innerType = s.Elem
		}

		return schema.TypeList, innerType, func(v reflect.Value) (interface{}, error) {
				if v.IsZero() {
					return nil, nil
				}

				a := make([]interface{}, 0)
				for i := 0; i < v.Len(); i++ {
					iv, err := b(v.Index(i))
					if err != nil {
						return nil, err
					}

					a = append(a, iv)
				}

				return a, nil
			}, func(v reflect.Value) (interface{}, error) {
				if v.IsZero() {
					return nil, nil
				}

				av := v.Interface().([]interface{})
				nv := reflect.MakeSlice(t, 0, len(av))

				for _, iv := range av {
					niv, err := ib(reflect.ValueOf(iv))
					if err != nil {
						return nil, err
					}

					nv = reflect.Append(nv, reflect.ValueOf(niv))
				}

				return nv.Interface(), nil
			}
	case reflect.Struct:
		if t.AssignableTo(reflect.TypeOf(time.Time{})) {
			return schema.TypeString, nil, func(v reflect.Value) (interface{}, error) {
					t := v.Interface().(time.Time)

					return t.Format(time.RFC3339), nil
				}, func(v reflect.Value) (interface{}, error) {
					sv := v.Interface().(string)

					return time.Parse(time.RFC3339, sv)
				}
		}

		if t.PkgPath() == "github.com/infiotinc/netskopebwan-go-client" ||
			t.PkgPath() == "github.com/netskopeoss/terraform-provider-netskopebwan/bwan" ||
			t.PkgPath() == "main" {
			s, bm, ibm := reflectSchemaType(path, t, cfg)

			st := schema.TypeSet
			if allowDirectObject {
				st = SchemaStruct
			}

			return st, &schema.Resource{Schema: s}, func(v reflect.Value) (interface{}, error) {
					m, err := ApplyBinderMap(bm, v)
					if err != nil {
						return nil, err
					}

					if allowDirectObject {
						return m, nil
					}

					return []interface{}{m}, nil
				}, func(v reflect.Value) (interface{}, error) {
					var m map[string]interface{}
					if allowDirectObject {
						m = v.Interface().(map[string]interface{})
					} else {
						var l []interface{}
						switch vi := v.Interface().(type) {
						case []interface{}:
							l = vi
						case *schema.Set:
							l = vi.List()
						default:
							panic(fmt.Sprintf("unsupported type %T", vi))
						}

						if len(l) < 1 {
							return nil, nil
						}

						m = l[0].(map[string]interface{})
					}

					vo, err := applyBinderInput(t, ibm, func(k string) (interface{}, bool) {
						v, ok := m[k]
						return v, ok
					})
					if err != nil {
						return nil, err
					}

					return vo.Interface(), nil
				}
		}
	}

	panic(fmt.Sprintf("unahandled type %v: %v", t.PkgPath(), t.String()))
}
