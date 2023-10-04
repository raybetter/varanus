package app

import (
	"os"
	"strings"
	"testing"
	"varanus/internal/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnsealConfigNominalNoForce(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "unseal_config_test.*.yaml")
	tempFile.Close()
	//delete the temp file so that we can use ForceOverwrite = false
	err := os.Remove(tempFile.Name())
	require.Nil(t, err)
	args := UnsealConfigArgs{
		Input:          util.Ptr("tests/example.yaml"),
		PrivateKey:     util.Ptr("tests/key-4096.pem"),
		Passphrase:     util.Ptr(""),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(false),
	}

	app := CreateApp()

	err = app.UnsealConfig(&args, &sb)
	assert.Nil(t, err)

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "The config was loaded successfully.")
	assert.Contains(t, stdOutput, "No validation errors")
	assert.Contains(t, stdOutput, "The unseal operation unsealed 1 items.")
	assert.Contains(t, stdOutput, "After the unseal operation, of 2 total items, 0 are sealed and 2 are unsealed.")

}

func TestUnsealConfigNoForceButFileExists(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "unseal_config_test.*.yaml")
	tempFile.Close()
	//don't delete the temp file
	args := UnsealConfigArgs{
		Input:          util.Ptr("tests/example.yaml"),
		PrivateKey:     util.Ptr("tests/key-4096.pem"),
		Passphrase:     util.Ptr(""),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(false),
	}

	app := CreateApp()

	err := app.UnsealConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Error writing output config")
	assert.ErrorContains(t, err, "could not open file")
	assert.ErrorContains(t, err, "file exists")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "The unseal operation unsealed 1 items.")
	assert.Contains(t, stdOutput, "After the unseal operation, of 2 total items, 0 are sealed and 2 are unsealed.")

}

func TestUnsealConfigUnvalidatable(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "unseal_config_test.*.yaml")
	tempFile.Close()
	args := UnsealConfigArgs{
		Input:          util.Ptr("tests/example-unvalidatable.yaml"),
		PrivateKey:     util.Ptr("tests/key-4096.pem"),
		Passphrase:     util.Ptr(""),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(true),
	}

	app := CreateApp()

	err := app.UnsealConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Refusing to unseal unvalidated config file")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "The config was loaded successfully.")
	assert.Contains(t, stdOutput, "2 Validation Errors")
	assert.Contains(t, stdOutput, "send_limits account name 'test1' that does not exist")
	assert.Contains(t, stdOutput, "account names must not be empty or whitespace")

}

func TestUnsealConfigValidationFailure(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "unseal_config_test.*.yaml")
	tempFile.Close()
	args := UnsealConfigArgs{
		Input:          util.Ptr("tests/example-validation-failure.yaml"),
		PrivateKey:     util.Ptr("tests/key-4096.pem"),
		Passphrase:     util.Ptr(""),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(true),
	}

	app := CreateApp()

	err := app.UnsealConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Refusing to unseal the configuration because validation had an error -- please report this as a bug")
	assert.ErrorContains(t, err, "intentional error from force_failure: validation_fails")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "Config validation failed: callback error at path=force_failure")
	assert.Contains(t, stdOutput, "intentional error from force_failure: validation_fails")

}

func TestUnsealConfigSealError(t *testing.T) {

	//this test uses a corrupted sealed() value to induce an unseal error for tests

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "unseal_config_test.*.yaml")
	tempFile.Close()
	args := UnsealConfigArgs{
		Input:          util.Ptr("tests/example-bad-seal.yaml"),
		PrivateKey:     util.Ptr("tests/key-4096.pem"),
		Passphrase:     util.Ptr(""),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(true),
	}

	app := CreateApp()

	err := app.UnsealConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "There were errors when unsealing the config.  Check the output for details.")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "The config was loaded successfully.")
	assert.Contains(t, stdOutput, "No validation errors")
	assert.Contains(t, stdOutput, "The unseal operation unsealed 0 items.")
	assert.Contains(t, stdOutput, "After the unseal operation, of 2 total items, 1 are sealed and 1 are unsealed.")
	assert.Contains(t, stdOutput, "1 unseal errors were detected")
	assert.Contains(t, stdOutput, "error at path mail.accounts[0].SMTP.password: failed to unseal secret crypto/rsa: decryption error")

}

func TestUnsealConfigInvalidYaml(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "unseal_config_test.*.yaml")
	tempFile.Close()
	args := UnsealConfigArgs{
		Input:          util.Ptr("tests/invalid.yaml"),
		PrivateKey:     util.Ptr("tests/key-4096.pem"),
		Passphrase:     util.Ptr(""),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(true),
	}

	app := CreateApp()

	err := app.UnsealConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Could not load config")
	assert.ErrorContains(t, err, "unmarshal error")

}

func TestUnsealConfigBadKeyFile(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "unseal_config_test.*.yaml")
	tempFile.Close()
	args := UnsealConfigArgs{
		Input:          util.Ptr("tests/example.yaml"),
		PrivateKey:     util.Ptr("tests/key-4096-bad.pem"),
		Passphrase:     util.Ptr(""),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(true),
	}

	app := CreateApp()

	err := app.UnsealConfig(&args, &sb)
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
