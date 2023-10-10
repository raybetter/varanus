package config

import (
	"net/mail"
	"strings"
	"varanus/internal/secrets"
	"varanus/internal/util"
	"varanus/internal/validation"
)

type SMTPConfig struct {
	SenderAddress string             `yaml:"sender_address"`
	ServerAddress string             `yaml:"server_address"`
	Port          uint               `yaml:"port"`
	Username      string             `yaml:"username"`
	Password      secrets.SealedItem `yaml:"password"`
}

func (c SMTPConfig) Validate(vet validation.ValidationErrorTracker, root interface{}) error {

	//validate fields
	c.SenderAddress = strings.TrimSpace(c.SenderAddress)

	_, addressError := mail.ParseAddress(c.SenderAddress)
	if addressError != nil {
		vet.AddValidationError(
			c,
			"sender_address '%s' is not a valid email: %s", c.SenderAddress, addressError,
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

	//password will be validated on its own because it is also Validatable

	return nil
}
