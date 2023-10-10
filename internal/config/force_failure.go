package config

import (
	"fmt"
	"os"
	"strings"
	"varanus/internal/secrets"
	"varanus/internal/validation"

	"gopkg.in/yaml.v3"
)

//constants for ForceConfigFailure

// FORCE_CONFIG_FAILURE_YAML_UNMARSHAL is used to cause an error during YAML unmarshalling
const FORCE_CONFIG_FAILURE_YAML_MARSHAL = "yaml_marshal_fails"

// FORCE_CONFIG_FAILURE_YAML_UNMARSHAL is used to cause an error during YAML unmarshalling
const FORCE_CONFIG_FAILURE_YAML_UNMARSHAL = "yaml_unmarshal_fails"

// FORCE_CONFIG_FAILURE_VALIDATION_FAILURE is used to cause an error during validation
const FORCE_CONFIG_FAILURE_VALIDATION_FAILURE = "validation_fails"

// FORCE_CONFIG_FAILURE_SEAL_CHECK_FAILURE is used to cause an error during seal checking
const FORCE_CONFIG_FAILURE_SEAL_CHECK_FAILURE = "seal_check_fails"

// FORCE_CONFIG_FAILURE_SEAL_FAILURE is used to cause an error during sealing
const FORCE_CONFIG_FAILURE_SEAL_FAILURE = "seal_fails"

// FORCE_CONFIG_FAILURE_UNSEAL_FAILURE is used to cause an error during unsealing
const FORCE_CONFIG_FAILURE_UNSEAL_FAILURE = "unseal_fails"

// FORCE_CONFIG_FAILURE_IS_SEALED is used to cause the value to appear sealed
const FORCE_CONFIG_FAILURE_IS_SEALED = "is_sealed"

// ForceConfigFailure, if set, will cause various failures in the configuration
type ForceConfigFailure struct {
	Value string
}

func (fcf ForceConfigFailure) warn() {
	fmt.Fprintf(os.Stderr, "Warning:  the force_failure config value is in use -- this should only be used for testing.\n")
}

func (fcf ForceConfigFailure) MarshalYAML() (interface{}, error) {
	fcf.warn()

	if fcf.Value == FORCE_CONFIG_FAILURE_YAML_MARSHAL {
		return nil, fmt.Errorf("intentional error from force_failure: %s", fcf.Value)
	}
	//otherwise marshal the string as normal
	return fcf.Value, nil
}

func (fcf *ForceConfigFailure) UnmarshalYAML(value *yaml.Node) error {
	fcf.warn()

	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected a scalar value")
	}
	fcf.Value = value.Value
	if strings.Contains(fcf.Value, FORCE_CONFIG_FAILURE_YAML_UNMARSHAL) {
		return fmt.Errorf("intentional error from force_failure: %s", fcf.Value)
	}
	return nil
}

func (fcf ForceConfigFailure) Validate(vet validation.ValidationErrorTracker, root interface{}) error {
	fcf.warn()

	if strings.Contains(fcf.Value, FORCE_CONFIG_FAILURE_VALIDATION_FAILURE) {
		return fmt.Errorf("intentional error from force_failure: %s", fcf.Value)
	}
	return nil
}

func (fcf ForceConfigFailure) IsValueSealed() bool {
	fcf.warn()
	return strings.Contains(fcf.Value, FORCE_CONFIG_FAILURE_IS_SEALED)
}
func (fcf ForceConfigFailure) GetValue() string {
	fcf.warn()
	if strings.Contains(fcf.Value, FORCE_CONFIG_FAILURE_IS_SEALED) {
		return "sealed(AAAA=)"
	} else {
		return "force_config_failure_value"
	}
}
func (fcf ForceConfigFailure) Check(unsealer secrets.SecretUnsealer) error {
	fcf.warn()

	if strings.Contains(fcf.Value, FORCE_CONFIG_FAILURE_SEAL_CHECK_FAILURE) {
		return fmt.Errorf("intentional error from force_failure: %s", fcf.Value)
	}
	return nil
}
func (fcf *ForceConfigFailure) Seal(sealer secrets.SecretSealer) error {
	fcf.warn()

	if strings.Contains(fcf.Value, FORCE_CONFIG_FAILURE_SEAL_FAILURE) {
		return fmt.Errorf("intentional error from force_failure: %s", fcf.Value)
	}
	return nil
}
func (fcf *ForceConfigFailure) Unseal(unsealer secrets.SecretUnsealer) error {
	fcf.warn()

	if strings.Contains(fcf.Value, FORCE_CONFIG_FAILURE_UNSEAL_FAILURE) {
		return fmt.Errorf("intentional error from force_failure: %s", fcf.Value)
	}
	return nil
}
