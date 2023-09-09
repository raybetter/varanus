package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUrlHost(t *testing.T) {
	testCases := map[string]bool{
		"google.com":            true,
		"smtp.varanus.org":      true,
		"smtp.red-and-blue.org": true,
		"poop.wtf":              true,
		"foo.com/argyle/socks":  false,
		"http://foo.com":        false,
		":::foo.com":            false,
		"foo,bar.com":           false,
		"foo__bar.com":          false,
		".leading.dot":          false,
	}

	for candidate, expectedResult := range testCases {
		actualResult := IsUrlHost(candidate)
		assert.Equalf(t, expectedResult, actualResult,
			"expected candidate value '%s' to be %t but it is %t",
			candidate, expectedResult, actualResult)
	}

}

func TestValidationProcessWithErrors(t *testing.T) {

	type TestObject struct {
		value string
	}

	vp := ValidationProcess{}

	obj1 := TestObject{"object 1"}
	vp.addValidationError(
		obj1,
		"error 1 '%s'", "string arg 1",
	)
	assert.Len(t, vp.Errors, 1)
	assert.Equal(t, obj1, vp.Errors[0].Object)
	assert.Equal(t, "error 1 'string arg 1'", vp.Errors[0].Error)

	obj2 := TestObject{"object 2"}
	vp.addValidationError(
		obj2,
		"error 2 '%s'", "string arg 2",
	)
	assert.Len(t, vp.Errors, 2)
	assert.Equal(t, obj2, vp.Errors[1].Object)
	assert.Equal(t, "error 2 'string arg 2'", vp.Errors[1].Error)

	finalizeError := vp.Finalize()

	expectedHumanReadableString := `*****************************************************
2 Validation Errors
Error error 1 'string arg 1' on object:
config.TestObject{value:"object 1"}
--------------------
Error error 2 'string arg 2' on object:
config.TestObject{value:"object 2"}
--------------------
*****************************************************
`

	assert.Equal(t, expectedHumanReadableString, vp.HumanReadable())

	errString1 := "error 1 'string arg 1' on object config.TestObject{value:\"object 1\"}"
	errString2 := "error 2 'string arg 2' on object config.TestObject{value:\"object 2\"}"

	assert.Contains(t, finalizeError.Error(), errString1)
	assert.Contains(t, finalizeError.Error(), errString2)

	vpe, ok := finalizeError.(ValidationProcessError)

	assert.Truef(t, ok, "Should alwasy be a ValidationProcessError")

	assert.Contains(t, vpe.ErrorValue, errString1)
	assert.Contains(t, vpe.ErrorValue, errString2)

	assert.Equal(t, vp.Errors, vpe.Errors)

	assert.Equal(t, expectedHumanReadableString, vpe.HumanReadable())

}

func TestValidationProcessWithoutErrors(t *testing.T) {

	vp := ValidationProcess{}

	//no errors added

	assert.Len(t, vp.Errors, 0)

	finalizeError := vp.Finalize()

	assert.Nil(t, finalizeError)

	assert.Equal(t, "No validation errors\n", vp.HumanReadable())

}
