package config

import (
	"bytes"
	"fmt"
	"os"
	"varanus/internal/validation"

	"gopkg.in/yaml.v3"
)

//**************************************************************************************************
//** VaranusConfig object
//**************************************************************************************************

type VaranusConfig struct {
	Mail             MailConfig          `yaml:"mail"`
	MonitoringConfig MonitorConfig       `yaml:"monitoring"`
	ForceFailure     *ForceConfigFailure `yaml:"force_failure,omitempty"`
}

// ToYAML marshalls the config to a YAML format and returns it as a string.
func (c *VaranusConfig) ToYAML() (string, error) {
	return objectToYaml(c)
}

// WriteConfigToFile marshals the VaranusConfig to a YAML format and writes it to filename.  If
// forceOverwrite is False, then an error will occur if the file already exists.
func (c *VaranusConfig) WriteConfigToFile(filename string, forceOverwrite bool) error {
	return writeObjectToFile(c, filename, forceOverwrite)
}

func (c VaranusConfig) Validate(vet validation.ValidationErrorTracker, root interface{}) error {
	//no top-level validation to do
	return nil
}

//**************************************************************************************************
//** Top level methods for loading and saving configs
//**************************************************************************************************

// ReadConfigFromFile creates a VaranusConfig object from the file at filename.
func ReadConfigFromFile(filename string) (*VaranusConfig, error) {

	ydata, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("file read error for config file %s: %s", filename, err)
	}

	config, err := ReadConfig(ydata)
	if err != nil {
		return nil, fmt.Errorf("error with the contents of %s: %s", filename, err)
	}

	return config, nil
}

func ReadConfig(yamlData []byte) (*VaranusConfig, error) {

	config := &VaranusConfig{}

	//use a decoder so we can set KnownFields
	decoder := yaml.NewDecoder(bytes.NewBuffer(yamlData))
	decoder.KnownFields(true)
	err := decoder.Decode(config)

	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return config, nil
}

//**************************************************************************************************
//** helper methods for loading and saving configs
//**************************************************************************************************

func objectToYaml(object interface{}) (string, error) {
	var yamlData bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&yamlData)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(object)
	if err != nil {
		return "", err
	}

	return yamlData.String(), nil
}

func writeObjectToFile(object interface{}, filename string, forceOverwrite bool) error {
	//write the config file back out
	var flags int
	if forceOverwrite {
		//okay to overwrite an existing file
		flags = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	} else {
		flags = os.O_RDWR | os.O_CREATE | os.O_EXCL
	}

	//convert to yaml
	yaml, err := objectToYaml(object)
	if err != nil {
		return fmt.Errorf("marshaling error: %w", err)
	}

	f, err := os.OpenFile(filename, flags, 0600)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", filename, err)
	}

	//write out file
	_, err = f.WriteString(yaml)
	if err != nil {
		//not tested because difficult to induce the write failure
		return fmt.Errorf("error writing to file %s: %w", filename, err)
	}

	return nil
}

//**************************************************************************************************
//** other helper methods used by the rest of the configs
//**************************************************************************************************

func castInterfaceToVaranusConfig(config interface{}) VaranusConfig {
	{
		configValue, ok := config.(VaranusConfig)
		if ok {
			return configValue
		}
	}
	{
		configValue, ok := config.(*VaranusConfig)
		if ok {
			return *configValue
		}
	}
	panic(fmt.Errorf("could not cast %#v to VaranusConfig or *VaranusConfig", config))
}
