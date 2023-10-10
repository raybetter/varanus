package config

import (
	"testing"
	"varanus/internal/secrets"
	"varanus/internal/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestForceFailureNoFailure(t *testing.T) {

	//nothing in this test should fail because the force_failure value doesn't match a defined value

	inputYaml := `mail:
  accounts: []
  send_limits: []
monitoring:
  email_monitors: []
force_failure: some value
`
	c, err := ReadConfig([]byte(inputYaml))
	assert.Nil(t, err)

	assert.NotNil(t, c.ForceFailure)
	assert.Equal(t, "some value", c.ForceFailure.Value)

	validationResult, err := validation.ValidateObject(c)
	assert.Nil(t, err)
	assert.Equal(t, 0, validationResult.GetErrorCount())

	assert.False(t, c.ForceFailure.IsValueSealed())
	assert.Equal(t, "force_config_failure_value", c.ForceFailure.GetValue())

	sealResult := secrets.SealObject(c, nil)
	assert.Equal(t, 1, sealResult.NumberSealed)
	assert.Len(t, sealResult.SealErrors, 0)

	outputYaml, err := c.ToYAML()
	assert.Nil(t, err)

	assert.Equal(t, inputYaml, outputYaml)
}

func TestForceFailureSealed(t *testing.T) {

	inputYaml := `mail:
  accounts: []
  send_limits: []
monitoring:
  email_monitors: []
force_failure: is_sealed
`
	c, err := ReadConfig([]byte(inputYaml))
	assert.Nil(t, err)

	assert.True(t, c.ForceFailure.IsValueSealed())
	assert.Equal(t, "sealed(AAAA=)", c.ForceFailure.GetValue())

	sealCheckResult := secrets.CheckSealsOnObject(c, nil)
	assert.Equal(t, 1, sealCheckResult.SealedCount)
	assert.Equal(t, 0, sealCheckResult.UnsealedCount)
	assert.Len(t, sealCheckResult.UnsealErrors, 0)

	unsealResult := secrets.UnsealObject(c, nil)
	assert.Equal(t, 1, unsealResult.NumberUnsealed)
	assert.Len(t, unsealResult.UnsealErrors, 0)

}

func TestForceFailureYamlUnmarshal(t *testing.T) {

	inputYaml := `mail:
  accounts: []
  send_limits: []
monitoring:
  email_monitors: []
force_failure: yaml_unmarshal_fails
`
	c, err := ReadConfig([]byte(inputYaml))
	assert.ErrorContains(t, err, "intentional error from force_failure: yaml_unmarshal_fails")
	assert.Nil(t, c)
}

func TestForceFailureMarshal(t *testing.T) {

	c := &VaranusConfig{
		ForceFailure: &ForceConfigFailure{FORCE_CONFIG_FAILURE_YAML_MARSHAL},
	}

	outputYaml, err := c.ToYAML()
	assert.Equal(t, "", outputYaml)
	assert.ErrorContains(t, err, "intentional error from force_failure: yaml_marshal_fails")
}

func TestForceFailureValidate(t *testing.T) {

	c := &VaranusConfig{
		ForceFailure: &ForceConfigFailure{FORCE_CONFIG_FAILURE_VALIDATION_FAILURE},
	}

	validationResult, err := validation.ValidateObject(c)
	assert.ErrorContains(t, err, "intentional error from force_failure: validation_fails")
	assert.Equal(t, validation.ValidationResult{}, validationResult)
}

func TestYamlUnmarshalingStructureError(t *testing.T) {
	fcf := ForceConfigFailure{"some value"}

	yamlNode := yaml.Node{
		Kind:  yaml.SequenceNode, //wrong node type
		Value: "doesn't matter",
	}

	err := fcf.UnmarshalYAML(&yamlNode)
	assert.ErrorContains(t, err, "expected a scalar value")
}

func TestForceFailureSealCheck(t *testing.T) {

	c := &VaranusConfig{
		ForceFailure: &ForceConfigFailure{
			FORCE_CONFIG_FAILURE_SEAL_CHECK_FAILURE + "|" + FORCE_CONFIG_FAILURE_IS_SEALED,
		},
	}

	sealCheckResult := secrets.CheckSealsOnObject(c, nil)
	assert.Equal(t, 1, sealCheckResult.SealedCount)
	assert.Equal(t, 0, sealCheckResult.UnsealedCount)
	require.Len(t, sealCheckResult.UnsealErrors, 1)
	assert.ErrorContains(t, sealCheckResult.UnsealErrors[0], "intentional error from force_failure: seal_check_fails|is_sealed")

}

func TestForceFailureUnseal(t *testing.T) {

	c := &VaranusConfig{
		ForceFailure: &ForceConfigFailure{
			FORCE_CONFIG_FAILURE_UNSEAL_FAILURE + "|" + FORCE_CONFIG_FAILURE_IS_SEALED,
		},
	}

	unsealResult := secrets.UnsealObject(c, nil)
	assert.Equal(t, 0, unsealResult.NumberUnsealed)
	require.Len(t, unsealResult.UnsealErrors, 1)
	assert.ErrorContains(t, unsealResult.UnsealErrors[0], "intentional error from force_failure: unseal_fails|is_sealed")
}

func TestForceFailureSeal(t *testing.T) {

	c := &VaranusConfig{
		ForceFailure: &ForceConfigFailure{
			FORCE_CONFIG_FAILURE_SEAL_FAILURE,
		},
	}

	sealResult := secrets.SealObject(c, nil)
	assert.Equal(t, 0, sealResult.NumberSealed)
	require.Len(t, sealResult.SealErrors, 1)
	assert.ErrorContains(t, sealResult.SealErrors[0], "intentional error from force_failure: seal_fails")
}
