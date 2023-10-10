package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockValidatable struct {
	validationError   string // if set, will go as a validation error added to the ValidationErrorTracker
	validationFailure string // if set, will return as a validation failure returned from Validate
	checkRoot         bool   //if set, try to cast the root object to mockValdationTargetTop
	expectedRootValue string //if checkRoot is set, expect mockValidationTargetTop to have this value
}

func (mv mockValidatable) Validate(vet ValidationErrorTracker, root interface{}) error {
	if mv.validationFailure != "" {
		return fmt.Errorf(mv.validationFailure)
	}
	//no validation failure

	if mv.validationError != "" {
		vet.AddValidationError(mv, mv.validationError)
	}
	//else no validation error

	if mv.checkRoot {
		rootObj, ok := root.(mockValidationTargetTop)
		if !ok {
			return fmt.Errorf("Could not get a mockValidationTargetTop from %s", root)
		}
		if rootObj.value != mv.expectedRootValue {
			vet.AddValidationError(mv, "root error mismatch %s vs %s", rootObj.value, mv.expectedRootValue)
		}
	}

	return nil
}

type mockValidationTargetTop struct {
	mockValidatable
	Middle1 mockValidationTargetMiddle
	Middle2 mockValidationTargetMiddle
	value   string
}

type mockValidationTargetMiddle struct {
	mockValidatable
	Bottom1 MockValidationTargetBottom
	Bottom2 MockValidationTargetBottom
}

type MockValidationTargetBottom struct {
	mockValidatable
}

func TestValidationProcessNoErrors(t *testing.T) {

	validationTarget := mockValidationTargetTop{
		mockValidatable: mockValidatable{"", "", false, ""},
		Middle1: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"", "", false, ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", false, ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", false, ""},
			},
		},
		Middle2: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"", "", false, ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", false, ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", false, ""},
			},
		},
	}

	expectedHumanReadableString := `No validation errors
`

	validationResult, err := ValidateObject(validationTarget)
	assert.Nil(t, err)
	assert.Equal(t, 0, validationResult.GetErrorCount())
	assert.Equal(t, 7, validationResult.GetValidationCount())
	assert.Len(t, validationResult.GetErrorList(), 0)
	assert.Nil(t, validationResult.AsError())
	assert.Equal(t, expectedHumanReadableString, validationResult.HumanReadable())

}

func TestValidationProcessRootCastNominal(t *testing.T) {

	validationTarget := mockValidationTargetTop{
		mockValidatable: mockValidatable{"", "", false, ""},
		Middle1: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"", "", false, ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", true, "argyle"},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", false, ""},
			},
		},
		Middle2: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"", "", false, ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", false, ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", false, ""},
			},
		},
		value: "argyle",
	}

	validationResult, err := ValidateObject(validationTarget)
	assert.Nil(t, err)
	assert.Equal(t, 0, validationResult.GetErrorCount())
	assert.Equal(t, 7, validationResult.GetValidationCount())
}

func TestValidationProcessRootCastValidationError(t *testing.T) {

	//the root comparison value is different, resulting in a validation error

	validationTarget := mockValidationTargetTop{
		mockValidatable: mockValidatable{"", "", false, ""},
		Middle1: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"", "", false, ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", true, "not argyle"},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", false, ""},
			},
		},
		Middle2: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"", "", false, ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", false, ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", "", false, ""},
			},
		},
		value: "argyle",
	}

	validationResult, err := ValidateObject(validationTarget)
	assert.Nil(t, err)
	assert.Equal(t, 1, validationResult.GetErrorCount())
	assert.Equal(t, 7, validationResult.GetValidationCount())
	assert.Equal(t, validationResult.errorList[0].Error, "root error mismatch argyle vs not argyle")
}

func TestValidationProcessWithErrors(t *testing.T) {

	validationTarget := mockValidationTargetTop{
		mockValidatable: mockValidatable{"top validation error", "", false, ""},
		Middle1: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"middle validation error 1", "", false, ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 11", "", false, ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 12", "", false, ""},
			},
		},
		Middle2: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"middle validation error 2", "", false, ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 21", "", false, ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 22", "", false, ""},
			},
		},
	}

	expectedErrors := []string{
		"top validation error",
		"middle validation error 1",
		"bottom validation error 11",
		"bottom validation error 12",
		"middle validation error 2",
		"bottom validation error 21",
		"bottom validation error 22",
	}
	expectedObjects := []interface{}{
		validationTarget.mockValidatable,
		validationTarget.Middle1.mockValidatable,
		validationTarget.Middle1.Bottom1.mockValidatable,
		validationTarget.Middle1.Bottom2.mockValidatable,
		validationTarget.Middle2.mockValidatable,
		validationTarget.Middle2.Bottom1.mockValidatable,
		validationTarget.Middle2.Bottom2.mockValidatable,
	}

	validationResult, err := ValidateObject(validationTarget)
	assert.Nil(t, err)
	assert.Equal(t, 7, validationResult.GetErrorCount())
	assert.Equal(t, 7, validationResult.GetValidationCount())
	errorList := validationResult.GetErrorList()
	humanReadableResult := validationResult.HumanReadable()
	errorResult := validationResult.AsError().Error()
	for index := 0; index < len(expectedErrors); index++ {
		assert.Equal(t, expectedErrors[index], errorList[index].Error, "for index %d", index)
		assert.Equal(t, expectedObjects[index], errorList[index].Object, "for index %d", index)
		assert.Contains(t, humanReadableResult, expectedErrors[index], "for index %d", index)
		assert.Contains(t, errorResult, expectedErrors[index], "for index %d", index)
	}

}

func TestValidatableWithValidationFailures(t *testing.T) {

	validationTarget := mockValidationTargetTop{
		mockValidatable: mockValidatable{"top validation error", "", false, ""},
		Middle1: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"middle validation error 1", "", false, ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 11", "fail at bottom 11", false, ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 12", "", false, ""},
			},
		},
		Middle2: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"middle validation error 2", "", false, ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 21", "", false, ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 22", "fail at bottom 22", false, ""},
			},
		},
	}

	validationResult, err := ValidateObject(validationTarget)
	assert.ErrorContains(t, err, "fail at bottom 11")
	assert.Equal(t, 0, validationResult.GetErrorCount())
	assert.Equal(t, 0, validationResult.GetValidationCount())
	assert.Len(t, validationResult.GetErrorList(), 0)
	assert.Nil(t, validationResult.AsError())

}
