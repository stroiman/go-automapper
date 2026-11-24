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

func (m Mapper) Map(source, dest interface{}) {
	var destType = reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		panic("Dest must be a pointer type")
	}
	var sourceVal = reflect.ValueOf(source)
	var destVal = reflect.ValueOf(dest).Elem()
	m.mapValues(sourceVal, destVal)
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

type typeMapper = func(src, dst interface{}) error

type Mapper struct {
	loose       bool
	typeMappers map[typeMapKey]typeMapper
}

func (m *Mapper) Register(srcType, dstType reflect.Type, fn typeMapper) {
	if m.typeMappers == nil {
		m.typeMappers = make(map[typeMapKey]typeMapper)
	}
	key := typeMapKey{srcType, dstType}
	if _, exists := m.typeMappers[key]; exists {
		panic(fmt.Sprintf("automapper: a converter already exists from %v to %v", srcType, dstType))
	}
	m.typeMappers[key] = fn
}

func (m Mapper) mapValues(sourceVal, destVal reflect.Value) {
	destType := destVal.Type()
	if destType.Kind() == reflect.Struct {
		if sourceVal.Type().Kind() == reflect.Ptr {
			if sourceVal.IsNil() {
				// If source is nil, it maps to an empty struct
				sourceVal = reflect.New(sourceVal.Type().Elem())
			}
			sourceVal = sourceVal.Elem()
		}
		for i := 0; i < destVal.NumField(); i++ {
			m.mapField(sourceVal, destVal, i)
		}
	} else if destType == sourceVal.Type() {
		destVal.Set(sourceVal)
	} else if destType.Kind() == reflect.Ptr {
		if valueIsNil(sourceVal) {
			return
		}
		val := reflect.New(destType.Elem())
		m.mapValues(sourceVal, val.Elem())
		destVal.Set(val)
	} else if destType.Kind() == reflect.Slice {
		m.mapSlice(sourceVal, destVal)
	} else {
		srcType := sourceVal.Type()
		fmt.Printf("Src: %v\nDst: %v\n", srcType, destType)
		if m, ok := m.typeMappers[typeMapKey{srcType, destType}]; ok {
			dstPtr := reflect.New(destType)
			err := m(sourceVal.Interface(), dstPtr.Interface())
			if err != nil {
				panic("Conversion error")
			}
			destVal.Set(dstPtr.Elem())
		} else {
			panic("Currently not supported")
		}
	}
}

func (m Mapper) mapSlice(sourceVal, destVal reflect.Value) {
	destType := destVal.Type()
	length := sourceVal.Len()
	target := reflect.MakeSlice(destType, length, length)
	for j := 0; j < length; j++ {
		val := reflect.New(destType.Elem()).Elem()
		m.mapValues(sourceVal.Index(j), val)
		target.Index(j).Set(val)
	}

	if length == 0 {
		m.verifyArrayTypesAreCompatible(sourceVal, destVal)
	}
	destVal.Set(target)
}

func (m Mapper) verifyArrayTypesAreCompatible(sourceVal, destVal reflect.Value) {
	dummyDest := reflect.New(reflect.PtrTo(destVal.Type()))
	dummySource := reflect.MakeSlice(sourceVal.Type(), 1, 1)
	m.mapValues(dummySource, dummyDest.Elem())
}

func (m Mapper) mapField(source, destVal reflect.Value, i int) {
	destType := destVal.Type()
	fieldName := destType.Field(i).Name
	defer func() {
		if r := recover(); r != nil {
			panic(
				fmt.Sprintf(
					"Error mapping field: %s. DestType: %v. SourceType: %v. Error: %v",
					fieldName,
					destType,
					source.Type(),
					r,
				),
			)
		}
	}()

	destField := destVal.Field(i)
	if destType.Field(i).Anonymous {
		m.mapValues(source, destField)
	} else {
		if valueIsContainedInNilEmbeddedType(source, fieldName) {
			return
		}
		sourceField := source.FieldByName(fieldName)
		if (sourceField == reflect.Value{}) {
			if m.loose {
				return
			}
			if destField.Kind() == reflect.Struct {
				m.mapValues(source, destField)
				return
			} else {
				for i := 0; i < source.NumField(); i++ {
					if source.Field(i).Kind() != reflect.Struct {
						continue
					}
					if sourceField = source.Field(i).
						FieldByName(fieldName); (sourceField != reflect.Value{}) {
						break
					}
				}
			}
		}
		m.mapValues(sourceField, destField)
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
