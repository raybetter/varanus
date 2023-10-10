package config

import (
	"testing"
	"varanus/internal/secrets"
	"varanus/internal/util"
	"varanus/internal/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMailConfig(t *testing.T) {

	config := MailConfig{
		Accounts: []MailAccountConfig{
			{
				Name: "test1",
				SMTP: &SMTPConfig{
					SenderAddress: "example@example.com",
					ServerAddress: "mail.example.com",
					Port:          465,
					Username:      "joe@example.com",
					Password:      secrets.CreateSealedItem("+abcdef=="),
				},
			},
			{
				Name: "test2",
				SMTP: &SMTPConfig{
					SenderAddress: "example@example.com",
					ServerAddress: "mail.example.com",
					Port:          465,
					Username:      "joe@example.com",
					Password:      secrets.CreateSealedItem("+abcdef=="),
				},
			},
		},
		SendLimits: []SendLimitConfig{
			{
				MinPeriod:    15,
				AccountNames: []string{"test1"},
			},
		},
	}

	assert.Equal(t, &config.Accounts[0], config.GetAccountByName("test1"))
	assert.Equal(t, &config.Accounts[1], config.GetAccountByName("test2"))
	assert.Nil(t, config.GetAccountByName("nonexistent"))

}

func TestMailConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator         func(c *MailConfig)
		Error           string
		ErrorObjectType interface{}
	}

	testCases := []TestCase{
		{
			Mutator:         func(c *MailConfig) { c.Accounts[1].Name = "test1" },
			Error:           "duplicate account name 'test1'",
			ErrorObjectType: MailConfig{},
		},
	}

	baseConfig := MailConfig{
		Accounts: []MailAccountConfig{
			{
				Name: "test1",
				SMTP: &SMTPConfig{
					SenderAddress: "example@example.com",
					ServerAddress: "mail.example.com",
					Port:          465,
					Username:      "joe@example.com",
					Password:      secrets.CreateSealedItem("sealed(+abcdef==)"),
				},
			},
			{
				Name: "test2",
				SMTP: &SMTPConfig{
					SenderAddress: "example@example.com",
					ServerAddress: "mail.example.com",
					Port:          465,
					Username:      "joe@example.com",
					Password:      secrets.CreateSealedItem("sealed(+abcdef==)"),
				},
			},
		},
		SendLimits: []SendLimitConfig{
			{
				MinPeriod:    15,
				AccountNames: []string{"test1"},
			},
		},
	}
	{
		//nominal case test should have no errors
		config := util.DeepCopy(baseConfig).(MailConfig) //make a copy of the config

		//call validation on the target object
		validationResult := validation.ValidationResult{}
		err := config.Validate(&validationResult, config)

		assert.Nil(t, err)
		assert.Equal(t, 0, validationResult.GetErrorCount())
	}
	// test loop
	for index, testCase := range testCases {
		//do validation
		config := util.DeepCopy(baseConfig).(MailConfig) //make a copy of the config
		testCase.Mutator(&config)                        //modify the config
		//call validation on the target object
		validationResult := validation.ValidationResult{}
		err := config.Validate(&validationResult, config)
		//checks
		assert.Nil(t, err)
		require.Equal(t, validationResult.GetErrorCount(), 1, "for test %d", index)
		singleError := validationResult.GetErrorList()[0]
		assert.IsType(t, testCase.ErrorObjectType, singleError.Object, "for test %d", index)
		assert.Contains(t, singleError.Error, testCase.Error, "for test %d", index)
	}

}
