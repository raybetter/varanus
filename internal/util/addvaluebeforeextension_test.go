package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddValueBeforeExtension(t *testing.T) {

	token := "NEWVALUE"

	type TestCase struct {
		filename         string
		expectedFilename string
	}
	testCases := []TestCase{
		{"example.txt", "example.NEWVALUE.txt"},
		{"foo.bar.example.txt", "foo.bar.example.NEWVALUE.txt"},
		{"exampletxt", "exampletxt.NEWVALUE"},
		{"", ".NEWVALUE"},
	}

	for index, testCase := range testCases {
		assert.Equal(t,
			testCase.expectedFilename, AddValueBeforeExtension(testCase.filename, token),
			"in test %d", index,
		)
	}

}
