package sanitiser

import (
	"fmt"
	"reflect"
	"strings"
)

type Logger func(format string, v ...interface{})

var dbg Logger

func init() {

	dbg = func(string, ...interface{}) {}
}

func parseTag(tag string) []string {

	return strings.Split(tag, ",")
}

func contains(contexts []string, context string) bool {

	for _, d := range contexts {

		if d == context {

			return true
		}
	}

	return false
}

func traverseObjects(obj interface{}, context string, hierarchy string) error {

	// TODO: improve debug messages

	var v reflect.Value
	var t reflect.Type
	var ok bool

	dbg("%v.%v(type %T)\n", hierarchy, obj, obj)

	// make sure this is a pointer, so that we can update the contents if needed
	if v, ok = obj.(reflect.Value); !ok {

		v = reflect.ValueOf(obj)
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {

		dbg("object is a pointer or an interface, calling Elem()\n")
		v = v.Elem()
	}

	if !v.IsValid() {

		return nil
	}

	dbg("%v.%v(type %T)\n", hierarchy, v, v)

	t = reflect.TypeOf(v.Interface())
	k := t.Kind()

	if k == reflect.Map {

		keys := v.MapKeys()
		for _, key := range keys {

			dbg("Processing object %v.%v[%v]\n", hierarchy, t.Name(), key)
			if err := traverseObjects(v.MapIndex(key), context, hierarchy+"["+fmt.Sprintf("%v", key)+"]"); err != nil {

				return err
			}
		}
	} else if k == reflect.Struct {

		for i := 0; i < t.NumField(); i++ {

			dbg("Processing field %v.%v(%v)\n", hierarchy, t.Field(i).Name, t.Field(i).Type)
			field := t.Field(i)
			field_kind := field.Type.Kind()

			if tag := field.Tag.Get("sanitise"); len(tag) > 0 {

				// the sanitise tag's value should be a comma-separated list of
				// contexts
				dbg("Field %v.%v(type %T) has a sanitise tag\n", hierarchy, field.Name, v.Field(i))
				contexts := parseTag(tag)
				if contains(contexts, context) || contains(contexts, "*") {
					// sanitise this field
					if !v.Field(i).CanSet() {

						return fmt.Errorf("Unable to set zero value for %v.%v", hierarchy, t.Field(i).Name)
					}

					dbg("Sanitising field %v.%v\n", hierarchy, t.Field(i).Name)
					v.Field(i).Set(reflect.New(t.Field(i).Type).Elem())
				}
			}

			if field_kind == reflect.Struct || field_kind == reflect.Interface || field_kind == reflect.Ptr || field_kind == reflect.Map {

				sv := v.Field(i)
				dbg("Processing object %v.%v(type %T)\n", hierarchy, sv, sv)

				if err := traverseObjects(sv, context, hierarchy+"."+t.Field(i).Name); err != nil {

					return err
				}
			}
		}
	}

	return nil
}

func Sanitise(obj interface{}, context string) error {

	return traverseObjects(obj, context, "")
}

func SetLogger(f Logger) {

	dbg = f
}
