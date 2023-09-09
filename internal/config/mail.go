package config

import (
	"net/mail"
	"strings"
)

const KeyMailConfiguration = "mail_configuration"

type SMTPConfig struct {
	SenderAddress string `yaml:"sender_address"`
	ServerAddress string `yaml:"server_address"`
	Port          uint
	Username      string
	Password      string
}

func (c *SMTPConfig) Validate(vp *ValidationProcess) error {

	//validate fields
	c.SenderAddress = strings.TrimSpace(c.SenderAddress)

	_, addressError := mail.ParseAddress(c.SenderAddress)
	if addressError != nil {
		vp.addValidationError(
			c,
			"sender_address '%s' is not a valid email: %s", c.SenderAddress, addressError,
		)
	}

	c.ServerAddress = strings.TrimSpace(c.ServerAddress)
	if !IsUrlHost(c.ServerAddress) {
		vp.addValidationError(
			c,
			"server_address '%s' is not a valid hostname", c.ServerAddress,
		)
	}

	if c.Port == 0 {
		vp.addValidationError(
			c,
			"port value is required and cannot be 0",
		)
	}

	c.Username = strings.TrimSpace(c.Username)
	if len(c.Username) == 0 {
		vp.addValidationError(
			c,
			"username must not be empty or whitespace",
		)
	}

	c.Password = strings.TrimSpace(c.Password)
	if len(c.Password) == 0 {
		vp.addValidationError(
			c,
			"password must not be empty or whitespace",
		)
	}

	return nil
}

type IMAPConfig struct {
	ServerAddress string `yaml:"server_address"`
	Port          uint
	Username      string
	Password      string
}

func (c *IMAPConfig) Validate(vp *ValidationProcess) error {

	c.ServerAddress = strings.TrimSpace(c.ServerAddress)
	if !IsUrlHost(c.ServerAddress) {
		vp.addValidationError(
			c,
			"server_address '%s' is not a valid hostname", c.ServerAddress,
		)
	}

	if c.Port == 0 {
		vp.addValidationError(
			c,
			"port value is required and cannot be 0",
		)
	}

	c.Username = strings.TrimSpace(c.Username)
	if len(c.Username) == 0 {
		vp.addValidationError(
			c,
			"username must not be empty or whitespace",
		)
	}

	c.Password = strings.TrimSpace(c.Password)
	if len(c.Password) == 0 {
		vp.addValidationError(
			c,
			"password must not be empty or whitespace",
		)
	}

	return nil
}

type MailAccountConfig struct {
	Name string
	SMTP *SMTPConfig
	IMAP *IMAPConfig
}

func (c *MailAccountConfig) Validate(vp *ValidationProcess) error {

	//validate fields
	c.Name = strings.TrimSpace(c.Name)
	if len(c.Name) == 0 {
		vp.addValidationError(
			c,
			"account names must not be empty or whitespace",
		)
	}

	//validate ServerConfig level logic

	//SMTP and IMAP cannot both be empty
	if c.SMTP == nil && c.IMAP == nil {
		vp.addValidationError(c,
			"Every server config must specify one of the imap or smtp sections.  They cannot both be empty.")
	}

	//validate sub structs
	if c.SMTP != nil {
		err := vp.Validate(c.SMTP)
		if err != nil {
			return err
		}
	}

	if c.IMAP != nil {
		err := vp.Validate(c.IMAP)
		if err != nil {
			return err
		}
	}

	return nil
}

type SendLimit struct {
	MinPeriodMinutes int      `yaml:"min_period_minutes"`
	AccountNames     []string `yaml:"account_names"`
}

func (c *SendLimit) Validate(vp *ValidationProcess) error {

	if len(c.AccountNames) == 0 {
		vp.addValidationError(
			c,
			"send limits account name list must not be empty",
		)
	}

	if c.MinPeriodMinutes <= 0 {
		vp.addValidationError(
			c,
			"send limit min_period_minutes must be non-negative, not '%d'", c.MinPeriodMinutes,
		)
	}

	return nil
}

type MailConfig struct {
	Accounts   []MailAccountConfig
	SendLimits []SendLimit `yaml:"send_limits"`
}

func (c *MailConfig) GetAccountByName(name string) *MailAccountConfig {
	for _, account := range c.Accounts {
		if account.Name == name {
			return &account
		}
	}
	return nil
}

func (c *MailConfig) Validate(vp *ValidationProcess) error {

	//validate individual struct members
	for _, account := range c.Accounts {
		err := vp.Validate(&account)
		if err != nil {
			return err
		}
	}

	for _, sendLimit := range c.SendLimits {
		err := vp.Validate(&sendLimit)
		if err != nil {
			return err
		}
	}

	//do cross item validation at the level where we have all the information to check

	//make sure all sendlimit accounts are named accounts
	for _, sendLimit := range c.SendLimits {
		for _, sendLimitAccountName := range sendLimit.AccountNames {
			account := c.GetAccountByName(sendLimitAccountName)
			if account == nil {
				vp.addValidationError(
					sendLimit,
					"send_limits account name '%s' that does not exist", sendLimitAccountName)
			} // else valid account found
		}
	}

	//make sure the account names are unique
	namesInUse := map[string]bool{}
	for _, account := range c.Accounts {
		//validation error if the name is already in use
		if namesInUse[account.Name] {
			vp.addValidationError(
				c,
				"duplicate account name '%s'", account.Name,
			)
		}
		//add the name in use to the map
		namesInUse[account.Name] = true
	}

	return nil
}
