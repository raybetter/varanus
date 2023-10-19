package config

import (
	"net/mail"

	"strings"
	"varanus/internal/secrets"
	"varanus/internal/util"
	"varanus/internal/validation"
)

type IMAPConfig struct {
	RecipientAddress string             `yaml:"recipient_address"`
	ServerAddress    string             `yaml:"server_address"`
	Port             uint               `yaml:"port"`
	UseTLS           bool               `yaml:"use_tls"`
	Username         string             `yaml:"username"`
	Password         secrets.SealedItem `yaml:"password"`
	MailboxName      string             `yaml:"mailbox_name"`
}

func (c IMAPConfig) Validate(vet validation.ValidationErrorTracker, root interface{}) error {

	_, addressError := mail.ParseAddress(c.RecipientAddress)
	if addressError != nil {
		vet.AddValidationError(
			c,
			"recipient_address '%s' is not a valid email: %s", c.RecipientAddress, addressError,
		)
	}

	c.ServerAddress = strings.TrimSpace(c.ServerAddress)
	if !util.IsUrlHost(c.ServerAddress) {
		vet.AddValidationError(
			c,
			"server_address '%s' is not a valid hostname", c.ServerAddress,
		)
	}

	if c.Port == 0 {
		vet.AddValidationError(
			c,
			"port value is required and cannot be 0",
		)
	}

	c.Username = strings.TrimSpace(c.Username)
	if len(c.Username) == 0 {
		vet.AddValidationError(
			c,
			"username must not be empty or whitespace",
		)
	}

	c.MailboxName = strings.TrimSpace(c.MailboxName)
	if len(c.MailboxName) == 0 {
		vet.AddValidationError(
			c,
			"mailbox_name must not be empty or whitespace",
		)
	}

	//password will be validated on its own because it is also Validatable

	return nil
}
