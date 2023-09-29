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

type mockValidatable struct {
	validationError   string // if set, will go as a validation error added to the ValidationErrorTracker
	validationFailure string // if set, will return as a validation failure returned from Validate
}

func (mv mockValidatable) Validate(vet ValidationErrorTracker) error {
	if mv.validationFailure != "" {
		return fmt.Errorf(mv.validationFailure)
	}
	if mv.validationError != "" {
		vet.AddValidationError(mv, mv.validationError)
	}
	//else no validation error

	//no validation failure
	return nil
}

type mockValidationTargetTop struct {
	mockValidatable
	Middle1 mockValidationTargetMiddle
	Middle2 mockValidationTargetMiddle
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
		mockValidatable: mockValidatable{"", ""},
		Middle1: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"", ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", ""},
			},
		},
		Middle2: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"", ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"", ""},
			},
		},
	}

	expectedHumanReadableString := `No validation errors
`

	validationResult, err := ValidateObject(validationTarget)
	assert.Nil(t, err)
	assert.Equal(t, 0, validationResult.GetErrorCount())
	assert.Len(t, validationResult.GetErrorList(), 0)
	assert.Nil(t, validationResult.AsError())
	assert.Equal(t, expectedHumanReadableString, validationResult.HumanReadable())

}

func TestValidationProcessWithErrors(t *testing.T) {

	validationTarget := mockValidationTargetTop{
		mockValidatable: mockValidatable{"top validation error", ""},
		Middle1: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"middle validation error 1", ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 11", ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 12", ""},
			},
		},
		Middle2: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"middle validation error 2", ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 21", ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 22", ""},
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
		mockValidatable: mockValidatable{"top validation error", ""},
		Middle1: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"middle validation error 1", ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 11", "fail at bottom 11"},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 12", ""},
			},
		},
		Middle2: mockValidationTargetMiddle{
			mockValidatable: mockValidatable{"middle validation error 2", ""},
			Bottom1: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 21", ""},
			},
			Bottom2: MockValidationTargetBottom{
				mockValidatable: mockValidatable{"bottom validation error 22", "fail at bottom 22"},
			},
		},
	}

	validationResult, err := ValidateObject(validationTarget)
	assert.ErrorContains(t, err, "fail at bottom 11")
	assert.Equal(t, 0, validationResult.GetErrorCount())
	assert.Len(t, validationResult.GetErrorList(), 0)
	assert.Nil(t, validationResult.AsError())

}
