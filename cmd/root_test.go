package cmd

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"varanus/internal/app"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockVaranusApp struct {
	calls                []mockAppCalls
	sealConfigError      string
	unsealConfigError    string
	sealCheckConfigError string
}

type mockAppCalls struct {
	function string
	argsObj  interface{}
}

func (mva *mockVaranusApp) SealConfig(args *app.SealConfigArgs, outputStream io.Writer) error {
	fmt.Fprintf(outputStream, "SealConfig called with args %#v", args)
	mva.calls = append(mva.calls, mockAppCalls{
		function: "SealConfig",
		argsObj:  args,
	})
	if mva.sealConfigError != "" {
		return fmt.Errorf(mva.sealConfigError)
	}
	return nil
}
func (mva *mockVaranusApp) UnsealConfig(args *app.UnsealConfigArgs, outputStream io.Writer) error {
	fmt.Fprintf(outputStream, "UnsealConfig called with args %#v", args)
	mva.calls = append(mva.calls, mockAppCalls{
		function: "UnsealConfig",
		argsObj:  args,
	})
	if mva.unsealConfigError != "" {
		return fmt.Errorf(mva.unsealConfigError)
	}
	return nil

}
func (mva *mockVaranusApp) CheckConfig(args *app.CheckConfigArgs, outputStream io.Writer) error {
	fmt.Fprintf(outputStream, "CheckConfig called with args %#v", args)
	mva.calls = append(mva.calls, mockAppCalls{
		function: "CheckConfig",
		argsObj:  args,
	})
	if mva.sealCheckConfigError != "" {
		return fmt.Errorf(mva.sealCheckConfigError)
	}
	return nil
}

type testCase struct {
	//arguments supplied to the command
	arguments []string
	//if not nil, called with the mockApp object before executing the command
	appMutator func(*mockVaranusApp)
	//a set of strings tested against the commandline output
	outputsContain []string
	// a set of strings tested against the error -- if none are supplied, and checkError is not set,
	// then assert the error should be nil
	errorContains []string
	// expected number of calls to the mock
	expectedCallCount int
	// if not nil, called with list of calls recorded by the mock after the command is executed
	checkCalls func(*testing.T, []mockAppCalls)
	// if not nil, called with the error returned by the command when it is executed
	checkError func(*testing.T, error)
}

func runTestCases(t *testing.T, testCases []testCase) {
	for index, testCase := range testCases {

		t.Logf("running testcase %d", index)

		//setup the mock and the command with the args from the testcase
		mockApp := mockVaranusApp{}

		if testCase.appMutator != nil {
			testCase.appMutator(&mockApp)
		}

		context := CmdContext{
			App: &mockApp,
		}

		var cmdOutput strings.Builder
		var cmdError strings.Builder

		command := MakeRootCmd(&context)
		command.SetArgs(testCase.arguments)
		command.SetOut(&cmdOutput)
		command.SetErr(&cmdError)

		//execute the command

		err := command.Execute()

		//check the err returned
		if testCase.checkError == nil && len(testCase.errorContains) == 0 {
			assert.Nil(t, err)
		} else {
			require.NotNil(t, err)
			if testCase.checkError != nil {
				testCase.checkError(t, err)
			}
			cmdErrorStr := cmdError.String()
			for _, errStr := range testCase.errorContains {
				assert.Contains(t, err.Error(), errStr)
				//these strings should also be in the stderr output
				assert.Contains(t, cmdErrorStr, errStr)
			}
		}

		//check the calls recorded by the mock
		assert.Len(t, mockApp.calls, testCase.expectedCallCount)

		if testCase.checkCalls != nil {
			testCase.checkCalls(t, mockApp.calls)
		}

		//check the commandline output from the command
		cmdOutputStr := cmdOutput.String()
		// fmt.Println(cmdOutputStr)
		for _, expectedOutput := range testCase.outputsContain {
			assert.Contains(t, cmdOutputStr, expectedOutput)
		}
	}
}

func TestEmptyCall(t *testing.T) {

	testCases := []testCase{
		{
			arguments: []string{},
			outputsContain: []string{
				"Usage:\n  varanus [command]",
				"Available Commands:",
				"Flags:",
			},
			expectedCallCount: 0,
		},
	}
	runTestCases(t, testCases)
}
