package sanitiser

import (
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

func Sanitise(obj interface{}, context string) interface{} {
	// Wrap the original in a reflect.Value
	original := reflect.ValueOf(obj)

	sanitised := reflect.New(original.Type()).Elem()
	sanitiseRecursive(sanitised, original, context)

	// Remove the reflection wrapper
	return sanitised.Interface()
}

func shouldSanitiseField(field reflect.Value, structField reflect.StructField, context string) bool {

	if !field.CanSet() {
		// Can sub-fields of a non-settable field be settable themselves?
		return false
	}

	if tag := structField.Tag.Get("sanitise"); len(tag) > 0 {
		// the sanitise tag's value should be a comma-separated list of
		// contexts
		contexts := parseTag(tag)
		if contains(contexts, context) || contains(contexts, "*") {
			return true
		}
	}

	return false
}

// Recursive traversal code based on code from https://gist.github.com/hvoecking/10772475
func sanitiseRecursive(sanitised, original reflect.Value, context string) {
	switch original.Kind() {
	// The first cases handle nested structures and sanitise them recursively

	case reflect.Ptr:
		// If it is a pointer we need to unwrap and call once again

		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalValue := original.Elem()
		// Check if the pointer is nil
		if !originalValue.IsValid() {

			return
		}
		// Allocate a new object and set the pointer to it
		sanitised.Set(reflect.New(originalValue.Type()))
		// Unwrap the newly created pointer
		sanitiseRecursive(sanitised.Elem(), originalValue, context)

	case reflect.Interface:
		// If it is an interface (which is very similar to a pointer), do basically the
		// same as for the pointer. Though a pointer is not the same as an interface so
		// note that we have to call Elem() after creating a new object because otherwise
		// we would end up with an actual pointer

		// Get rid of the wrapping interface
		originalValue := original.Elem()
		// Create a new object. Now new gives us a pointer, but we want the value it
		// points to, so we have to call Elem() to unwrap it
		sanitisedValue := reflect.New(originalValue.Type()).Elem()
		sanitiseRecursive(sanitisedValue, originalValue, context)
		sanitised.Set(sanitisedValue)

	case reflect.Struct:
		// If it is a struct we sanitise each field
		typ := reflect.TypeOf(original.Interface())
		for i := 0; i < original.NumField(); i += 1 {

			dbg("Processing field %v\n", typ.Field(i).Name)
			if shouldSanitiseField(original.Field(i), typ.Field(i), context) {
				// sanitise this field
				dbg("-> Sanitising field %v\n", typ.Field(i).Name)
			} else {

				sanitiseRecursive(sanitised.Field(i), original.Field(i), context)
			}
		}

	case reflect.Slice:
		// If it is a slice we create a new slice and sanitise each element
		sanitised.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i += 1 {

			sanitiseRecursive(sanitised.Index(i), original.Index(i), context)
		}

	case reflect.Map:
		// If it is a map we create a new map and sanitise each value
		sanitised.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {

			originalValue := original.MapIndex(key)
			// New gives us a pointer, but again we want the value
			sanitisedValue := reflect.New(originalValue.Type()).Elem()
			sanitiseRecursive(sanitisedValue, originalValue, context)

			sanitisedKey := reflect.New(key.Type()).Elem()
			sanitiseRecursive(sanitisedKey, key, context)
			sanitised.SetMapIndex(sanitisedKey, sanitisedValue)
		}

	default:
		// And everything else will simply be taken from the original
		if sanitised.CanSet() {

			sanitised.Set(original)
		}
	}

	if sanitised.CanInterface() {

		if s, ok := sanitised.Interface().(Sanitiser); ok {

			dbg("Object of type %T supports the Sanitiser interface, invoking Sanitise()\n", sanitised.Interface())
			s.Sanitise(context)
		} else if sanitised.CanAddr() {

			if s, ok := sanitised.Addr().Interface().(Sanitiser); ok {

				dbg("Object of type %T supports the Sanitiser interface, invoking Sanitise()\n", sanitised.Addr().Interface())
				s.Sanitise(context)
			}
		} else {

			dbg("Object of type %T does not supports the Sanitiser interface\n", sanitised.Interface())
		}
	} else {

		dbg("Not attempting to invoke Sanitiser interface on an inaccessible object\n")
	}
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
