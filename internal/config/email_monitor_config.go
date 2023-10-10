package config

import (
	"time"
	"varanus/internal/validation"
)

type EmailMonitorConfig struct {
	FromAccount   string               `yaml:"from_account"`
	ToAccount     string               `yaml:"to_account"`
	TestPeriod    time.Duration        `yaml:"test_period"`
	Notifications []NotificationConfig `yaml:"notifications"`
}

func (c EmailMonitorConfig) Validate(vet validation.ValidationErrorTracker, root interface{}) error {

	vConfig := castInterfaceToVaranusConfig(root)

	if len(c.FromAccount) == 0 {
		vet.AddValidationError(
			c,
			"from_account must not be empty",
		)
	} else { //else avoids a double validation error from an empty string
		account := vConfig.Mail.GetAccountByName(c.FromAccount)
		if account == nil {
			vet.AddValidationError(
				c,
				"from_account named '%s' does not exist", c.FromAccount,
			)
		} else {
			//only check for SMTP if we can get the account
			if account.SMTP == nil {
				vet.AddValidationError(
					c,
					"from_account named '%s' must have an SMTP configuration", c.FromAccount,
				)
			}

		}
	}

	if len(c.ToAccount) == 0 {
		vet.AddValidationError(
			c,
			"to_account must not be empty",
		)
	} else { //else avoids a double validation error from an empty string
		account := vConfig.Mail.GetAccountByName(c.ToAccount)
		if account == nil {
			vet.AddValidationError(
				c,
				"to_account named '%s' does not exist", c.ToAccount,
			)
		} else {
			//only check for IMAP if we can get the account
			if account.IMAP == nil {
				vet.AddValidationError(
					c,
					"to_account named '%s' must have an IMAP configuration", c.ToAccount,
				)
			}
		}
	}

	if c.TestPeriod.Nanoseconds() <= 0 {
		vet.AddValidationError(
			c,
			"test_period must be a positive value, not '%d'", c.TestPeriod,
		)
	}

	if len(c.Notifications) == 0 {
		vet.AddValidationError(
			c,
			"the list of notifications is empty. Each monitor must have at least on notification defined",
		)
	}

	if len(c.Notifications) == 1 && c.Notifications[0].Mail == c.ToAccount {
		vet.AddValidationError(
			c,
			"the only notification for a monitor cannot be the to_account since this is one of the accounts being tested",
		)
	}

	if len(c.Notifications) == 1 && c.Notifications[0].Mail == c.FromAccount {
		vet.AddValidationError(
			c,
			"the only notification for a monitor cannot be the from_account since this is one of the accounts being tested",
		)
	}

	return nil
}
