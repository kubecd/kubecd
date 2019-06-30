package model

import "fmt"

type AggregateError struct {
	Errors []error
}

func NewAggregateError(errors []error) *AggregateError {
	return &AggregateError{Errors: errors}
}

func (a AggregateError) Error() string {
	return fmt.Sprintf("[%v]", a.Errors)
}

var _ error = AggregateError{}
