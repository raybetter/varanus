package config

import (
	"testing"
	"varanus/internal/secrets"
	"varanus/internal/util"
	"varanus/internal/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMailAccountConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator         func(c *MailAccountConfig)
		Error           string
		ErrorObjectType interface{}
	}

	testCases := []TestCase{
		{
			Mutator:         func(c *MailAccountConfig) { c.Name = "" },
			Error:           "account names must not be empty or whitespace",
			ErrorObjectType: MailAccountConfig{},
		},
		{
			Mutator:         func(c *MailAccountConfig) { c.SMTP = nil; c.IMAP = nil },
			Error:           "Every server config must specify one of the imap or smtp sections.  They cannot both be empty.",
			ErrorObjectType: MailAccountConfig{},
		},
		//pass one error through to each of the IMAP and SMTP structs to check end to end behavior
		{
			Mutator:         func(c *MailAccountConfig) { c.SMTP.Username = "" },
			Error:           "username must not be empty or whitespace",
			ErrorObjectType: SMTPConfig{},
		},
		{
			Mutator:         func(c *MailAccountConfig) { c.IMAP.Username = "" },
			Error:           "username must not be empty or whitespace",
			ErrorObjectType: IMAPConfig{},
		},
	}

	baseConfig := MailAccountConfig{
		Name: "test1",
		SMTP: &SMTPConfig{
			SenderAddress: "example@example.com",
			ServerAddress: "mail.example.com",
			Port:          465,
			Username:      "joe@example.com",
			Password:      secrets.CreateSealedItem("sealed(+abcdef==)"),
		},
		IMAP: &IMAPConfig{
			ServerAddress: "mail.example.com",
			Port:          993,
			Username:      "joe@example.com",
			Password:      secrets.CreateSealedItem("sealed(+abcdef==)"),
		},
	}

	{ //nominal case test should have no errors
		config := util.DeepCopy(baseConfig).(MailAccountConfig) //make a copy of the config
		validationResult, err := validation.ValidateObject(config)
		assert.Nil(t, err)
		assert.Equal(t, 0, validationResult.GetErrorCount())
	}

	// test loop
	for index, testCase := range testCases {
		//do validation
		config := util.DeepCopy(baseConfig).(MailAccountConfig) //make a copy of the config
		testCase.Mutator(&config)                               //modify the config
		validationResult, err := validation.ValidateObject(config)
		//checks
		assert.Nil(t, err)
		require.Equal(t, validationResult.GetErrorCount(), 1, "for test %d", index)
		singleError := validationResult.GetErrorList()[0]
		assert.IsType(t, testCase.ErrorObjectType, singleError.Object, "for test %d", index)
		assert.Contains(t, singleError.Error, testCase.Error, "for test %d", index)
	}

}
