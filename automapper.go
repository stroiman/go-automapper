// Package automapper provides support for mapping between two different types
// with compatible fields
package automapper

import (
	"fmt"
	"reflect"
)

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
		fmt.Printf("Field %s. Type %s, IsArray: %v\n", fieldName, destFieldType, destFieldType.Kind() == reflect.Slice)
		if destFieldType == sourceField.Type() {
			destField.Set(sourceField)
		} else if destFieldType.Kind() == reflect.Slice {
			length := sourceField.Len()
			arrayElmType := destFieldType.Elem()
			fmt.Printf("Array elm type: %v\n", arrayElmType)
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
