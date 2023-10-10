package cmd

import (
	"testing"
	"varanus/internal/app"

	"github.com/stretchr/testify/assert"
)

func TestConfigCmd(t *testing.T) {

	testCases := []testCase{
		//nominal case with just the config should not call anything
		{
			arguments: []string{"config"},
			outputsContain: []string{
				"Usage:\n  varanus config [command]",
				"Available Commands:",
				"Flags:",
			},
			expectedCallCount: 0,
		},
	}
	runTestCases(t, testCases)
}

func TestConfigCheckCmd(t *testing.T) {

	testCases := []testCase{
		//call check with no args --> missing input arg error
		{
			arguments: []string{"config", "check"},
			outputsContain: []string{
				"Usage:\n  varanus config check [flags]",
				"Flags:",
			},
			errorContains: []string{
				"required flag(s) \"input\" not set",
			},
			expectedCallCount: 0,
		},
		//call check with input arg only
		{
			arguments: []string{"config", "check", "-i", "foo.yaml"},
			outputsContain: []string{
				"CheckConfig called with args",
			},
			expectedCallCount: 1,
			checkCalls: func(t *testing.T, calls []mockAppCalls) {
				assert.Equal(t, "CheckConfig", calls[0].function)
				argObj := calls[0].argsObj.(*app.CheckConfigArgs)
				assert.Equal(t, "foo.yaml", *argObj.Input)
				assert.Equal(t, "", *argObj.Passphrase)
				assert.Equal(t, "", *argObj.PrivateKey)
			},
		},
		//call check with input and key args
		{
			arguments: []string{"config", "check", "-i", "foo.yaml", "-k", "keyfile.pem", "--passphrase", "argyle"},
			outputsContain: []string{
				"CheckConfig called with args",
			},
			expectedCallCount: 1,
			checkCalls: func(t *testing.T, calls []mockAppCalls) {
				assert.Equal(t, "CheckConfig", calls[0].function)
				argObj := calls[0].argsObj.(*app.CheckConfigArgs)
				assert.Equal(t, "foo.yaml", *argObj.Input)
				assert.Equal(t, "keyfile.pem", *argObj.PrivateKey)
				assert.Equal(t, "argyle", *argObj.Passphrase)
			},
		},
		//call check that returns an error
		{
			arguments:  []string{"config", "check", "-i", "foo.yaml"},
			appMutator: func(mva *mockVaranusApp) { mva.sealCheckConfigError = "injected error" },
			outputsContain: []string{
				"CheckConfig called with args",
			},
			errorContains:     []string{"injected error"},
			expectedCallCount: 1,
			checkCalls: func(t *testing.T, calls []mockAppCalls) {
				assert.Equal(t, "CheckConfig", calls[0].function)
				argObj := calls[0].argsObj.(*app.CheckConfigArgs)
				assert.Equal(t, "foo.yaml", *argObj.Input)
				assert.Equal(t, "", *argObj.PrivateKey)
				assert.Equal(t, "", *argObj.Passphrase)
			},
		},
	}
	runTestCases(t, testCases)
}

func TestConfigSealCmd(t *testing.T) {

	testCases := []testCase{
		//call seal with no args --> missing flags
		{
			arguments: []string{"config", "seal"},
			outputsContain: []string{
				"Usage:\n  varanus config seal [flags]",
				"Flags:",
			},
			errorContains: []string{
				"required flag(s) \"input\", \"publicKey\" not set",
			},
			expectedCallCount: 0,
		},
		//call seal with input only --> missing flags
		{
			arguments: []string{"config", "seal", "-i", "input.yaml"},
			outputsContain: []string{
				"Usage:\n  varanus config seal [flags]",
				"Flags:",
			},
			errorContains: []string{
				"required flag(s) \"publicKey\" not set",
			},
			expectedCallCount: 0,
		},
		//call seal with key only --> missing flags
		{
			arguments: []string{"config", "seal", "-k", "keyfile"},
			outputsContain: []string{
				"Usage:\n  varanus config seal [flags]",
				"Flags:",
			},
			errorContains: []string{
				"required flag(s) \"input\" not set",
			},
			expectedCallCount: 0,
		},
		//call seal with input and key args
		{
			arguments: []string{"config", "seal", "-i", "input.yaml", "-k", "keyfile.pub"},
			outputsContain: []string{
				"SealConfig called with args",
			},
			expectedCallCount: 1,
			checkCalls: func(t *testing.T, calls []mockAppCalls) {
				assert.Equal(t, "SealConfig", calls[0].function)
				argObj := calls[0].argsObj.(*app.SealConfigArgs)
				assert.Equal(t, "input.yaml", *argObj.Input)
				assert.Equal(t, "keyfile.pub", *argObj.PublicKey)
				assert.Equal(t, "input.sealed.yaml", *argObj.Output)
				assert.Equal(t, false, *argObj.ForceOverwrite)
			},
		},
		//call seal with input, key, output, force args
		{
			arguments: []string{"config", "seal", "-i", "input.yaml", "-k", "keyfile.pub", "-o", "output.yaml", "-f"},
			outputsContain: []string{
				"SealConfig called with args",
			},
			expectedCallCount: 1,
			checkCalls: func(t *testing.T, calls []mockAppCalls) {
				assert.Equal(t, "SealConfig", calls[0].function)
				argObj := calls[0].argsObj.(*app.SealConfigArgs)
				assert.Equal(t, "input.yaml", *argObj.Input)
				assert.Equal(t, "keyfile.pub", *argObj.PublicKey)
				assert.Equal(t, "output.yaml", *argObj.Output)
				assert.Equal(t, true, *argObj.ForceOverwrite)
			},
		},
		//call seal that returns an error
		{
			arguments:  []string{"config", "seal", "-i", "input.yaml", "-k", "keyfile.pub"},
			appMutator: func(mva *mockVaranusApp) { mva.sealConfigError = "injected error" },
			outputsContain: []string{
				"SealConfig called with args",
			},
			errorContains:     []string{"injected error"},
			expectedCallCount: 1,
			checkCalls: func(t *testing.T, calls []mockAppCalls) {
				assert.Equal(t, "SealConfig", calls[0].function)
				argObj := calls[0].argsObj.(*app.SealConfigArgs)
				assert.Equal(t, "input.yaml", *argObj.Input)
				assert.Equal(t, "keyfile.pub", *argObj.PublicKey)
				assert.Equal(t, "input.sealed.yaml", *argObj.Output)
				assert.Equal(t, false, *argObj.ForceOverwrite)
			},
		},
	}
	runTestCases(t, testCases)
}

func TestConfigUnsealCmd(t *testing.T) {

	testCases := []testCase{
		//call seal with no args --> missing flags
		{
			arguments: []string{"config", "unseal"},
			outputsContain: []string{
				"Usage:\n  varanus config unseal [flags]",
				"Flags:",
			},
			errorContains: []string{
				"required flag(s) \"input\", \"privateKey\" not set",
			},
			expectedCallCount: 0,
		},
		//call seal with input only --> missing flags
		{
			arguments: []string{"config", "unseal", "-i", "input.yaml"},
			outputsContain: []string{
				"Usage:\n  varanus config unseal [flags]",
				"Flags:",
			},
			errorContains: []string{
				"required flag(s) \"privateKey\" not set",
			},
			expectedCallCount: 0,
		},
		//call seal with key only --> missing flags
		{
			arguments: []string{"config", "unseal", "-k", "keyfile"},
			outputsContain: []string{
				"Usage:\n  varanus config unseal [flags]",
				"Flags:",
			},
			errorContains: []string{
				"required flag(s) \"input\" not set",
			},
			expectedCallCount: 0,
		},
		//call seal with input and key args
		{
			arguments: []string{"config", "unseal", "-i", "input.yaml", "-k", "keyfile.pem"},
			outputsContain: []string{
				"UnsealConfig called with args",
			},
			expectedCallCount: 1,
			checkCalls: func(t *testing.T, calls []mockAppCalls) {
				assert.Equal(t, "UnsealConfig", calls[0].function)
				argObj := calls[0].argsObj.(*app.UnsealConfigArgs)
				assert.Equal(t, "input.yaml", *argObj.Input)
				assert.Equal(t, "keyfile.pem", *argObj.PrivateKey)
				assert.Equal(t, "", *argObj.Passphrase)
				assert.Equal(t, "input.unsealed.yaml", *argObj.Output)
				assert.Equal(t, false, *argObj.ForceOverwrite)
			},
		},
		//call seal with input, key, output, force, passphrase args
		{
			arguments: []string{"config", "unseal", "-i", "input.yaml", "-k", "keyfile.pem", "-o", "output.yaml", "-p", "mypassphrase", "-f"},
			outputsContain: []string{
				"UnsealConfig called with args",
			},
			expectedCallCount: 1,
			checkCalls: func(t *testing.T, calls []mockAppCalls) {
				assert.Equal(t, "UnsealConfig", calls[0].function)
				argObj := calls[0].argsObj.(*app.UnsealConfigArgs)
				assert.Equal(t, "input.yaml", *argObj.Input)
				assert.Equal(t, "keyfile.pem", *argObj.PrivateKey)
				assert.Equal(t, "mypassphrase", *argObj.Passphrase)
				assert.Equal(t, "output.yaml", *argObj.Output)
				assert.Equal(t, true, *argObj.ForceOverwrite)
			},
		},
		//call seal that returns an error
		{
			arguments:  []string{"config", "unseal", "-i", "input.yaml", "-k", "keyfile.pem"},
			appMutator: func(mva *mockVaranusApp) { mva.unsealConfigError = "injected error" },
			outputsContain: []string{
				"UnsealConfig called with args",
			},
			errorContains:     []string{"injected error"},
			expectedCallCount: 1,
			checkCalls: func(t *testing.T, calls []mockAppCalls) {
				assert.Equal(t, "UnsealConfig", calls[0].function)
				argObj := calls[0].argsObj.(*app.UnsealConfigArgs)
				assert.Equal(t, "input.yaml", *argObj.Input)
				assert.Equal(t, "keyfile.pem", *argObj.PrivateKey)
				assert.Equal(t, "", *argObj.Passphrase)
				assert.Equal(t, "input.unsealed.yaml", *argObj.Output)
				assert.Equal(t, false, *argObj.ForceOverwrite)
			},
		},
	}
	runTestCases(t, testCases)
}
