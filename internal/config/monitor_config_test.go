package config

import (
	"testing"
	"varanus/internal/validation"

	"github.com/stretchr/testify/assert"
)

func TestMonitorConfigValidation(t *testing.T) {
	vet := validation.ValidationResult{}
	mc := MonitorConfig{}

	err := mc.Validate(&vet, nil)
	assert.Nil(t, err)
}
