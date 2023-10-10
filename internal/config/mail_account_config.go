package config

import (
	"strings"
	"varanus/internal/validation"
)

type MailAccountConfig struct {
	Name string
	SMTP *SMTPConfig `yaml:",omitempty"`
	IMAP *IMAPConfig `yaml:",omitempty"`
}

func (c MailAccountConfig) Validate(vet validation.ValidationErrorTracker, root interface{}) error {

	//validate fields
	c.Name = strings.TrimSpace(c.Name)
	if len(c.Name) == 0 {
		vet.AddValidationError(
			c,
			"account names must not be empty or whitespace",
		)
	}

	//validate ServerConfig level logic

	//SMTP and IMAP cannot both be empty
	if c.SMTP == nil && c.IMAP == nil {
		vet.AddValidationError(c,
			"Every server config must specify one of the imap or smtp sections.  They cannot both be empty.")
	}

	return nil
}
