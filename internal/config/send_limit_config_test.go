package config

import (
	"testing"
	"varanus/internal/util"
	"varanus/internal/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendLimitValidation(t *testing.T) {

	type TestCase struct {
		Mutator func(c *SendLimitConfig)
		Error   string
	}

	testCases := []TestCase{
		{
			Mutator: func(c *SendLimitConfig) { c.AccountNames = nil },
			Error:   "send limits account name list must not be empty",
		},
		{
			Mutator: func(c *SendLimitConfig) { c.AccountNames = []string{} },
			Error:   "send limits account name list must not be empty",
		},
		{
			Mutator: func(c *SendLimitConfig) { c.MinPeriod = 0 },
			Error:   "send limit min_period must be non-negative, not '0'",
		},
	}

	baseConfig :=
		VaranusConfig{
			Mail: MailConfig{
				Accounts: []MailAccountConfig{
					{
						Name: "test1",
					},
					{
						Name: "test2",
					},
				},
				SendLimits: []SendLimitConfig{
					{
						MinPeriod:    15,
						AccountNames: []string{"test1", "test2"},
					},
				},
			},
		}

	{
		//nominal case test should have no errors
		config := util.DeepCopy(baseConfig).(VaranusConfig) //make a copy of the config
		//call validation on the target object
		validationResult := validation.ValidationResult{}
		err := config.Mail.SendLimits[0].Validate(&validationResult, config)
		assert.Nil(t, err)
		assert.Equal(t, 0, validationResult.GetErrorCount())
	}

	// test loop
	for index, testCase := range testCases {
		//do validation
		config := util.DeepCopy(baseConfig).(VaranusConfig) //make a copy of the config
		testCase.Mutator(&config.Mail.SendLimits[0])        //modify the config

		//call validation on the target object
		validationResult := validation.ValidationResult{}
		err := config.Mail.SendLimits[0].Validate(&validationResult, config)

		//checks
		assert.Nil(t, err)
		require.Equal(t, validationResult.GetErrorCount(), 1, "for test %d", index)
		singleError := validationResult.GetErrorList()[0]
		assert.IsType(t, SendLimitConfig{}, singleError.Object, "for test %d", index)
		assert.Contains(t, singleError.Error, testCase.Error, "for test %d", index)
	}

}
