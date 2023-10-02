package app

import (
	"strings"
	"testing"
	"varanus/internal/util"

	"github.com/stretchr/testify/assert"
)

func TestCheckConfigNominal(t *testing.T) {

	var sb strings.Builder

	args := CheckConfigArgs{
		Input:      util.Ptr("tests/example.yaml"),
		PrivateKey: util.Ptr("tests/key-4096.pem"),
		Passphrase: util.Ptr(""),
	}

	app := CreateApp()

	err := app.CheckConfig(&args, &sb)
	assert.Nil(t, err)

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	assert.Contains(t, stdOutput, "Of 2 total items, 1 are sealed and 1 are unsealed.")
	assert.Contains(t, stdOutput, "The config was loaded successfully.")
	assert.Contains(t, stdOutput, "The integrity of the sealed values was verified with the private key.")
	assert.Contains(t, stdOutput, "The configuration appears to be valid.")

}

func TestCheckConfigUnvalidatableNoKey(t *testing.T) {

	var sb strings.Builder

	args := CheckConfigArgs{
		Input:      util.Ptr("tests/example-unvalidatable.yaml"),
		PrivateKey: util.Ptr(""),
		Passphrase: util.Ptr(""),
	}

	app := CreateApp()

	err := app.CheckConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "There are some issues with the configuration")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "Of 2 total items, 2 are sealed and 0 are unsealed.")
	assert.Contains(t, stdOutput, "The config was loaded successfully.")
	assert.Contains(t, stdOutput, "2 Validation Errors")
	assert.Contains(t, stdOutput, "send_limits account name 'test1' that does not exist")
	assert.Contains(t, stdOutput, "account names must not be empty or whitespace")
	assert.Contains(t, stdOutput, "The integrity of sealed values was not checked because no private key was provided.")

}

func TestCheckConfigValidationFailure(t *testing.T) {

	var sb strings.Builder

	args := CheckConfigArgs{
		Input:      util.Ptr("tests/example-validation-failure.yaml"),
		PrivateKey: util.Ptr(""),
		Passphrase: util.Ptr(""),
	}

	app := CreateApp()

	err := app.CheckConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Checking did not complete because the configuration validation had an error -- please report this as a bug")
	assert.ErrorContains(t, err, "intentional error from force_failure: validation_fails")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "Config validation failed: callback error at path=force_failure")
	assert.Contains(t, stdOutput, "intentional error from force_failure: validation_fails")

}

func TestCheckConfigSealError(t *testing.T) {

	var sb strings.Builder

	args := CheckConfigArgs{
		Input:      util.Ptr("tests/example-bad-seal.yaml"),
		PrivateKey: util.Ptr("tests/key-4096.pem"),
		Passphrase: util.Ptr(""),
	}

	app := CreateApp()

	err := app.CheckConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "There are some issues with the configuration.  See the output above for details.")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "The config was loaded successfully.")
	assert.Contains(t, stdOutput, "No validation errors")
	assert.Contains(t, stdOutput, "Of 2 total items, 1 are sealed and 1 are unsealed.")
	assert.Contains(t, stdOutput, "The integrity of the sealed values was checked with the private key, but there were some errors.")
	assert.Contains(t, stdOutput, "1 seal errors were detected")
	assert.Contains(t, stdOutput, "error at path mail.accounts[0].SMTP.password: crypto/rsa: decryption error")

}

func TestCheckConfigInvalidYaml(t *testing.T) {

	var sb strings.Builder

	args := CheckConfigArgs{
		Input:      util.Ptr("tests/invalid.yaml"),
		PrivateKey: util.Ptr(""),
		Passphrase: util.Ptr(""),
	}

	app := CreateApp()

	err := app.CheckConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Could not load config")
	assert.ErrorContains(t, err, "unmarshal error")

}

func TestCheckConfigBadKeyFile(t *testing.T) {

	var sb strings.Builder

	args := CheckConfigArgs{
		Input:      util.Ptr("tests/example.yaml"),
		PrivateKey: util.Ptr("tests/key-4096-bad.pem"),
		Passphrase: util.Ptr(""),
	}

	app := CreateApp()

	err := app.CheckConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Could not load private key")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "The config was loaded successfully.")
	assert.Contains(t, stdOutput, "No validation errors")

}
