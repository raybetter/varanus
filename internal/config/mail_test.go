package config

import (
	"testing"
	"varanus/internal/secrets"
	"varanus/internal/validation"

	"github.com/stretchr/testify/assert"
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

	//nominal case test should have no errors
	vp := &validation.ValidationProcess{}
	config := baseConfig
	//validation
	vp.Validate(config)
	//checks
	assert.Len(t, vp.ErrorList, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &validation.ValidationProcess{}
		config := baseConfig
		testCase.Mutator(&config)
		//validation
		vp.Validate(config)
		//checks
		assert.Len(t, vp.ErrorList, 1, "for test %d", index)
		assert.IsType(t, testCase.ErrorObjectType, vp.ErrorList[0].Object, "for test %d", index)
		assert.Contains(t, vp.ErrorList[0].Error, testCase.Error, "for test %d", index)
	}

}

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

	//nominal case test should have no errors
	vp := &validation.ValidationProcess{}
	config := baseConfig
	//validation
	vp.Validate(config)
	//checks
	assert.Len(t, vp.ErrorList, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &validation.ValidationProcess{}
		config := baseConfig
		testCase.Mutator(&config)
		//validation
		vp.Validate(config)
		//checks
		assert.Len(t, vp.ErrorList, 1, "for test %d", index)
		assert.IsType(t, testCase.ErrorObjectType, vp.ErrorList[0].Object, "for test %d", index)
		assert.Contains(t, vp.ErrorList[0].Error, testCase.Error, "for test %d", index)
	}

}

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

	//nominal case test should have no errors
	vp := &validation.ValidationProcess{}
	config := baseConfig
	//validation
	vp.Validate(config)
	//checks
	assert.Len(t, vp.ErrorList, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &validation.ValidationProcess{}
		config := deepCopyForTesting(baseConfig).(MailAccountConfig)

		testCase.Mutator(&config)
		//validation
		vp.Validate(config)
		//checks
		assert.Len(t, vp.ErrorList, 1, "for test %d", index)
		assert.IsType(t, testCase.ErrorObjectType, vp.ErrorList[0].Object, "for test %d", index)
		assert.Contains(t, vp.ErrorList[0].Error, testCase.Error, "for test %d", index)
	}

}

func TestSendLimitValidation(t *testing.T) {

	type TestCase struct {
		Mutator func(c *SendLimit)
		Error   string
	}

	testCases := []TestCase{
		{
			Mutator: func(c *SendLimit) { c.AccountNames = nil },
			Error:   "send limits account name list must not be empty",
		},
		{
			Mutator: func(c *SendLimit) { c.AccountNames = []string{} },
			Error:   "send limits account name list must not be empty",
		},
		{
			Mutator: func(c *SendLimit) { c.MinPeriodMinutes = 0 },
			Error:   "send limit min_period_minutes must be non-negative, not '0'",
		},
	}

	baseConfig := SendLimit{
		MinPeriodMinutes: 15,
		AccountNames:     []string{"test1", "test2"},
	}

	//nominal case test should have no errors
	vp := &validation.ValidationProcess{}
	config := baseConfig
	//validation
	vp.Validate(config)
	//checks
	assert.Len(t, vp.ErrorList, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &validation.ValidationProcess{}
		config := baseConfig
		testCase.Mutator(&config)
		//validation
		vp.Validate(config)
		//checks
		assert.Len(t, vp.ErrorList, 1, "for test %d", index)
		assert.IsType(t, SendLimit{}, vp.ErrorList[0].Object, "for test %d", index)
		assert.Contains(t, vp.ErrorList[0].Error, testCase.Error, "for test %d", index)
	}

}

func TestMailConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator         func(c *MailConfig)
		Error           string
		ErrorObjectType interface{}
	}

	testCases := []TestCase{
		{
			Mutator:         func(c *MailConfig) { c.SendLimits[0].AccountNames[0] = "nonexistent" },
			Error:           "send_limits account name 'nonexistent' that does not exist",
			ErrorObjectType: SendLimit{},
		},
		{
			Mutator:         func(c *MailConfig) { c.Accounts[1].Name = "test1" },
			Error:           "duplicate account name 'test1'",
			ErrorObjectType: MailConfig{},
		},
		//pass one error through to each of the MailConfig and SendLimit structs to check end to end behavior
		{
			Mutator:         func(c *MailConfig) { c.Accounts[0].SMTP.Username = "" },
			Error:           "username must not be empty or whitespace",
			ErrorObjectType: SMTPConfig{},
		},
		{
			Mutator:         func(c *MailConfig) { c.SendLimits[0].MinPeriodMinutes = 0 },
			Error:           "send limit min_period_minutes must be non-negative, not '0'",
			ErrorObjectType: SendLimit{},
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
		SendLimits: []SendLimit{
			{
				MinPeriodMinutes: 15,
				AccountNames:     []string{"test1"},
			},
		},
	}

	//nominal case test should have no errors
	vp := &validation.ValidationProcess{}
	//validation
	vp.Validate(baseConfig)
	//checks
	assert.Len(t, vp.ErrorList, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &validation.ValidationProcess{}
		config := deepCopyForTesting(baseConfig).(MailConfig)
		testCase.Mutator(&config)
		//validation
		vp.Validate(config)
		//checks
		assert.Len(t, vp.ErrorList, 1, "for test %d", index)
		assert.IsType(t, testCase.ErrorObjectType, vp.ErrorList[0].Object, "for test %d", index)
		assert.Contains(t, vp.ErrorList[0].Error, testCase.Error, "for test %d", index)
	}

}

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
		SendLimits: []SendLimit{
			{
				MinPeriodMinutes: 15,
				AccountNames:     []string{"test1"},
			},
		},
	}

	assert.Equal(t, &config.Accounts[0], config.GetAccountByName("test1"))
	assert.Equal(t, &config.Accounts[1], config.GetAccountByName("test2"))
	assert.Nil(t, config.GetAccountByName("nonexistent"))

}
