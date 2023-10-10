package config

import (
	"testing"
	"varanus/internal/secrets"
	"varanus/internal/util"
	"varanus/internal/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIMAPConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator         func(c *IMAPConfig)
		Error           string
		ErrorObjectType interface{}
	}

	testCases := []TestCase{
		{
			Mutator:         func(c *IMAPConfig) { c.ServerAddress = "not/a/valid/hostname" },
			Error:           "server_address 'not/a/valid/hostname' is not a valid hostname",
			ErrorObjectType: IMAPConfig{},
		},
		{
			Mutator:         func(c *IMAPConfig) { c.Port = 0 },
			Error:           "port value is required and cannot be 0",
			ErrorObjectType: IMAPConfig{},
		},
		{
			Mutator:         func(c *IMAPConfig) { c.Username = "" },
			Error:           "username must not be empty or whitespace",
			ErrorObjectType: IMAPConfig{},
		},
		{
			Mutator:         func(c *IMAPConfig) { c.Username = "  " },
			Error:           "username must not be empty or whitespace",
			ErrorObjectType: IMAPConfig{},
		},
		{
			Mutator:         func(c *IMAPConfig) { c.Password = secrets.CreateUnsafeSealedItem("", false) },
			Error:           "SealedItem with an unsealed value should not be empty",
			ErrorObjectType: secrets.CreateSealedItem(""),
		},
		{
			Mutator: func(c *IMAPConfig) {
				c.Password = secrets.CreateUnsafeSealedItem("sealed(f0o bar not a valid encoded password)", true)
			},
			Error:           "value does not match the expected format for an encrypted, encoded string",
			ErrorObjectType: secrets.CreateSealedItem(""),
		},
	}

	baseConfig := IMAPConfig{
		ServerAddress: "mail.example.com",
		Port:          993,
		Username:      "joe@example.com",
		Password:      secrets.CreateSealedItem("+abcdef=="),
	}

	{ //nominal case test should have no errors
		config := util.DeepCopy(baseConfig).(IMAPConfig) //make a copy of the config
		validationResult, err := validation.ValidateObject(config)
		assert.Nil(t, err)
		assert.Equal(t, 0, validationResult.GetErrorCount())
	}

	// test loop
	for index, testCase := range testCases {
		//do validation
		config := util.DeepCopy(baseConfig).(IMAPConfig) //make a copy of the config
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
