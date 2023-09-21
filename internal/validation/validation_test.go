package validation

import (
	"fmt"
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
	vp.AddValidationError(
		obj1,
		"error 1 '%s'", "string arg 1",
	)
	assert.Len(t, vp.ErrorList, 1)
	assert.Equal(t, obj1, vp.ErrorList[0].Object)
	assert.Equal(t, "error 1 'string arg 1'", vp.ErrorList[0].Error)

	obj2 := TestObject{"object 2"}
	vp.AddValidationError(
		obj2,
		"error 2 '%s'", "string arg 2",
	)
	assert.Len(t, vp.ErrorList, 2)
	assert.Equal(t, obj2, vp.ErrorList[1].Object)
	assert.Equal(t, "error 2 'string arg 2'", vp.ErrorList[1].Error)

	finalizeError := vp.GetFinalValidationError()

	expectedHumanReadableString := `*****************************************************
2 Validation Errors
Error error 1 'string arg 1' on object:
validation.TestObject{value:"object 1"}
--------------------
Error error 2 'string arg 2' on object:
validation.TestObject{value:"object 2"}
--------------------
*****************************************************
`

	assert.Equal(t, expectedHumanReadableString, vp.HumanReadable())

	errString1 := "error 1 'string arg 1' on object validation.TestObject{value:\"object 1\"}"
	errString2 := "error 2 'string arg 2' on object validation.TestObject{value:\"object 2\"}"

	assert.Contains(t, finalizeError.Error(), errString1)
	assert.Contains(t, finalizeError.Error(), errString2)

	vpe, ok := finalizeError.(ValidationError)

	assert.Truef(t, ok, "Should alwasy be a ValidationProcessError")

	assert.Equal(t, vp.ErrorList, vpe.ErrorList)

	assert.Equal(t, expectedHumanReadableString, vpe.HumanReadable())

}

type validatableTestObjectNoErrors struct {
}

func (vtone validatableTestObjectNoErrors) Validate(vp *ValidationProcess) error {
	return nil
}

func TestValidationProcessWithoutErrors(t *testing.T) {

	vp := ValidationProcess{}

	//no errors added
	vto := validatableTestObjectNoErrors{}

	assert.Len(t, vp.ErrorList, 0)

	vp.Validate(vto)

	assert.Len(t, vp.ErrorList, 0)

	finalizeError := vp.GetFinalValidationError()

	assert.Nil(t, finalizeError)

	assert.Equal(t, "No validation errors\n", vp.HumanReadable())

}

type validatableTestObjectFails struct {
}

func (vtof validatableTestObjectFails) Validate(vp *ValidationProcess) error {
	return fmt.Errorf("Test error")
}

func TestValidatableWithValidationFailures(t *testing.T) {

	vp := ValidationProcess{}

	vto := validatableTestObjectFails{}

	err := vp.Validate(vto)

	assert.NotNil(t, vp.validationFailedError)

	assert.ErrorContains(t, err, "Test error")

	assert.PanicsWithError(t,
		"illegal call to AddValidationError after validation failure: Test error",
		func() { vp.AddValidationError(nil, "") },
	)

	assert.PanicsWithError(t,
		"illegal call to Validate after validation failure: Test error",
		func() { vp.Validate(validatableTestObjectNoErrors{}) },
	)

	assert.PanicsWithError(t,
		"illegal call to GetFinalValidationError after validation failure: Test error",
		func() { vp.GetFinalValidationError() },
	)

	assert.PanicsWithError(t,
		"illegal call to HumanReadable after validation failure: Test error",
		func() { vp.HumanReadable() },
	)

}
