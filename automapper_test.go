package automapper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPanicWhenDestIsNotPointer(t *testing.T) {
	defer func() { recover() }()
	source, dest := SourceTypeA{}, DestTypeA{}
	Map(source, dest)

	t.Error("Should have panicked")
}

func TestDestinationIsUpdatedFromSource(t *testing.T) {
	source, dest := SourceTypeA{Foo: 42}, DestTypeA{}
	Map(source, &dest)
	assert.Equal(t, 42, dest.Foo)
}

func TestDestinationIsUpdatedFromSourceWhenSourcePassedAsPtr(t *testing.T) {
	source, dest := SourceTypeA{42, "Bar"}, DestTypeA{}
	Map(&source, &dest)
	assert.Equal(t, 42, dest.Foo)
	assert.Equal(t, "Bar", dest.Bar)
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
	assert.Equal(t, "Baz", dest.Baz)
	assert.Equal(t, "Bar", dest.Child.Bar)
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
	assert.Equal(t, 1, dest.Children[0].Foo)
	assert.Equal(t, 2, dest.Children[1].Foo)
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
	assert.Equal(t, "Baz", dest.Baz)
	assert.Equal(t, 42, dest.DestTypeA.Foo)
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
	assert.Nil(t, dest.Foo)
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
	assert.NotNil(t, dest.Foo)
	assert.Equal(t, 42, dest.Foo.Foo)
}

func TestWhenUsingIncompatibleTypes(t *testing.T) {
	defer func() { recover() }()
	source := struct{ Foo string }{}
	dest := struct{ Foo int }{}
	Map(&source, &dest)
	t.Error("Should have panicked")
}

type SourceTypeA struct {
	Foo int
	Bar string
}

type DestTypeA struct {
	Foo int
	Bar string
}
