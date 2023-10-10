package config

import "varanus/internal/validation"

type MailConfig struct {
	Accounts   []MailAccountConfig `yaml:"accounts"`
	SendLimits []SendLimitConfig   `yaml:"send_limits"`
}

func (c MailConfig) GetAccountByName(name string) *MailAccountConfig {
	for _, account := range c.Accounts {
		if account.Name == name {
			return &account
		}
	}
	return nil
}

func (c MailConfig) Validate(vet validation.ValidationErrorTracker, root interface{}) error {

	//make sure the account names are unique
	namesInUse := map[string]bool{}
	for _, account := range c.Accounts {
		//validation error if the name is already in use
		if namesInUse[account.Name] {
			vet.AddValidationError(
				c,
				"duplicate account name '%s'", account.Name,
			)
		}
		//add the name in use to the map
		namesInUse[account.Name] = true
	}

	return nil
}
