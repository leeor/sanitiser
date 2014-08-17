package sanitiser

import (
	"fmt"
	"reflect"
	"strings"
)

type Logger func(format string, v ...interface{})

type Sanitiser interface {
	Sanitise(context string)
}

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

	if v, ok = obj.(reflect.Value); !ok {

		v = reflect.ValueOf(obj)
	}

	dbg("%v.%v(type %T)\n", hierarchy, v, obj)

	for v.Kind() == reflect.Ptr && reflect.Indirect(v).Kind() == reflect.Interface {

		dbg("object is a pointer to an interface, calling Elem()\n")
		v = v.Elem()
	}

	// Start by calling the Sanitise method if the object has the Sanitiser
	// interface
	if s, ok := v.Interface().(Sanitiser); ok {

		dbg("Object of type %T supports the Sanitiser interface, invoking Sanitise()\n", v.Interface())
		s.Sanitise(context)
	} else if v.CanAddr() {

		if s, ok := v.Addr().Interface().(Sanitiser); ok {

			dbg("Object of type %T supports the Sanitiser interface, invoking Sanitise()\n", v.Interface())
			s.Sanitise(context)
		}
	} else {

		dbg("Object of type %T does not supports the Sanitiser interface\n", v.Interface())
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {

		dbg("object is a pointer or an interface, calling Elem()\n")
		v = v.Elem()
	}

	if !v.IsValid() {

		return nil
	}

	dbg("%v.%v(type %T)\n", hierarchy, v, v.Interface())

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
	} else if (k == reflect.Slice || k == reflect.Array) &&
		(v.Len() > 0) &&
		((v.Index(0).Kind() == reflect.Struct) || ((v.Index(0).Kind() == reflect.Ptr) && (reflect.Indirect(v.Index(0)).Kind() == reflect.Struct))) {

		dbg("Processing list %v.%v(type %T)\n", hierarchy, v, v.Interface())

		for i := 0; i < v.Len(); i++ {

			dbg("Processing list %v.%v(type %T) item #%v\n", hierarchy, v, v.Interface(), i)

			if err := traverseObjects(v.Index(i), context, fmt.Sprint(hierarchy, "[", i, "]")); err != nil {

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
				dbg("Field %v.%v(type %T) has a sanitise tag\n", hierarchy, field.Name, v.Field(i).Interface())
				contexts := parseTag(tag)
				if contains(contexts, context) || contains(contexts, "*") {
					// sanitise this field
					if !v.Field(i).CanSet() {

						return fmt.Errorf("Unable to set zero value for %v.%v", hierarchy, t.Field(i).Name)
					}

					dbg("Sanitising field %v.%v\n", hierarchy, t.Field(i).Name)
					v.Field(i).Set(reflect.New(t.Field(i).Type).Elem())

					// no point in continuing to traverse this field, even if
					// it's a struct of some sort, it was assigned the zero
					// value.
					continue
				}
			}

			if field_kind == reflect.Struct ||
				field_kind == reflect.Interface ||
				field_kind == reflect.Ptr ||
				field_kind == reflect.Map ||
				field_kind == reflect.Slice ||
				field_kind == reflect.Array {

				var sv reflect.Value

				if field_kind != reflect.Ptr {

					sv = v.Field(i).Addr()
				} else {

					sv = v.Field(i)
				}

				if !sv.IsNil() {

					dbg("Processing object %v.%v(type %T)\n", hierarchy, sv, sv.Interface())

					if err := traverseObjects(sv, context, hierarchy+"."+t.Field(i).Name); err != nil {

						return err
					}
				}
			}
		}
	}

	return nil
}

func Sanitise(obj interface{}, context string) error {

	return traverseObjects(obj, context, "")
}

func nullLogger(format string, v ...interface{}) {
}

func SetLogger(f Logger) {

	if f != nil {

		dbg = f
	} else {

		dbg = nullLogger
	}
}
