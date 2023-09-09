package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSmtpConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator func(c *SMTPConfig)
		Error   string
	}

	testCases := []TestCase{
		{
			Mutator: func(c *SMTPConfig) { c.SenderAddress = "Not an email address" },
			Error:   "sender_address 'Not an email address' is not a valid email",
		},
		{
			Mutator: func(c *SMTPConfig) { c.SenderAddress = "  " },
			Error:   "sender_address '' is not a valid email",
		},
		{
			Mutator: func(c *SMTPConfig) { c.ServerAddress = "not/a/valid/hostname" },
			Error:   "server_address 'not/a/valid/hostname' is not a valid hostname",
		},
		{
			Mutator: func(c *SMTPConfig) { c.Port = 0 },
			Error:   "port value is required and cannot be 0",
		},
		{
			Mutator: func(c *SMTPConfig) { c.Username = "" },
			Error:   "username must not be empty or whitespace",
		},
		{
			Mutator: func(c *SMTPConfig) { c.Username = "  " },
			Error:   "username must not be empty or whitespace",
		},
		{
			Mutator: func(c *SMTPConfig) { c.Password = "" },
			Error:   "password must not be empty or whitespace",
		},
		{
			Mutator: func(c *SMTPConfig) { c.Password = "  " },
			Error:   "password must not be empty or whitespace",
		},
	}

	baseConfig := SMTPConfig{
		SenderAddress: "example@example.com",
		ServerAddress: "mail.example.com",
		Port:          465,
		Username:      "joe@example.com",
		Password:      "example_password",
	}

	//nominal case test should have no errors
	vp := &ValidationProcess{}
	config := baseConfig
	//validation
	config.Validate(vp)
	//checks
	assert.Len(t, vp.Errors, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &ValidationProcess{}
		config := baseConfig
		testCase.Mutator(&config)
		//validation
		config.Validate(vp)
		//checks
		assert.Len(t, vp.Errors, 1, "for test %d", index)
		assert.Equal(t, &config, vp.Errors[0].Object, "for test %d", index)
		assert.Contains(t, vp.Errors[0].Error, testCase.Error, "for test %d", index)
	}

}

func TestIMAPConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator func(c *IMAPConfig)
		Error   string
	}

	testCases := []TestCase{
		{
			Mutator: func(c *IMAPConfig) { c.ServerAddress = "not/a/valid/hostname" },
			Error:   "server_address 'not/a/valid/hostname' is not a valid hostname",
		},
		{
			Mutator: func(c *IMAPConfig) { c.Port = 0 },
			Error:   "port value is required and cannot be 0",
		},
		{
			Mutator: func(c *IMAPConfig) { c.Username = "" },
			Error:   "username must not be empty or whitespace",
		},
		{
			Mutator: func(c *IMAPConfig) { c.Username = "  " },
			Error:   "username must not be empty or whitespace",
		},
		{
			Mutator: func(c *IMAPConfig) { c.Password = "" },
			Error:   "password must not be empty or whitespace",
		},
		{
			Mutator: func(c *IMAPConfig) { c.Password = "  " },
			Error:   "password must not be empty or whitespace",
		},
	}

	baseConfig := IMAPConfig{
		ServerAddress: "mail.example.com",
		Port:          993,
		Username:      "joe@example.com",
		Password:      "example_password",
	}

	//nominal case test should have no errors
	vp := &ValidationProcess{}
	config := baseConfig
	//validation
	config.Validate(vp)
	//checks
	assert.Len(t, vp.Errors, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &ValidationProcess{}
		config := baseConfig
		testCase.Mutator(&config)
		//validation
		config.Validate(vp)
		//checks
		assert.Len(t, vp.Errors, 1, "for test %d", index)
		assert.Equal(t, &config, vp.Errors[0].Object, "for test %d", index)
		assert.Contains(t, vp.Errors[0].Error, testCase.Error, "for test %d", index)
	}

}

func TestMailAccountConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator  func(c *MailAccountConfig)
		Error    string
		ErrorObj func(c *MailAccountConfig) interface{}
	}

	testCases := []TestCase{
		{
			Mutator:  func(c *MailAccountConfig) { c.Name = "" },
			Error:    "account names must not be empty or whitespace",
			ErrorObj: func(c *MailAccountConfig) interface{} { return c },
		},
		{
			Mutator:  func(c *MailAccountConfig) { c.SMTP = nil; c.IMAP = nil },
			Error:    "Every server config must specify one of the imap or smtp sections.  They cannot both be empty.",
			ErrorObj: func(c *MailAccountConfig) interface{} { return c },
		},
		//pass one error through to each of the IMAP and SMTP structs to check end to end behavior
		{
			Mutator:  func(c *MailAccountConfig) { c.SMTP.Username = "" },
			Error:    "username must not be empty or whitespace",
			ErrorObj: func(c *MailAccountConfig) interface{} { return c.SMTP },
		},
		{
			Mutator:  func(c *MailAccountConfig) { c.IMAP.Username = "" },
			Error:    "username must not be empty or whitespace",
			ErrorObj: func(c *MailAccountConfig) interface{} { return c.IMAP },
		},
	}

	baseConfig := MailAccountConfig{
		Name: "test1",
		SMTP: &SMTPConfig{
			SenderAddress: "example@example.com",
			ServerAddress: "mail.example.com",
			Port:          465,
			Username:      "joe@example.com",
			Password:      "example_password",
		},
		IMAP: &IMAPConfig{
			ServerAddress: "mail.example.com",
			Port:          993,
			Username:      "joe@example.com",
			Password:      "example_password",
		},
	}

	//nominal case test should have no errors
	vp := &ValidationProcess{}
	config := baseConfig
	//validation
	config.Validate(vp)
	//checks
	assert.Len(t, vp.Errors, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &ValidationProcess{}
		config := deepCopyForTesting(baseConfig).(MailAccountConfig)

		testCase.Mutator(&config)
		//validation
		config.Validate(vp)
		//checks
		assert.Len(t, vp.Errors, 1, "for test %d", index)
		assert.Equal(t, testCase.ErrorObj(&config), vp.Errors[0].Object, "for test %d", index)
		assert.Contains(t, vp.Errors[0].Error, testCase.Error, "for test %d", index)
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
	vp := &ValidationProcess{}
	config := baseConfig
	//validation
	config.Validate(vp)
	//checks
	assert.Len(t, vp.Errors, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &ValidationProcess{}
		config := baseConfig
		testCase.Mutator(&config)
		//validation
		config.Validate(vp)
		//checks
		assert.Len(t, vp.Errors, 1, "for test %d", index)
		assert.Equal(t, &config, vp.Errors[0].Object, "for test %d", index)
		assert.Contains(t, vp.Errors[0].Error, testCase.Error, "for test %d", index)
	}

}

func TestMailConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator  func(c *MailConfig)
		Error    string
		ErrorObj func(c *MailConfig) interface{}
	}

	testCases := []TestCase{
		{
			Mutator:  func(c *MailConfig) { c.SendLimits[0].AccountNames[0] = "nonexistent" },
			Error:    "send_limits account name 'nonexistent' that does not exist",
			ErrorObj: func(c *MailConfig) interface{} { return c.SendLimits[0] },
		},
		{
			Mutator:  func(c *MailConfig) { c.Accounts[1].Name = "test1" },
			Error:    "duplicate account name 'test1'",
			ErrorObj: func(c *MailConfig) interface{} { return c },
		},
		//pass one error through to each of the MailConfig and SendLimit structs to check end to end behavior
		{
			Mutator:  func(c *MailConfig) { c.Accounts[0].SMTP.Username = "" },
			Error:    "username must not be empty or whitespace",
			ErrorObj: func(c *MailConfig) interface{} { return c.Accounts[0].SMTP },
		},
		{
			Mutator:  func(c *MailConfig) { c.SendLimits[0].MinPeriodMinutes = 0 },
			Error:    "send limit min_period_minutes must be non-negative, not '0'",
			ErrorObj: func(c *MailConfig) interface{} { return &c.SendLimits[0] },
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
					Password:      "example_password",
				},
			},
			{
				Name: "test2",
				SMTP: &SMTPConfig{
					SenderAddress: "example@example.com",
					ServerAddress: "mail.example.com",
					Port:          465,
					Username:      "joe@example.com",
					Password:      "example_password",
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
	vp := &ValidationProcess{}
	//validation
	baseConfig.Validate(vp)
	//checks
	assert.Len(t, vp.Errors, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &ValidationProcess{}
		config := deepCopyForTesting(baseConfig).(MailConfig)
		testCase.Mutator(&config)
		//validation
		config.Validate(vp)
		//checks
		assert.Len(t, vp.Errors, 1, "for test %d", index)
		assert.Equal(t, testCase.ErrorObj(&config), vp.Errors[0].Object, "for test %d", index)
		assert.Contains(t, vp.Errors[0].Error, testCase.Error, "for test %d", index)
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
					Password:      "example_password",
				},
			},
			{
				Name: "test2",
				SMTP: &SMTPConfig{
					SenderAddress: "example@example.com",
					ServerAddress: "mail.example.com",
					Port:          465,
					Username:      "joe@example.com",
					Password:      "example_password",
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
