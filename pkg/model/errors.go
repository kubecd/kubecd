package model

import (
	"strings"
)

type AggregateError struct {
	Errors []error
}

func NewAggregateError(errors []error) *AggregateError {
	return &AggregateError{Errors: errors}
}

func (a AggregateError) Error() string {
	strs := make([]string, len(a.Errors))
	for i := range a.Errors {
		strs[i] = a.Errors[i].Error()
	}
	return strings.Join(strs, "\n\t")
}

var _ error = AggregateError{}
