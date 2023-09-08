package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func ReadConfig(filename string) (*VaranusConfig, error) {

	yfile, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("file read error for config file %s: %s", filename, err)
	}

	config, err := parseConfig(yfile)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config file %s: %s", filename, err)
	}

	return config, nil
}

func parseConfig(yamlData []byte) (*VaranusConfig, error) {

	config := &VaranusConfig{}

	err := yaml.Unmarshal(yamlData, config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	validationErrors, err := config.Validate()

	return config, nil
}

type VaranusConfig struct {
	Mail MailConfig `yaml:"mail"`
}

func (c *VaranusConfig) Validate() ([]ValidationError, error) {
	errors := make([]ValidationError, 0)

	sub_errors, err := c.Mail.Validate()
	if err != nil {
		return []ValidationError{}, err
	}
	errors = append(errors, sub_errors...)

	return errors, nil
}
