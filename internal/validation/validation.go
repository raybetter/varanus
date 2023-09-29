package validation

import (
	"fmt"
	"reflect"
	"strings"
	"varanus/internal/walker"

	"github.com/kr/pretty"
)

// SingleValidationError captures the context for a single validation error discovered during
// validation.
//
// It is the primary type created by the AddValidationError of ValidationResult
type SingleValidationError struct {
	// Error describes the validation error
	Error string
	// Object provides the context of the object where the error occured.
	Object interface{}
}

// ValidationResult is a collection of SingleValidationErrors accumulated during a call to
// ValidateObject.
type ValidationResult struct {
	errorList []SingleValidationError
}

// GetErrorList provides a copy of validation errors it holds.
func (vr ValidationResult) GetErrorList() []SingleValidationError {
	listCopy := make([]SingleValidationError, len(vr.errorList))
	copy(listCopy, vr.errorList)
	return listCopy
}

// HumanReadable returns a human-readable formatted list of the validation errors suitable for
// printing to the terminal.
func (vr ValidationResult) HumanReadable() string {
	var sb strings.Builder

	if len(vr.errorList) > 0 {

		sb.WriteString("*****************************************************\n")
		sb.WriteString(fmt.Sprintf("%d Validation Errors\n", len(vr.errorList)))

		for _, validationError := range vr.errorList {
			sb.WriteString(fmt.Sprintf("Error %s on object:\n%# v\n",
				validationError.Error, pretty.Formatter(validationError.Object)))
			sb.WriteString("--------------------\n")
		}

		sb.WriteString("*****************************************************\n")

	} else {
		sb.WriteString("No validation errors\n")
	}
	return sb.String()
}

// GetErrorCount returns the number of validation errors in the result
func (vr ValidationResult) GetErrorCount() int {
	return len(vr.errorList)
}

// AsError returns an error object containing all validation errors.  Returns nil if there are
// no validation errors
//
// The error value returned here is not very user-friendly.  Consider using GetErrorList() or
// HumanReadable if users are meant to parse the error list.
func (vr ValidationResult) AsError() error {
	if len(vr.errorList) == 0 {
		return nil
	}

	var sb strings.Builder
	for _, validationError := range vr.errorList {
		sb.Write([]byte(fmt.Sprintf("Error: %s on object %#v\n", validationError.Error, validationError.Object)))
	}
	return fmt.Errorf(sb.String())
}

// AddValidationError should be called by the Validatable objects in their Validate() impelementation
// to add a validation error.
// - object is the object the validation error occurs on (provided for context)
// - message is a format string describing the error
// - any remaining parameters will be supplied when parsing the message format string
//
// Panics if called after a validation failure (see Validate() function description).
//
// Implmentsf ValidationErrorTracker
func (vr *ValidationResult) AddValidationError(object interface{}, message string, messageArgs ...interface{}) {
	vr.errorList = append(vr.errorList, SingleValidationError{
		Error:  fmt.Sprintf(message, messageArgs...),
		Object: object,
	})
}

// Validate walks the target object looking for elements (including the top level element) that
// implement the validatable interface.
//
// If validation errors are found, they will be accumulated in the ValidationResult.
//
// After successful validation, this function returns a ValidationResult which contains all the
// validation errors for the target.
//
// This function only returns an error if a failure keeps the validation from completing or
// the validation result is invalid due to an error.
func ValidateObject(target interface{}) (ValidationResult, error) {

	result := ValidationResult{}

	validationWorker := func(needle interface{}, path string) error {
		validationTarget := needle.(Validatable)

		//validate
		err := validationTarget.Validate(&result)
		//this error indicates the validation results are not valid, so we propogate it out and
		//stop validation
		if err != nil {
			return err
		}
		//validation complete
		return nil
	}

	validatableType := reflect.TypeOf((*Validatable)(nil)).Elem()
	err := walker.WalkObjectImmutable(target, validatableType, validationWorker)
	if err != nil {
		return ValidationResult{}, err
	}
	return result, nil
}
