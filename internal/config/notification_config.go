package config

import "varanus/internal/validation"

type NotificationConfig struct {
	Mail string `yaml:"mail"`
	//TODO add more notification methods when we create more account types
}

func (c NotificationConfig) Validate(vet validation.ValidationErrorTracker, root interface{}) error {

	vConfig := castInterfaceToVaranusConfig(root)

	if len(c.Mail) == 0 {
		vet.AddValidationError(
			c,
			"mail entry must not be empty",
		)
	} else { //else avoids a double validation error from an empty string
		if vConfig.Mail.GetAccountByName(c.Mail) == nil {
			vet.AddValidationError(
				c,
				"notification mail account named '%s' does not exist", c.Mail,
			)
		}

	}

	return nil
}
