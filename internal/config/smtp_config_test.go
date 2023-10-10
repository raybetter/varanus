package config

import (
	"testing"
	"varanus/internal/secrets"
	"varanus/internal/util"
	"varanus/internal/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmtpConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator         func(c *SMTPConfig)
		Error           string
		ErrorObjectType interface{}
	}

	testCases := []TestCase{
		{
			Mutator:         func(c *SMTPConfig) { c.SenderAddress = "Not an email address" },
			Error:           "sender_address 'Not an email address' is not a valid email",
			ErrorObjectType: SMTPConfig{},
		},
		{
			Mutator:         func(c *SMTPConfig) { c.SenderAddress = "  " },
			Error:           "sender_address '' is not a valid email",
			ErrorObjectType: SMTPConfig{},
		},
		{
			Mutator:         func(c *SMTPConfig) { c.ServerAddress = "not/a/valid/hostname" },
			Error:           "server_address 'not/a/valid/hostname' is not a valid hostname",
			ErrorObjectType: SMTPConfig{},
		},
		{
			Mutator:         func(c *SMTPConfig) { c.Port = 0 },
			Error:           "port value is required and cannot be 0",
			ErrorObjectType: SMTPConfig{},
		},
		{
			Mutator:         func(c *SMTPConfig) { c.Username = "" },
			Error:           "username must not be empty or whitespace",
			ErrorObjectType: SMTPConfig{},
		},
		{
			Mutator:         func(c *SMTPConfig) { c.Username = "  " },
			Error:           "username must not be empty or whitespace",
			ErrorObjectType: SMTPConfig{},
		},
		{
			Mutator:         func(c *SMTPConfig) { c.Password = secrets.CreateUnsafeSealedItem("  ", true) },
			Error:           "value does not match the expected format for an encrypted, encoded string",
			ErrorObjectType: secrets.CreateSealedItem(""),
		},
		{
			Mutator: func(c *SMTPConfig) {
				c.Password = secrets.CreateUnsafeSealedItem("foo bar is not an encoded password", true)
			},
			Error:           "value does not match the expected format for an encrypted, encoded string",
			ErrorObjectType: secrets.CreateSealedItem(""),
		},
	}

	baseConfig := SMTPConfig{
		SenderAddress: "example@example.com",
		ServerAddress: "mail.example.com",
		Port:          465,
		Username:      "joe@example.com",
		Password:      secrets.CreateSealedItem("sealed(+abcdef==)"),
	}
	{
		//nominal case test should have no errors
		config := util.DeepCopy(baseConfig).(SMTPConfig) //make a copy of the config
		validationResult, err := validation.ValidateObject(config)
		assert.Nil(t, err)
		assert.Equal(t, 0, validationResult.GetErrorCount())
	}
	// test loop
	for index, testCase := range testCases {
		//do validation
		config := util.DeepCopy(baseConfig).(SMTPConfig) //make a copy of the config
		testCase.Mutator(&config)                        //modify the config
		validationResult, err := validation.ValidateObject(config)
		//checks
		assert.Nil(t, err)
		require.Equal(t, validationResult.GetErrorCount(), 1, "for test %d", index)
		singleError := validationResult.GetErrorList()[0]
		assert.IsType(t, testCase.ErrorObjectType, singleError.Object, "for test %d", index)
		assert.Contains(t, singleError.Error, testCase.Error, "for test %d", index)
	}

}
