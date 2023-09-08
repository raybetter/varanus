package config

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Error  string
	Object interface{}
}

type ValidationProcess struct {
	MetaError error
	Errors    []ValidationError
}

func (vp *ValidationProcess) addValidationError(object interface{}, message string, messageArgs ...interface{}) {
	vp.Errors = append(vp.Errors, ValidationError{
		Error:  fmt.Sprintf(message, messageArgs...),
		Object: object,
	})
}

func (vp *ValidationProcess) Validate(target Validatable) bool {
	new_errors, err := target.Validate(vp)
	if err != nil {
		vp.MetaError = err
		return false
	}
	vp.Errors = append(vp.Errors, new_errors...)
	return true
}

func (vp *ValidationProcess) Finalize() error {
	if vp.MetaError != nil {
		return fmt.Errorf("failed to complete validation: %w", err)
	}
	if len(validationErrors) > 0 {
		var sb strings.Builder
		for _, validationError := range validationErrors {
			sb.Write([]byte(fmt.Sprintf("Error: %s on object %#v", validationError.Error, validationError.Object)))
		}
		return nil, fmt.Errorf("%s", sb.String())
	}

}

type Validatable interface {
	Validate(vp *ValidationProcess) ([]ValidationError, error)
}
