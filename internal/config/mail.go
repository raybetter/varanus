package config

const KeyMailConfiguration = "mail_configuration"

type SMTPConfig struct {
	SenderAddress string `yaml:"sender_address"`
	ServerAddress string `yaml:"server_address"`
	Port          int
	Username      string
	Password      string
}

type IMAPConfig struct {
	ServerAddress string `yaml:"server_address"`
	Port          int
	Username      string
	Password      string
}

type ServerConfig struct {
	Name string
	SMTP *SMTPConfig
	IMAP *IMAPConfig
}

func (c *ServerConfig) Validate() ([]ValidationError, error) {
	errors := make([]ValidationError, 0)

	//validate fields
	if len(c.Accounts) == 0 {
		addValidationError(errors,
			c,
			"send limits account lists must not be empty", []interface{}{},
		)
	}

	if c.SendLimit <= 0 {
		addValidationError(errors,
			c,
			"send limit values must be non-negative, not '%d'", c.SendLimit,
		)
	}

	return errors, nil
}

type SendLimit struct {
	SendLimit int
	Accounts  []string
}

func (c *SendLimit) Validate() ([]ValidationError, error) {
	errors := make([]ValidationError, 0)

	if len(c.Accounts) == 0 {
		addValidationError(errors,
			c,
			"send limits account lists must not be empty", []interface{}{},
		)
	}

	if c.SendLimit <= 0 {
		addValidationError(errors,
			c,
			"send limit values must be non-negative, not '%d'", c.SendLimit,
		)
	}

	return errors, nil
}

type MailConfig struct {
	Accounts   []ServerConfig
	SendLimits []SendLimit
}

func (c *MailConfig) GetAccountByName(name string) *ServerConfig {
	for _, account := range c.Accounts {
		if account.Name == name {
			return &account
		}
	}
	return nil
}

func (c *MailConfig) Validate() ([]ValidationError, error) {
	errors := make([]ValidationError, 0)

	//validate individual structs
	sub_errors, err := c.Accounts.Validate()
	if err != nil {
		return []ValidationError{}, err
	}
	errors = append(errors, sub_errors...)

	sub_errors, err := c.SendLimits.Validate()
	if err != nil {
		return []ValidationError{}, err
	}
	errors = append(errors, sub_errors...)

	//do cross struct-validation

	//make sure all sendlimit accounts are named accounts
	for _, sendLimit := range c.SendLimits {
		for _, sendLimitAccountName := range sendLimit.Accounts {
			account := c.GetAccountByName(sendLimitAccountName)
			if account == nil {
				addValidationError(
					errors,
					sendLimit,
					"SendLimit references account %s that does not exist", sendLimitAccountName)
			} // else valid account found
		}
	}

	return errors, nil
}
