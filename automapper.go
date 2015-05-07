// Package automapper provides support for mapping between two different types
// with compatible fields
package automapper

import "errors"
import "reflect"

func Map(source, dest interface{}) error {
	var destType = reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		return errors.New("Dest must be a pointer type")
	}
	var sourceVal = reflect.ValueOf(source)
	var destVal = reflect.ValueOf(dest).Elem()
	destVal.Field(0).Set(sourceVal.Field(0))
	return nil
}

func MustMap(source, dest interface{}) {
	if err := Map(source, dest); err != nil {
		panic(err)
	}
}
