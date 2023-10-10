package config

import (
	"testing"
	"varanus/internal/validation"
)

func TestMailConfigsAreValidatable(t *testing.T) {
	vf := func(validatable validation.Validatable) {}

	//these won't compile if we've failed to implement validatable
	vf(EmailMonitorConfig{})
	vf(ForceConfigFailure{})
	vf(IMAPConfig{})
	vf(MailAccountConfig{})
	vf(MailConfig{})
	vf(MonitorConfig{})
	vf(NotificationConfig{})
	vf(SendLimitConfig{})
	vf(SMTPConfig{})
	vf(VaranusConfig{})

}
