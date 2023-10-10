package config

import (
	"varanus/internal/validation"
)

type MonitorConfig struct {
	EmailMonitors []EmailMonitorConfig `yaml:"email_monitors"`
}

func (c MonitorConfig) Validate(vet validation.ValidationErrorTracker, root interface{}) error {

	//nothing to validate at this level
	return nil
}
