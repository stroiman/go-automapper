// Copyright (c) 2015 Peter Strøiman, distributed under the MIT license

// Package automapper provides support for mapping between two different types
// with compatible fields. The intended application for this is when you use
// one set of types to represent DTOs (data transfer objects, e.g. json data),
// and a different set of types internally in the application. Using this
// package can help converting from one type to another.
//
// This package uses reflection to perform mapping which should be fine for
// all but the most demanding applications.
package automapper

import (
	"fmt"
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
// destination does not result in subtle silent bug.
func Map(source, dest interface{}) {
	Mapper{}.Map(source, dest)
}

// Map maps from src to dst.
//
// If mapping is not possible because of incompatible types, Map panics. But if
// mapping is not possible because of incompatible values, e.g. string-to-int
// conversion fails, it returns an error.
func (m Mapper) Map(src, dst interface{}) error {
	var destType = reflect.TypeOf(dst)
	if destType.Kind() != reflect.Ptr {
		panic("Dest must be a pointer type")
	}
	var sourceVal = reflect.ValueOf(src)
	var destVal = reflect.ValueOf(dst).Elem()
	return m.mapValues(sourceVal, destVal)
}

// MapLoose works just like Map, except it doesn't fail when the destination
// type contains fields not supplied by the source.
//
// This function is meant to be a temporary solution - the general idea is
// that the Map function should take a number of options that can modify its
// behavior - but I'd rather not add that functionality before I have a better
// idea what is a good options format.
func MapLoose(source, dest interface{}) {
	var destType = reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		panic("Dest must be a pointer type")
	}
	var sourceVal = reflect.ValueOf(source)
	var destVal = reflect.ValueOf(dest).Elem()
	Mapper{loose: true}.mapValues(sourceVal, destVal)
}

type typeMapKey struct {
	srcType reflect.Type
	dstType reflect.Type
}

type TypeMapper = func(src, dst interface{}) error

type innerTypeMapper = func(src, dst reflect.Value) error

type Mapper struct {
	loose       bool
	typeMappers map[typeMapKey]innerTypeMapper
}

// Registers a type mapper working on reflect.Value instances.
//
// This could potentially be exported, but I prefer to keep the interface
// minimal to avoid future breaking changes, and I see no need to export this.
//
// The exported Register function is easier to use from client code, as you only
// need to express yourself in Go-syntax, and not learn the details of the reflect
// API.
func (m *Mapper) register(srcType, dstType reflect.Type, fn innerTypeMapper) {
	if m.typeMappers == nil {
		m.typeMappers = make(map[typeMapKey]innerTypeMapper)
	}
	key := typeMapKey{srcType, dstType}
	if _, exists := m.typeMappers[key]; exists {
		panic(fmt.Sprintf("automapper: a converter already exists from %v to %v", srcType, dstType))
	}
	m.typeMappers[key] = fn
}

// Register adds a custom function for converting between two types, e.g.
// between a string and int. The conversion function fn receives two arguments,
// src and dst. Src will contain the source value, and dest will contain a
// pointer to the destination value.
//
// # Example - convert from string to integers
//
//		var m automapper.Mapper
//		m.Register(
//			reflect.TypeFor[string](), reflect.TypeFor[int](),
//			func(src, dst any) error {
//	 		s := src.(string)
//				d := dst.(*int) // <- pointer value
//				(*d), err := strconv.Atoi(s)
//				return err
//			})
//
// Note: The library works with go 1.17, the example just use newer go-syntax for simplicity.
func (m *Mapper) Register(srcType, dstType reflect.Type, fn TypeMapper) {
	m.register(srcType, dstType, func(src, dst reflect.Value) error {
		dstPtr := reflect.New(dstType)
		err := fn(src.Interface(), dstPtr.Interface())
		if err == nil {
			dst.Set(dstPtr.Elem())
		}
		return err
	})
}

func (m Mapper) mapValues(sourceVal, destVal reflect.Value) error {
	destType := destVal.Type()
	srcType := sourceVal.Type()

	if m, ok := m.typeMappers[typeMapKey{srcType, destType}]; ok {
		return m(sourceVal, destVal)
	}
	if destType.Kind() == reflect.Struct {
		if sourceVal.Type().Kind() == reflect.Ptr {
			if sourceVal.IsNil() {
				// If source is nil, it maps to an empty struct
				sourceVal = reflect.New(sourceVal.Type().Elem())
			}
			sourceVal = sourceVal.Elem()
		}
		var errs []FieldError
		for i := 0; i < destVal.NumField(); i++ {
			field := destType.Field(i)
			if err := m.mapField(sourceVal, destVal, i); err != nil {
				errs = append(errs, FieldError{Field: field.Name, Err: err})
			}
		}
		if len(errs) > 0 {
			return StructError{errs}
		}
		return nil
	} else if destType == sourceVal.Type() {
		destVal.Set(sourceVal)
		return nil
	} else if destType.Kind() == reflect.Ptr {
		if valueIsNil(sourceVal) {
			return nil
		}
		val := reflect.New(destType.Elem())
		err := m.mapValues(sourceVal, val.Elem())
		destVal.Set(val)
		return err
	} else if destType.Kind() == reflect.Slice {
		return m.mapSlice(sourceVal, destVal)
	} else {
		panic("Currently not supported")
	}
}

func (m Mapper) mapSlice(sourceVal, destVal reflect.Value) error {
	destType := destVal.Type()
	length := sourceVal.Len()
	target := reflect.MakeSlice(destType, length, length)
	errs := make([]error, 0, length)
	for j := 0; j < length; j++ {
		val := reflect.New(destType.Elem()).Elem()
		if err := m.mapValues(sourceVal.Index(j), val); err != nil {
			errs = append(errs, err)
		}
		target.Index(j).Set(val)
	}

	if length == 0 {
		m.verifyArrayTypesAreCompatible(sourceVal, destVal)
	}
	destVal.Set(target)
	if len(errs) > 0 {
		return SliceError{errs: errs}
	}
	return nil
}

func (m Mapper) verifyArrayTypesAreCompatible(sourceVal, destVal reflect.Value) {
	dummyDest := reflect.New(reflect.PtrTo(destVal.Type()))
	dummySource := reflect.MakeSlice(sourceVal.Type(), 1, 1)
	m.mapValues(dummySource, dummyDest.Elem())
}

func (m Mapper) mapField(srcStruct, dstStruct reflect.Value, i int) error {
	destType := dstStruct.Type()
	dstField := destType.Field(i)
	fieldName := dstField.Name
	defer func() {
		if r := recover(); r != nil {
			panic(
				fmt.Sprintf(
					"Error mapping field: %s. DestType: %v. SourceType: %v. Error: %v",
					fieldName,
					destType,
					srcStruct.Type(),
					r,
				),
			)
		}
	}()

	dstValue := dstStruct.Field(i)
	if destType.Field(i).Anonymous {
		return m.mapValues(srcStruct, dstValue)
	} else {
		if valueIsContainedInNilEmbeddedType(srcStruct, fieldName) {
			return nil
		}
		sourceField := srcStruct.FieldByName(fieldName)
		if (sourceField == reflect.Value{}) {
			if m.loose {
				return nil
			}
			if dstValue.Kind() == reflect.Struct {
				return m.mapValues(srcStruct, dstValue)
			} else {
				for i := 0; i < srcStruct.NumField(); i++ {
					if srcStruct.Field(i).Kind() != reflect.Struct {
						continue
					}
					if sourceField = srcStruct.Field(i).
						FieldByName(fieldName); (sourceField != reflect.Value{}) {
						break
					}
				}
			}
		}
		return m.mapValues(sourceField, dstValue)
	}
}

func valueIsNil(value reflect.Value) bool {
	return value.Type().Kind() == reflect.Ptr && value.IsNil()
}

func valueIsContainedInNilEmbeddedType(source reflect.Value, fieldName string) bool {
	structField, _ := source.Type().FieldByName(fieldName)
	ix := structField.Index
	if len(structField.Index) > 1 {
		parentField := source.FieldByIndex(ix[:len(ix)-1])
		if valueIsNil(parentField) {
			return true
		}
	}
	return false
}
