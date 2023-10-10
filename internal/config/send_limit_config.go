package config

import (
	"time"
	"varanus/internal/validation"
)

type SendLimitConfig struct {
	MinPeriod    time.Duration `yaml:"min_period"`
	AccountNames []string      `yaml:"account_names"`
}

func (c SendLimitConfig) Validate(vet validation.ValidationErrorTracker, root interface{}) error {

	if len(c.AccountNames) == 0 {
		vet.AddValidationError(
			c,
			"send limits account name list must not be empty",
		)
	}

	if c.MinPeriod.Nanoseconds() <= 0 {
		vet.AddValidationError(
			c,
			"send limit min_period must be non-negative, not '%d'", c.MinPeriod,
		)
	}

	//make sure the account name exists
	rootConfig := castInterfaceToVaranusConfig(root)
	for _, sendLimitAccountName := range c.AccountNames {
		account := rootConfig.Mail.GetAccountByName(sendLimitAccountName)
		if account == nil {
			vet.AddValidationError(
				c,
				"send_limits account name '%s' that does not exist", sendLimitAccountName)
		} // else valid account found
	}
	return nil
}
