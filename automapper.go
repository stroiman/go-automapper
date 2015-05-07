// Package automapper provides support for mapping between two different types
// with compatible fields. The intended application for this is when you use
// one set of types to represent DTOs (data transfer objects, e.g. json data),
// and a different set of types internally in the application. Using this
// package can help converting from one type to another.
package automapper

import (
	"reflect"
)

// Map fills out the fields in dest with values from source. All fields in the
// destination object must exist in the source object.
//
// Object hierarchies with nested structs and slices are supported, as long as
// type types of nested structs/slices follow the same rules, i.e. all fields
// in destination structs must be found on the source struct.
//
// Embedded/anonymous structs are supported
//
// Values that are not exported/not public will not be mapped.
//
// It is a design decision to panic when a field cannot be mapped in the
// destination to ensure that a renamed field in either the source or
// destination does not result in subtle silent buge.
//
// BUG(ps) - It is the intend that the code should panic early when mapping
// incompatible types. Empty slices are not supported however
func Map(source, dest interface{}) {
	var destType = reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		panic("Dest must be a pointer type")
	}
	var sourceVal = reflect.ValueOf(source)
	if sourceVal.Type().Kind() == reflect.Ptr {
		sourceVal = sourceVal.Elem()
	}
	var destVal = reflect.ValueOf(dest).Elem()
	mapValues(sourceVal, destVal)
}

func mapValues(sourceVal, destVal reflect.Value) {
	destType := destVal.Type()
	for i := 0; i < destVal.NumField(); i++ {
		fieldName := destType.Field(i).Name
		sourceField := sourceVal.FieldByName(fieldName)
		destField := destVal.Field(i)
		if destType.Field(i).Anonymous {
			mapValues(sourceVal, destField)
			continue
		}
		destFieldType := destField.Type()
		if destFieldType == sourceField.Type() {
			destField.Set(sourceField)
		} else if destFieldType.Kind() == reflect.Slice {
			length := sourceField.Len()
			target := reflect.MakeSlice(destFieldType, length, length)
			for j := 0; j < length; j++ {
				val := reflect.New(destFieldType.Elem()).Elem()
				mapValues(sourceField.Index(j), val)
				target.Index(j).Set(val)
			}
			destField.Set(target)
		} else {
			mapValues(sourceField, destField)
		}
	}
}
