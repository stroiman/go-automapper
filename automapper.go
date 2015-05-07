// Package automapper provides support for mapping between two different types
// with compatible fields
package automapper

import "reflect"

func Map(source, dest interface{}) {
	var destType = reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		panic("Dest must be a pointer type")
	}
	var sourceVal = reflect.ValueOf(source)
	if sourceVal.Type().Kind() == reflect.Ptr {
		sourceVal = sourceVal.Elem()
	}
	destType = destType.Elem()
	var destVal = reflect.ValueOf(dest).Elem()
	for i := 0; i < destVal.NumField(); i++ {
		fieldName := destType.Field(i).Name
		value := sourceVal.FieldByName(fieldName)
		destVal.Field(i).Set(value)
	}
}
