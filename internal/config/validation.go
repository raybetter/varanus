package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/kr/pretty"
)

// SingleValidationError captures the context for a validation error discovered by the execution of
// a ValidationProcess.
type SingleValidationError struct {
	// Error describes the validation error
	Error string
	// Object provides the context of the object where the error occured.
	Object interface{}
}

// ValidationError is a collection of SingleValidationErrors that can be returns as a go error
// object.
type ValidationError struct {
	ErrorList []SingleValidationError
}

// HumanReadable returns a human-readable formatted list of the validation errors suitable for
// printing to the terminal.
func (ve ValidationError) HumanReadable() string {
	return makeHumanReadableErrorList(ve.ErrorList)
}

// Error implements the go error interface by returning a string concatenation of all the
// SingleValidationErrors in the ValidationError.
func (ve ValidationError) Error() string {
	var sb strings.Builder
	for _, validationError := range ve.ErrorList {
		sb.Write([]byte(fmt.Sprintf("Error: %s on object %#v\n", validationError.Error, validationError.Object)))
	}
	return sb.String()
}

// ValidationProcess is the worker that collects validation errors from one or more objects that
// implement the Validatable interface.
type ValidationProcess struct {
	ErrorList             []SingleValidationError
	validationFailedError *error
}

// AddValidationError should be called by the Validatable objects in their Validate() impelementation
// to add a validation error.
// - object is the object the validation error occurs on (provided for context)
// - message is a format string describing the error
// - any remaining parameters will be supplied when parsing the message format string
//
// Panics if called after a validation failure (see Validate() function description).
func (vp *ValidationProcess) AddValidationError(object interface{}, message string, messageArgs ...interface{}) {
	if vp.validationFailedError != nil {
		panic(fmt.Errorf("illegal call to AddValidationError after validation failure: %w", *vp.validationFailedError))
	}

	vp.ErrorList = append(vp.ErrorList, SingleValidationError{
		Error:  fmt.Sprintf(message, messageArgs...),
		Object: object,
	})
}

// Validate should be called to execute the validation process of the target.
//
// If validation errors are found, they will be accumulated in the ValidationProcess Errors list.
//
// This function should only return an error if a failure keeps the validation from completing,
// causing target.Validate() to return an error.  Once this has occured, the
// ValidationProcess object should not be used for validation.  Any successive call to
// Validate, GetFinalValidationError, or AddValidationError will panic.
//
// Panics if called after a validation failure from a previous call to Validate.
func (vp *ValidationProcess) Validate(target Validatable) error {
	if vp.validationFailedError != nil {
		panic(fmt.Errorf("illegal call to Validate after validation failure: %w", *vp.validationFailedError))
	}

	err := target.Validate(vp)

	if err != nil {
		vp.validationFailedError = &err
	}

	return err
}

// Finalize returns an ValidationError if any calls to Validate() resulted in validation errors,
// otherwise it returns nil.
//
// Panics if called after a validation failure (see Validate() function description).
func (vp *ValidationProcess) GetFinalValidationError() error {
	if vp.validationFailedError != nil {
		panic(
			fmt.Errorf("illegal call to GetFinalValidationError after validation failure: %w", *vp.validationFailedError),
		)
	}

	if len(vp.ErrorList) > 0 {

		ve := ValidationError{
			ErrorList: vp.ErrorList,
		}

		return ve
	}

	return nil
}

// HumanReadable returns a human-readable formatted list of the validation errors suitable for
// printing to the terminal.
//
// Panics if called after a validation failure (see Validate() function description).
func (vp *ValidationProcess) HumanReadable() string {
	if vp.validationFailedError != nil {
		panic(fmt.Errorf("illegal call to HumanReadable after validation failure: %w", *vp.validationFailedError))
	}

	return makeHumanReadableErrorList(vp.ErrorList)
}

// makeHumanReadableErrorList is a helper function that implementat the formatting of a list
// of SingleValidationErrors into a human-readable format.
func makeHumanReadableErrorList(errors []SingleValidationError) string {
	var sb strings.Builder

	if len(errors) > 0 {

		sb.WriteString("*****************************************************\n")
		sb.WriteString(fmt.Sprintf("%d Validation Errors\n", len(errors)))

		for _, validationError := range errors {
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

// Validatable is the interface that objects that are to be validated with a ValidationProcess
// should implement this interface.
type Validatable interface {
	// Validate implementations should contain checks to ensure that the Validatable object is
	// valid. Vaidation errors that are encountered should be added to the Validation process with
	// vp.AddValidationError().
	//
	// If a Validatable object is heirarchical (e.g. a struct containing other Validatable objects)
	// it can call vp.Validate() with the lower level objects.
	//
	// Validate should only return an error if an failure prevents the validation checks from
	// completing.
	Validate(vp *ValidationProcess) error
}

//=================================================================================================
// Validation helper functions

// source: https://stackoverflow.com/questions/106179/regular-expression-to-match-dns-hostname-or-ip-address
var HostnameRe = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`)

// IsURLHost returns true if the candidate string is a valid hostname
func IsUrlHost(candidate string) bool {
	//we don't want the candidate to have to have a scheme, so we add our own
	candidateWithScheme := "varanus://" + candidate

	u, err := url.ParseRequestURI(candidateWithScheme)

	// fmt.Printf("%s --> %#v\n\n", candidate, u)

	//not a valid URL if:
	// - err is not nil
	// - the scheme is not the one we added
	// - the host is not the whole candidate string (this saves from having to check a bunch of
	//	 path variables in the u result)
	// - the candidate has anything other than letters, numbers, dashes, and dots

	if !HostnameRe.Match([]byte(candidate)) {
		return false
	}

	return err == nil && u.Scheme == "varanus" && u.Host == candidate
}
