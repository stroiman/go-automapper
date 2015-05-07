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
	var destVal = reflect.ValueOf(dest).Elem()
	mapValues(sourceVal, destVal)
}

func mapValues(sourceVal, destVal reflect.Value) {
	destType := destVal.Type()
	for i := 0; i < destVal.NumField(); i++ {
		fieldName := destType.Field(i).Name
		sourceField := sourceVal.FieldByName(fieldName)
		destField := destVal.Field(i)
		if destField.Type() == sourceField.Type() {
			destField.Set(sourceField)
		} else {
			mapValues(sourceField, destField)
		}
	}
}
