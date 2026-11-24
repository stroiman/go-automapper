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
	field string
	err   error
}

// FieldName returns the name of the destination struct field mapping to
func (e FieldError) FieldName() string { return e.field }

// Err returns the original error generated when mapping the struct field
func (e FieldError) Err() error { return e.err }

// Error implements interface error
func (e FieldError) Error() string { return fmt.Sprintf("%s: %v", e.field, e.err) }

func (e FieldError) Unwrap() error { return e.err }

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

func (e StructError) Unwrap() []error {
	errs := make([]error, len(e.Fields))
	for i, f := range e.Fields {
		errs[i] = f
	}
	return errs
}

type SliceElementError struct {
	Index int
	Err   error
}

func (e SliceElementError) Error() string { return fmt.Sprintf("[%d]: %v", e.Index, e.Err) }
func (e SliceElementError) Unwrap() error { return e.Err }

type SliceError struct {
	Errs []SliceElementError
}

func (e SliceError) Error() string {
	strs := make([]string, len(e.Errs))
	for i, f := range e.Errs {
		strs[i] = f.Error()
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
}

func (e SliceError) Unwrap() []error {
	errs := make([]error, len(e.Errs))
	for i, f := range e.Errs {
		errs[i] = f
	}
	return errs
}
