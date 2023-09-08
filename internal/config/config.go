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

	config, err := parseAndValidateConfig(yfile)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config file %s: %s", filename, err)
	}

	return config, nil
}

func parseAndValidateConfig(yamlData []byte) (*VaranusConfig, error) {

	config := &VaranusConfig{}

	err := yaml.Unmarshal(yamlData, config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	vp := ValidationProcess{}
	err = vp.Validate(config)
	if err != nil {
		return nil, fmt.Errorf("config validation failed to complete: %w", err)
	}

	err = vp.Finalize()
	if err != nil {
		return nil, err
	}

	return config, nil
}

type VaranusConfig struct {
	Mail MailConfig `yaml:"mail"`
}

func (c *VaranusConfig) Validate(vp *ValidationProcess) error {

	//validate struct members
	err := vp.Validate(&c.Mail)
	if err != nil {
		return err
	}

	return nil
}
