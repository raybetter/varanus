package config

import (
	"fmt"
	"os"
	"testing"
	"varanus/internal/util"
	"varanus/internal/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestEndToEnd(t *testing.T) {
	input_filename := "tests/example.yaml"

	c, err := ReadConfigFromFile(input_filename)
	require.Nil(t, err)

	validationResult, err := validation.ValidateObject(c)
	require.Nil(t, err)
	assert.Equal(t, 0, validationResult.GetErrorCount())

	assert.Len(t, c.Mail.Accounts, 1)
	assert.Equal(t, "test1", c.Mail.Accounts[0].Name)
	assert.NotNil(t, c.Mail.Accounts[0].SMTP)
	assert.Equal(t, "example@example.com", c.Mail.Accounts[0].SMTP.SenderAddress)
	assert.Equal(t, "smtp.example.com", c.Mail.Accounts[0].SMTP.ServerAddress)
	assert.Equal(t, uint(465), c.Mail.Accounts[0].SMTP.Port)
	assert.Equal(t, "joeuser@example.com", c.Mail.Accounts[0].SMTP.Username)
	assert.Equal(t, "it's a secret", c.Mail.Accounts[0].SMTP.Password.GetValue())

	assert.NotNil(t, c.Mail.Accounts[0].IMAP)
	assert.Equal(t, "imap.example.com", c.Mail.Accounts[0].IMAP.ServerAddress)
	assert.Equal(t, uint(993), c.Mail.Accounts[0].IMAP.Port)
	assert.Equal(t, "janeuser@example.com", c.Mail.Accounts[0].IMAP.Username)
	assert.Equal(t, "sealed(+bbbbbb==)", c.Mail.Accounts[0].IMAP.Password.GetValue())

	assert.Len(t, c.Mail.SendLimits, 1)
	assert.Equal(t, c.Mail.SendLimits[0].MinPeriodMinutes, 10)
	assert.Equal(t, c.Mail.SendLimits[0].AccountNames, []string{"test1"})

	//write the config back out

	expectedOutput := `mail:
  accounts:
    - name: test1
      smtp:
        sender_address: example@example.com
        server_address: smtp.example.com
        port: 465
        username: joeuser@example.com
        password: it's a secret
      imap:
        server_address: imap.example.com
        port: 993
        username: janeuser@example.com
        password: sealed(+bbbbbb==)
  send_limits:
    - min_period_minutes: 10
      account_names:
        - test1
`

	//check the YAML conversion
	yamlOutput, err := c.ToYAML()
	assert.Equal(t, expectedOutput, yamlOutput)

	//use the temp to get a filename, but close that tempFile because we really just want the filename
	tempFileContents := "not the real config result"
	tempFile := util.CreateTempFileAndDir("test_output", "output.*.yaml")
	_, err = tempFile.WriteString(tempFileContents)
	assert.Nil(t, err)
	err = tempFile.Close()
	assert.Nil(t, err)

	//take advantage of the existing file to test the overwrite
	err = c.WriteConfigToFile(tempFile.Name(), false)
	assert.ErrorContains(t, err, "could not open file")
	assert.ErrorContains(t, err, "file exists")

	//make sure the file was not modified
	outputData, err := os.ReadFile(tempFile.Name())
	assert.Nil(t, err)
	assert.Equal(t, tempFileContents, string(outputData))

	//overwrite the temp file
	err = c.WriteConfigToFile(tempFile.Name(), true)
	assert.Nil(t, err)

	//read back the file and check its contents
	outputData, err = os.ReadFile(tempFile.Name())
	assert.Nil(t, err)
	assert.Equal(t, expectedOutput, string(outputData))
}

func TestEndToEndWithValidationErrors(t *testing.T) {
	filename := "tests/example-unvalidatable.yaml"

	config, err := ReadConfigFromFile(filename)
	assert.Nil(t, err)

	validationResult, err := validation.ValidateObject(config)
	assert.Nil(t, err)

	err = validationResult.AsError()
	assert.ErrorContains(t, err, "send_limits account name 'test1' that does not exist")
	assert.ErrorContains(t, err, "account names must not be empty or whitespace")

	//write back the invalidated config anyway
	//use the temp to get a filename, but close that tempFile because we really just want the filename
	tempFile := util.CreateTempFileAndDir("test_output", "output.*.yaml")
	err = tempFile.Close()
	assert.Nil(t, err)

	//take advantage of the existing file to test the overwrite
	err = config.WriteConfigToFile(tempFile.Name(), true)
	assert.Nil(t, err)

	//read back the file and check its contents
	expectedFileOutput := `mail:
  accounts:
    - name: ""
      smtp:
        sender_address: example@example.com
        server_address: smtp.example.com
        port: 465
        username: joeuser@example.com
        password: sealed(+aaaaaa==)
      imap:
        server_address: imap.example.com
        port: 993
        username: janeuser@example.com
        password: sealed(+bbbbbb==)
  send_limits:
    - min_period_minutes: 10
      account_names:
        - test1
`
	outputData, err := os.ReadFile(tempFile.Name())
	assert.Nil(t, err)
	assert.Equal(t, expectedFileOutput, string(outputData))

}

func TestReadInvalidFilename(t *testing.T) {
	filename := "tests/nonexistent.yaml"

	c, err := ReadConfigFromFile(filename)

	assert.Nil(t, c)
	assert.ErrorContains(t, err, "file read error for config file")

}

func TestReadInvalidYaml(t *testing.T) {
	filename := "tests/invalid.yaml"

	c, err := ReadConfigFromFile(filename)

	assert.Nil(t, c)
	assert.ErrorContains(t, err, "unmarshal error")

}

type FailsYamlMarshal struct {
	Value string
}

func (o FailsYamlMarshal) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("intentional marshal failure")
}

func (o *FailsYamlMarshal) UnmarshalYAML(value *yaml.Node) error {
	return fmt.Errorf("intentional unmarshal failure")
}

func TestConfigWriteFailures(t *testing.T) {

	//this will fail marshalling
	object := FailsYamlMarshal{"foobar"}

	//try to convert it to yaml
	yamldata, err := objectToYaml(object)
	assert.Equal(t, yamldata, "")
	assert.ErrorContains(t, err, "intentional marshal failure")

	//try to write it to a file
	err = writeObjectToFile(object, "test.output", false)
	assert.ErrorContains(t, err, "intentional marshal failure")

	//try to write a valid config to an invalid filename
	input_filename := "tests/example.yaml"

	c, err := ReadConfigFromFile(input_filename)
	require.Nil(t, err)

	err = c.WriteConfigToFile("//////not_valid_filename", false)
	assert.ErrorContains(t, err, "could not open file")
}
