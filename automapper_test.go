package automapper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDestinationPassedAsPointer(t *testing.T) {
	source, dest := SourceTypeA{}, DestTypeA{}
	err := Map(source, &dest)
	assert.Nil(t, err)
}

func TestDestinationNotPassedAsPointer(t *testing.T) {
	source, dest := SourceTypeA{}, DestTypeA{}
	err := Map(source, dest)
	assert.NotNil(t, err)
}

func TestDestinationIsUpdatedFromSource(t *testing.T) {
	source, dest := SourceTypeA{42}, DestTypeA{}
	MustMap(source, &dest)
	assert.Equal(t, 42, dest.Foo)
}

type SourceTypeA struct {
	Foo int
}

type DestTypeA struct {
	Foo int
}
