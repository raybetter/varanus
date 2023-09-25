package config

import (
	"bytes"
	"fmt"
	"os"
	"varanus/internal/secrets"
	"varanus/internal/validation"

	"gopkg.in/yaml.v3"
)

func ReadConfig(filename string) (*VaranusConfig, error) {

	yfile, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("file read error for config file %s: %s", filename, err)
	}

	config, err := parseAndValidateConfig(yfile)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config file %s: %s", filename, err)
	}

	return config, nil
}

func (c *VaranusConfig) ToYAML() (string, error) {
	// //marshall config output
	// yamlData, err := yaml.Marshal(c)
	// if err != nil {
	// 	return "", fmt.Errorf("YAML marshalling error: %w", err)
	// }
	// return string(yamlData), nil

	var yamlData bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&yamlData)
	yamlEncoder.SetIndent(2) // this is what you're looking for
	err := yamlEncoder.Encode(c)
	if err != nil {
		return "", err
	}

	return yamlData.String(), nil

}

func (c *VaranusConfig) WriteConfig(filename string, forceOverwrite bool) error {
	//write the config file back out
	var flags int
	if forceOverwrite {
		//okay to overwrite an existing file
		flags = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	} else {
		flags = os.O_RDWR | os.O_CREATE | os.O_EXCL
	}

	f, err := os.OpenFile(filename, flags, 0600)
	if err != nil {
		return fmt.Errorf("could not open file")
	}

	//convert to yaml
	yaml, err := c.ToYAML()
	if err != nil {
		return err
	}

	//write out file
	f.WriteString(yaml)

	return nil

}

func parseAndValidateConfig(yamlData []byte) (*VaranusConfig, error) {

	config := &VaranusConfig{}

	err := yaml.Unmarshal(yamlData, config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	vp := validation.ValidationProcess{}
	err = vp.Validate(config)
	if err != nil {
		//can't cover with test because none of the config validation has an induceable error
		return nil, fmt.Errorf("config validation failed to complete: %w", err)
	}

	err = vp.GetFinalValidationError()
	if err != nil {
		return nil, err
	}

	return config, nil
}

type VaranusConfig struct {
	Mail MailConfig `yaml:"mail"`
}

func (c *VaranusConfig) Validate(vp *validation.ValidationProcess) error {

	//validate struct members
	err := vp.Validate(&c.Mail)
	if err != nil {
		//can't cover with test because none of the config validation has an induceable error
		return err
	}

	return nil
}

func (c *VaranusConfig) Seal(sealer secrets.SecretSealer) error {
	return AddTokenToPathError(c.Mail.Seal(sealer), "mail")
}

func (c *VaranusConfig) Unseal(unsealer secrets.SecretUnsealer) error {
	return AddTokenToPathError(c.Mail.Unseal(unsealer), "mail")
}

func (c *VaranusConfig) CheckSeals(unsealer secrets.SecretUnsealer) secrets.SealCheckResult {
	return c.Mail.CheckSeals(unsealer)
}
