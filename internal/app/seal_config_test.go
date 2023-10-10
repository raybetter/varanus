package app

import (
	"os"
	"strings"
	"testing"
	"varanus/internal/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSealConfigNominalNoForce(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "seal_config_test.*.yaml")
	tempFile.Close()
	//delete the temp file so that we can use ForceOverwrite = false
	err := os.Remove(tempFile.Name())
	require.Nil(t, err)
	args := SealConfigArgs{
		Input:          util.Ptr("tests/example.yaml"),
		PublicKey:      util.Ptr("tests/key-4096.pub"),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(false),
	}

	app := CreateApp()

	err = app.SealConfig(&args, &sb)
	assert.Nil(t, err)

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "The seal operation sealed 1 items.")
	assert.Contains(t, stdOutput, "After the seal operation, of 2 total items, 2 are sealed and 0 are unsealed.")

}

func TestSealConfigNoForceButFileExists(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "seal_config_test.*.yaml")
	tempFile.Close()
	//don't delete the temp file
	args := SealConfigArgs{
		Input:          util.Ptr("tests/example.yaml"),
		PublicKey:      util.Ptr("tests/key-4096.pub"),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(false),
	}

	app := CreateApp()

	err := app.SealConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Error writing output config")
	assert.ErrorContains(t, err, "could not open file")
	assert.ErrorContains(t, err, "file exists")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "The seal operation sealed 1 items.")
	assert.Contains(t, stdOutput, "After the seal operation, of 2 total items, 2 are sealed and 0 are unsealed.")

}

func TestSealConfigUnvalidatable(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "seal_config_test.*.yaml")
	tempFile.Close()
	args := SealConfigArgs{
		Input:          util.Ptr("tests/example-unvalidatable.yaml"),
		PublicKey:      util.Ptr("tests/key-4096.pub"),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(true),
	}

	app := CreateApp()

	err := app.SealConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Refusing to seal unvalidated config file")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "The config was loaded successfully.")
	assert.Contains(t, stdOutput, "2 Validation Errors")
	assert.Contains(t, stdOutput, "send_limits account name 'test1' that does not exist")
	assert.Contains(t, stdOutput, "account names must not be empty or whitespace")

}

func TestSealConfigValidationFailure(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "seal_config_test.*.yaml")
	tempFile.Close()
	args := SealConfigArgs{
		Input:          util.Ptr("tests/example-validation-failure.yaml"),
		PublicKey:      util.Ptr("tests/key-4096.pub"),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(true),
	}

	app := CreateApp()

	err := app.SealConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Refusing to seal the configuration because validation had an error -- please report this as a bug")
	assert.ErrorContains(t, err, "intentional error from force_failure: validation_fails")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "Config validation failed: callback error at path=force_failure")
	assert.Contains(t, stdOutput, "intentional error from force_failure: validation_fails")

}

func TestSealConfigSealError(t *testing.T) {

	//this test uses the force_failure in the yaml to induce a seal error for tests

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "seal_config_test.*.yaml")
	tempFile.Close()
	args := SealConfigArgs{
		Input:          util.Ptr("tests/example-seal-failure.yaml"),
		PublicKey:      util.Ptr("tests/key-4096.pub"),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(true),
	}

	app := CreateApp()

	err := app.SealConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "There were errors when sealing the config.  Check the output for details.")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "The config was loaded successfully.")
	assert.Contains(t, stdOutput, "No validation errors")
	assert.Contains(t, stdOutput, "The seal operation sealed 1 items.")
	assert.Contains(t, stdOutput, "After the seal operation, of 3 total items, 2 are sealed and 1 are unsealed.")
	assert.Contains(t, stdOutput, "1 seal errors were detected")
	assert.Contains(t, stdOutput, "error at path force_failure: intentional error from force_failure: seal_fails")

}

func TestSealConfigInvalidYaml(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "seal_config_test.*.yaml")
	tempFile.Close()
	args := SealConfigArgs{
		Input:          util.Ptr("tests/invalid.yaml"),
		PublicKey:      util.Ptr("tests/key-4096.pub"),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(true),
	}

	app := CreateApp()

	err := app.SealConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Could not load config")
	assert.ErrorContains(t, err, "unmarshal error")

}

func TestSealConfigBadKeyFile(t *testing.T) {

	var sb strings.Builder

	tempFile := util.CreateTempFileAndDir("test_output", "seal_config_test.*.yaml")
	tempFile.Close()
	args := SealConfigArgs{
		Input:          util.Ptr("tests/example.yaml"),
		PublicKey:      util.Ptr("tests/key-4096-bad.pem"),
		Output:         util.Ptr(tempFile.Name()),
		ForceOverwrite: util.Ptr(true),
	}

	app := CreateApp()

	err := app.SealConfig(&args, &sb)
	assert.NotNil(t, err)
	_, ok := err.(ApplicationError)
	assert.True(t, ok)
	assert.ErrorContains(t, err, "Could not load public key")

	stdOutput := sb.String()
	// fmt.Println(stdOutput)
	// t.FailNow()
	assert.Contains(t, stdOutput, "The config was loaded successfully.")
	assert.Contains(t, stdOutput, "No validation errors")

}
