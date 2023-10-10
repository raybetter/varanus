package config

import (
	"testing"
	"time"
	"varanus/internal/secrets"
	"varanus/internal/util"
	"varanus/internal/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailMonitorConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator func(c *EmailMonitorConfig)
		Error   string
	}

	nominalTestCases := []TestCase{
		{
			//base configuration should have no errors
			Mutator: func(c *EmailMonitorConfig) {},
			Error:   "",
		},
		{
			//should accept the from_account with another account
			Mutator: func(c *EmailMonitorConfig) { c.Notifications = []NotificationConfig{{"test1"}, {"test3"}} },
			Error:   "",
		},
		{
			//should accept the to_account with another account
			Mutator: func(c *EmailMonitorConfig) { c.Notifications = []NotificationConfig{{"test2"}, {"test3"}} },
			Error:   "",
		},
		{
			//should accept the from_account and to_account together
			Mutator: func(c *EmailMonitorConfig) { c.Notifications = []NotificationConfig{{"test1"}, {"test2"}} },
			Error:   "",
		},
	}
	errorTestCases := []TestCase{
		{
			Mutator: func(c *EmailMonitorConfig) { c.ToAccount = "" },
			Error:   "to_account must not be empty",
		},
		{
			Mutator: func(c *EmailMonitorConfig) { c.ToAccount = "nonexistent" },
			Error:   "to_account named 'nonexistent' does not exist",
		},
		{
			Mutator: func(c *EmailMonitorConfig) { c.ToAccount = "test1" },
			Error:   "to_account named 'test1' must have an IMAP configuration",
		},
		{
			Mutator: func(c *EmailMonitorConfig) { c.FromAccount = "" },
			Error:   "from_account must not be empty",
		},
		{
			Mutator: func(c *EmailMonitorConfig) { c.FromAccount = "nonexistent" },
			Error:   "from_account named 'nonexistent' does not exist",
		},
		{
			Mutator: func(c *EmailMonitorConfig) { c.FromAccount = "test2" },
			Error:   "from_account named 'test2' must have an SMTP configuration",
		},
		{
			Mutator: func(c *EmailMonitorConfig) { c.TestPeriod = time.Duration(-1) },
			Error:   "test_period must be a positive value, not '-1'",
		},
		{
			Mutator: func(c *EmailMonitorConfig) { c.TestPeriod = time.Duration(0) },
			Error:   "test_period must be a positive value, not '0'",
		},
		{
			Mutator: func(c *EmailMonitorConfig) { c.Notifications = []NotificationConfig{} },
			Error:   "the list of notifications is empty. Each monitor must have at least on notification defined",
		},
		{
			Mutator: func(c *EmailMonitorConfig) { c.Notifications = []NotificationConfig{{"test2"}} },
			Error:   "the only notification for a monitor cannot be the to_account since this is one of the accounts being tested",
		},
		{
			Mutator: func(c *EmailMonitorConfig) { c.Notifications = []NotificationConfig{{"test1"}} },
			Error:   "the only notification for a monitor cannot be the from_account since this is one of the accounts being tested",
		},
	}

	// we construct a whole config object because the EmailMonitorConfig references the accounts list
	// in its validation
	baseConfig :=
		VaranusConfig{
			Mail: MailConfig{
				Accounts: []MailAccountConfig{
					{
						Name: "test1",
						SMTP: &SMTPConfig{
							SenderAddress: "foo1@example.com",
							ServerAddress: "mail.example.com",
							Port:          465,
							Username:      "username1",
							Password:      secrets.CreateSealedItem("password1"),
						},
					},
					{
						Name: "test2",
						IMAP: &IMAPConfig{
							ServerAddress: "mail2.example.com",
							Port:          990,
							Username:      "username2",
							Password:      secrets.CreateSealedItem("password2"),
						},
					},
					{
						Name: "test3",
						SMTP: &SMTPConfig{
							SenderAddress: "foo3@example.com",
							ServerAddress: "mail3.example.com",
							Port:          465,
							Username:      "username3",
							Password:      secrets.CreateSealedItem("password3"),
						},
						IMAP: &IMAPConfig{
							ServerAddress: "mail.example.com",
							Port:          990,
							Username:      "username3",
							Password:      secrets.CreateSealedItem("password3"),
						},
					},
				},
				SendLimits: []SendLimitConfig{},
			},
			MonitoringConfig: MonitorConfig{
				EmailMonitors: []EmailMonitorConfig{
					{
						FromAccount: "test1",
						ToAccount:   "test2",
						TestPeriod:  time.Duration(10) * time.Minute,
						Notifications: []NotificationConfig{
							{
								Mail: "test3",
							},
						},
					},
				},
			},
		}

	//nominal case test should have no errors
	for index, testCase := range nominalTestCases {
		config := util.DeepCopy(baseConfig).(VaranusConfig) //make a copy of the config
		//mutate for the nominal case
		testCase.Mutator(&config.MonitoringConfig.EmailMonitors[0]) //modify the config

		//call validation on the target object
		validationResult := validation.ValidationResult{}
		err := config.MonitoringConfig.EmailMonitors[0].Validate(&validationResult, config)

		assert.Nil(t, err, "for index %d", index)
		assert.Equal(t, 0, validationResult.GetErrorCount(), "for index %d", index)
	}

	// test loop
	for index, testCase := range errorTestCases {
		//do validation
		config := util.DeepCopy(baseConfig).(VaranusConfig)         //make a copy of the config
		testCase.Mutator(&config.MonitoringConfig.EmailMonitors[0]) //modify the config

		//call validation on the target object
		validationResult := validation.ValidationResult{}
		err := config.MonitoringConfig.EmailMonitors[0].Validate(&validationResult, config)
		//checks
		assert.Nil(t, err)
		require.Equal(t, validationResult.GetErrorCount(), 1, "for test %d", index)
		singleError := validationResult.GetErrorList()[0]
		assert.IsType(t, EmailMonitorConfig{}, singleError.Object, "for test %d", index)
		assert.Contains(t, singleError.Error, testCase.Error, "for test %d", index)
	}

}
