package main

import (
	"fmt"
	"reflect"
	"strings"
)

func parseTag(tag string) []string {

	return strings.Split(tag, ",")
}

func contains(domains []string, domain string) bool {

	for _, d := range domains {

		if d == domain {

			return true
		}
	}

	return false
}

func Sanitise(obj interface{}, domain string) error {

	var v reflect.Value
	var t reflect.Type
	var ok bool

	fmt.Printf("%v(type %T)\n", obj, obj)

	// make sure this is a pointer, so that we can update the contents if needed
	if v, ok = obj.(reflect.Value); !ok {

		v = reflect.ValueOf(obj)
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {

		fmt.Println("object is a pointer or an interface, calling Elem()")
		v = v.Elem()
	}

	if !v.IsValid() {

		return nil
	}

	fmt.Printf("%v(type %T)\n", v, v)

	t = reflect.TypeOf(v.Interface())
	k := t.Kind()

	if k == reflect.Map {

		keys := v.MapKeys()
		for _, key := range keys {

			Sanitise(v.MapIndex(key), domain)
		}
	} else if k == reflect.Struct {

		for i := 0; i < t.NumField(); i++ {

			fmt.Printf("Processing field %v(%v)\n", t.Field(i).Type, t.Field(i).Name)
			field := t.Field(i)
			field_kind := field.Type.Kind()

			if tag := field.Tag.Get("sanitise"); len(tag) > 0 {

				// the sanitise tag's value should be a comma-separated list of
				// domains
				fmt.Printf("Field %v(type %T) has a sanitise tag\n", field.Name, v.Field(i))
				domains := parseTag(tag)
				if contains(domains, domain) || contains(domains, "*") {
					// sanitise this field
					if !v.Field(i).CanSet() {

						return fmt.Errorf("Unable to set zero value for %v", t.Field(i).Name)
					}

					fmt.Printf("Sanitising field %v\n", t.Field(i).Name)
					v.Field(i).Set(reflect.New(t.Field(i).Type).Elem())
				}
			} else if field_kind == reflect.Struct || field_kind == reflect.Interface {

				sv := v.Field(i)
				fmt.Printf("%v(type %T)\n", sv, sv)

				if err := Sanitise(sv, domain); err != nil {

					return err
				}
			}
		}
	}

	return nil
}
