// Copyright (c) 2015 Peter Strøiman, distributed under the MIT license

package automapper

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func TestPanicWhenDestIsNotPointer(t *testing.T) {
	defer func() { recover() }()
	source, dest := SourceTypeA{}, DestTypeA{}
	Map(source, dest)

	t.Error("Should have panicked")
}

func isNil(got interface{}) bool {
	if got == nil {
		return true
	}
	v := reflect.ValueOf(got)
	if v.Type().Kind() == reflect.Ptr && v.IsNil() {
		return true
	}
	return false
}

func assertNil(t testing.TB, got interface{}) {
	t.Helper()
	if isNil(got) {
		return
	}
	t.Errorf("Expected nil value - got: %v", got)
}

func assertNotNil(t testing.TB, got interface{}) {
	t.Helper()
	if !isNil(got) {
		return
	}
	t.Errorf("Expected non-nil value")
}

func assertError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		return
	}
	t.Error("Expected an error, received none")
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		return
	}
	t.Errorf("Unexpected error: %v", err)
}

func assertEqual(t testing.TB, got, want interface{}) {
	t.Helper()

	if reflect.DeepEqual(got, want) {
		return
	}
	t.Errorf("Expected: %v - got: %v", want, got)
}

func TestDestinationIsUpdatedFromSource(t *testing.T) {
	source, dest := SourceTypeA{Foo: 42}, DestTypeA{}
	Map(source, &dest)
	assertEqual(t, dest.Foo, 42)
}

func TestDestinationIsUpdatedFromSourceWhenSourcePassedAsPtr(t *testing.T) {
	source, dest := SourceTypeA{42, "Bar"}, DestTypeA{}
	Map(&source, &dest)
	assertEqual(t, dest.Foo, 42)
	assertEqual(t, dest.Bar, "Bar")
}

func TestWithNestedTypes(t *testing.T) {
	source := struct {
		Baz   string
		Child SourceTypeA
	}{}
	dest := struct {
		Baz   string
		Child DestTypeA
	}{}

	source.Baz = "Baz"
	source.Child.Bar = "Bar"
	Map(&source, &dest)
	assertEqual(t, dest.Baz, "Baz")
	assertEqual(t, dest.Child.Bar, "Bar")
}

func TestWithSourceSecondLevel(t *testing.T) {
	source := struct {
		Child DestTypeA
	}{}
	dest := SourceTypeA{}

	source.Child.Bar = "Bar"
	Map(&source, &dest)
	assertEqual(t, dest.Bar, "Bar")
}

func TestWithDestSecondLevel(t *testing.T) {
	source := SourceTypeA{}
	dest := struct {
		Child DestTypeA
	}{}

	source.Bar = "Bar"
	Map(&source, &dest)
	assertEqual(t, dest.Child.Bar, "Bar")
}

func TestWithSliceTypes(t *testing.T) {
	source := struct {
		Children []SourceTypeA
	}{}
	dest := struct {
		Children []DestTypeA
	}{}
	source.Children = []SourceTypeA{
		SourceTypeA{Foo: 1},
		SourceTypeA{Foo: 2}}

	Map(&source, &dest)
	assertEqual(t, dest.Children[0].Foo, 1)
	assertEqual(t, dest.Children[1].Foo, 2)
}

func TestWithMultiLevelSlices(t *testing.T) {
	source := struct {
		Parents []SourceParent
	}{}
	dest := struct {
		Parents []DestParent
	}{}
	source.Parents = []SourceParent{
		SourceParent{
			Children: []SourceTypeA{
				SourceTypeA{Foo: 42},
				SourceTypeA{Foo: 43},
			},
		},
		SourceParent{
			Children: []SourceTypeA{},
		},
	}

	Map(&source, &dest)
}

func TestWithEmptySliceAndIncompatibleTypes(t *testing.T) {
	defer func() { recover() }()

	source := struct {
		Children []struct{ Foo string }
	}{}
	dest := struct {
		Children []struct{ Bar int }
	}{}

	Map(&source, &dest)
	t.Error("Should have panicked")
}

func TestWhenSourceIsMissingField(t *testing.T) {
	defer func() { recover() }()
	source := struct {
		A string
	}{}
	dest := struct {
		A, B string
	}{}
	Map(&source, &dest)
	t.Error("Should have panicked")
}

func TestWithUnnamedFields(t *testing.T) {
	source := struct {
		Baz string
		SourceTypeA
	}{}
	dest := struct {
		Baz string
		DestTypeA
	}{}
	source.Baz = "Baz"
	source.SourceTypeA.Foo = 42

	Map(&source, &dest)

	assertEqual(t, dest.Baz, "Baz")
	assertEqual(t, dest.DestTypeA.Foo, 42)
}

func TestWithPointerFieldsNotNil(t *testing.T) {
	source := struct {
		Foo *SourceTypeA
	}{}
	dest := struct {
		Foo *DestTypeA
	}{}
	source.Foo = nil

	Map(&source, &dest)
	assertNil(t, dest.Foo)
}

func TestWithPointerFieldsNil(t *testing.T) {
	source := struct {
		Foo *SourceTypeA
	}{}
	dest := struct {
		Foo *DestTypeA
	}{}
	source.Foo = &SourceTypeA{Foo: 42}

	Map(&source, &dest)
	assertNotNil(t, dest.Foo)
	assertEqual(t, dest.Foo.Foo, 42)
}

func TestMapFromPointerToNonPointerTypeWithData(t *testing.T) {
	source := struct {
		Foo *SourceTypeA
	}{}
	dest := struct {
		Foo DestTypeA
	}{}
	source.Foo = &SourceTypeA{Foo: 42}

	Map(&source, &dest)
	assertNotNil(t, dest.Foo)
	assertEqual(t, dest.Foo.Foo, 42)
}

func TestMapFromPointerToNonPointerTypeWithoutData(t *testing.T) {
	source := struct {
		Foo *SourceTypeA
	}{}
	dest := struct {
		Foo DestTypeA
	}{}
	source.Foo = nil

	Map(&source, &dest)
	assertNotNil(t, dest.Foo)
	assertEqual(t, dest.Foo.Foo, 0)
}

func TestMapFromPointerToAnonymousTypeToFieldName(t *testing.T) {
	source := struct {
		*SourceTypeA
	}{}
	dest := struct {
		Foo int
	}{}
	source.SourceTypeA = nil

	Map(&source, &dest)
	assertEqual(t, dest.Foo, 0)
}

func TestMapFromPointerToNonPointerTypeWithoutDataAndIncompatibleType(t *testing.T) {
	defer func() { recover() }()
	// Just make sure we stil panic
	source := struct {
		Foo *SourceTypeA
	}{}
	dest := struct {
		Foo struct {
			Baz string
		}
	}{}
	source.Foo = nil

	Map(&source, &dest)
	t.Error("Should have panicked")
}

func TestWhenUsingIncompatibleTypes(t *testing.T) {
	defer func() { recover() }()
	source := struct{ Foo string }{}
	dest := struct{ Foo int }{}
	Map(&source, &dest)
	t.Error("Should have panicked")
}

func TestWithLooseOption(t *testing.T) {
	source := struct {
		Foo string
		Baz int
	}{"Foo", 42}
	dest := struct {
		Foo string
		Bar int
	}{}
	MapLoose(&source, &dest)
	assertEqual(t, dest.Foo, "Foo")
	assertEqual(t, dest.Bar, 0)
}

func TestConvertFromStringToInt(t *testing.T) {
	source := struct {
		Foo string
	}{"42"}
	dest := struct {
		Foo int
	}{}
	m := Mapper{}
	m.Register(
		reflect.TypeOf(dest.Foo),
		reflect.TypeOf(source.Foo),
		func(dst, src interface{}) error {
			a := src.(string)
			b := dst.(*int)
			var err error
			*b, err = strconv.Atoi(a)
			return err
		},
	)
	m.Map(&dest, source)
	assertEqual(t, dest.Foo, 42)
}

func TestConvertFromStringToIntWithReflectValues(t *testing.T) {
	// This test touches the unexported register, which is currently just an
	// implementation detail; but could potentially be exported.
	source := struct {
		Foo string
	}{"42"}
	dest := struct {
		Foo int
	}{}
	m := Mapper{}
	typeStr := reflect.TypeOf(source.Foo)
	typeInt := reflect.TypeOf(dest.Foo)
	m.register(
		typeInt,
		typeStr,
		func(dst, src reflect.Value) error {
			str := src.Interface().(string)
			i, err := strconv.Atoi(str)
			if err != nil {
				return err
			}
			dst.Set(reflect.ValueOf(i))
			return nil
		},
	)
	m.Map(&dest, source)
	assertEqual(t, dest.Foo, 42)
}

func TestExplicitConversionTakesPriority(t *testing.T) {
	// If client code specifies explicit conversion between two struct types,
	// let that take priority.
	source := struct {
		Foo string
	}{"42"}
	dest := struct {
		Foo int
	}{}
	m := Mapper{}
	// If this is used, the result would be 42
	m.Register(
		reflect.TypeOf(dest.Foo),
		reflect.TypeOf(source.Foo),
		func(dst, src interface{}) error {
			a := src.(string)
			b := dst.(*int)
			var err error
			*b, err = strconv.Atoi(a)
			return err
		},
	)
	// If this is used, the result would be 43
	m.Register(
		reflect.TypeOf(dest),
		reflect.TypeOf(source),
		func(dst, src interface{}) error {
			a := src.(struct{ Foo string })
			b := dst.(*struct{ Foo int })
			var err error
			(*b).Foo, err = strconv.Atoi(a.Foo)
			b.Foo++
			return err
		},
	)
	m.Map(&dest, source)
	assertEqual(t, dest.Foo, 43)
}

func TestConversionError(t *testing.T) {
	source := struct {
		Foo string
	}{"not-an-int"}
	dest := struct {
		Foo int
	}{}
	m := Mapper{}
	// If this is used, the result would be 43
	m.Register(
		reflect.TypeOf(dest.Foo),
		reflect.TypeOf(source.Foo),
		func(dst, src interface{}) error {
			return ErrValidation
		},
	)
	err := m.Map(&dest, source)
	assertError(t, err)
	fmt.Printf("T: %T", err)
	assertEqual(t, err, StructError{Fields: []FieldError{{field: "Foo", err: ErrValidation}}})
	// assertNoError(t, )
	assertEqual(t, dest.Foo, 0)
}

var ErrValidation = errors.New("validation")

type SourceParent struct {
	Children []SourceTypeA
}

type DestParent struct {
	Children []DestTypeA
}

type SourceTypeA struct {
	Foo int
	Bar string
}

type DestTypeA struct {
	Foo int
	Bar string
}
