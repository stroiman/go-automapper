// Package automapper provides support for mapping between two different types
// with compatible fields
package automapper

import "errors"
import "reflect"

func Map(source, dest interface{}) error {
	var destType = reflect.TypeOf(dest)
	if destType.Kind() == reflect.Ptr {
		return nil
	}
	return errors.New("Dest must be a pointer type")
}
