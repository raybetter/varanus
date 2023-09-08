package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/kr/pretty"
)

type ValidationError struct {
	Error  string
	Object interface{}
}

type ValidationProcessError struct {
	Errors     []ValidationError
	ErrorValue string
}

func (vpe ValidationProcessError) Print() {

	if len(vpe.Errors) > 0 {

		fmt.Println("*****************************************************")
		fmt.Printf("%d Validation Errors\n", len(vpe.Errors))

		for _, validationError := range vpe.Errors {
			fmt.Printf("Error %s on object:\n%# v\n", validationError.Error, pretty.Formatter(validationError.Object))
			fmt.Println("--------------------")
		}

		fmt.Println("*****************************************************")

	} else {
		fmt.Printf("No validation errors")
	}
}

func (vpe ValidationProcessError) Error() string {
	return vpe.ErrorValue
}

type ValidationProcess struct {
	Errors []ValidationError
}

func (vp *ValidationProcess) addValidationError(object interface{}, message string, messageArgs ...interface{}) {
	vp.Errors = append(vp.Errors, ValidationError{
		Error:  fmt.Sprintf(message, messageArgs...),
		Object: object,
	})
}

func (vp *ValidationProcess) Validate(target Validatable) error {
	err := target.Validate(vp)
	return err
}

func (vp *ValidationProcess) Finalize() error {

	if len(vp.Errors) > 0 {

		var sb strings.Builder
		for _, validationError := range vp.Errors {
			sb.Write([]byte(fmt.Sprintf("Error: %s on object %#v", validationError.Error, validationError.Object)))
		}

		vpe := ValidationProcessError{
			Errors:     vp.Errors,
			ErrorValue: sb.String(),
		}

		return vpe
	}

	return nil
}

func (vp *ValidationProcess) Print() {

	if len(vp.Errors) > 0 {

		fmt.Println("*****************************************************")
		fmt.Printf("%d Validation Errors\n", len(vp.Errors))

		for _, validationError := range vp.Errors {
			fmt.Printf("Error %s on object:\n%# v\n", validationError.Error, pretty.Formatter(validationError.Object))
			fmt.Println("--------------------")
		}

		fmt.Println("*****************************************************")

	} else {
		fmt.Printf("No validation errors")
	}
}

type Validatable interface {
	Validate(vp *ValidationProcess) error
}

//=================================================================================================
// Validation helper functions

// source: https://stackoverflow.com/questions/106179/regular-expression-to-match-dns-hostname-or-ip-address
var HostnameRe = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`)

func IsUrlHost(candidate string) bool {
	//we don't want the candidate to have to have a scheme, so we add our own
	candidateWithScheme := "varanus://" + candidate

	u, err := url.ParseRequestURI(candidateWithScheme)

	fmt.Printf("%s --> %#v\n\n", candidate, u)

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
