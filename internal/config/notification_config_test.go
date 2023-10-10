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

func TestNotificationConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator func(c *NotificationConfig)
		Error   string
	}

	testCases := []TestCase{
		{
			Mutator: func(c *NotificationConfig) { c.Mail = "" },
			Error:   "mail entry must not be empty",
		},
		{
			Mutator: func(c *NotificationConfig) { c.Mail = "nonexistent" },
			Error:   "notification mail account named 'nonexistent' does not exist",
		},
	}

	baseConfig :=
		VaranusConfig{
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
			Mail: MailConfig{
				Accounts: []MailAccountConfig{
					{
						Name: "test1",
						SMTP: &SMTPConfig{
							SenderAddress: "foo@foo.com",
							ServerAddress: "bar.example.com",
							Port:          465,
							Username:      "user",
							Password:      secrets.CreateSealedItem("password"),
						},
					},
					{
						Name: "test2",
						IMAP: &IMAPConfig{
							ServerAddress: "bar.example.com",
							Port:          993,
							Username:      "user",
							Password:      secrets.CreateSealedItem("password"),
						},
					},
					{
						Name: "test3",
						IMAP: &IMAPConfig{
							ServerAddress: "bar.example.com",
							Port:          993,
							Username:      "user",
							Password:      secrets.CreateSealedItem("password"),
						},
					},
				},
				SendLimits: []SendLimitConfig{},
			},
		}

	{
		//nominal case test should have no errors
		config := util.DeepCopy(baseConfig).(VaranusConfig) //make a copy of the config

		//call validation on the target object
		validationResult := validation.ValidationResult{}
		err := config.MonitoringConfig.EmailMonitors[0].Notifications[0].Validate(&validationResult, config)

		assert.Nil(t, err)
		assert.Equal(t, 0, validationResult.GetErrorCount())
	}
	// test loop
	for index, testCase := range testCases {
		//do validation
		config := util.DeepCopy(baseConfig).(VaranusConfig)                          //make a copy of the config
		testCase.Mutator(&config.MonitoringConfig.EmailMonitors[0].Notifications[0]) //modify the config
		//call validation on the target object
		validationResult := validation.ValidationResult{}
		err := config.MonitoringConfig.EmailMonitors[0].Notifications[0].Validate(&validationResult, config)
		//checks
		assert.Nil(t, err)
		require.Equal(t, validationResult.GetErrorCount(), 1, "for test %d", index)
		singleError := validationResult.GetErrorList()[0]
		assert.IsType(t, NotificationConfig{}, singleError.Object, "for test %d", index)
		assert.Contains(t, singleError.Error, testCase.Error, "for test %d", index)
	}

}
