package automapper

import (
	"fmt"
	"strings"
)

type MapError struct {
	path string
	err  error
}

type FieldError struct {
	Field string
	Err   error
}

func (e FieldError) Error() string { return fmt.Sprintf("%s: %v", e.Field, e.Err) }

type StructError struct {
	Fields []FieldError
}

func (e StructError) Error() string {
	strs := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		strs[i] = f.Error()
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
}

type SliceError struct {
	errs []error
}

func (e SliceError) Error() string {
	return "error"
}
