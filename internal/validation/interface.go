package validation

// Validatable is the interface that objects that are to be validated with a ValidationProcess
// should implement this interface.
type Validatable interface {
	// Validate implementations should contain checks to ensure that the Validatable object is
	// valid. Vaidation errors that are encountered should be added to the ValidationErrorTracker
	// process with vet.AddValidationError().
	//
	// Each validateable object only needs to run the checks that are necessary at that objects level
	// It does not need to call Validate on objects lower down in the hierarchy that also implement
	// Validatable.
	//
	// Validate should only return an error if an failure prevents the validation checks from
	// completing.  Any validation error shourl be added to the ValidationErrorTracker.
	//
	// the root parameter is the top-level validation object, e.g. the object that ValidateObject
	// is calle with.  It is provided to allow an object to access global context, (e.g. if
	// validating config, checking a list elsewhere in the config for the presence of a name).
	// The implementation of a validation hierarchy should know what object is expected for root
	// and return an error if the wrong object is passed.
	Validate(vet ValidationErrorTracker, root interface{}) error
}

// ValidationErrorTracker provides the required interface that Validateable objects can access to
// log their validation errors.
//
// The ValidationError error struct is the primary implementation of this.
type ValidationErrorTracker interface {
	AddValidationError(object interface{}, message string, messageArgs ...interface{})
}
