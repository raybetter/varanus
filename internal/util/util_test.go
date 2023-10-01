package util

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MyObject struct {
	Inner1 *InnerObject
	Inner2 *InnerObject
}

type InnerObject struct {
	Value string
}

func TestDeepCopy(t *testing.T) {
	original := MyObject{&InnerObject{"value1"}, &InnerObject{"value2"}}

	copy := DeepCopy(original).(MyObject)

	original.Inner1.Value = "new value1"
	original.Inner2.Value = "new value2"

	assert.Equal(t, "value1", copy.Inner1.Value)
	assert.Equal(t, "value2", copy.Inner2.Value)

	copy.Inner1.Value = "another value1"
	copy.Inner2.Value = "another value2"

	assert.Equal(t, "new value1", original.Inner1.Value)
	assert.Equal(t, "new value2", original.Inner2.Value)

}

type FailsMarshal struct {
	Value string
}

func (v FailsMarshal) MarshalJSON() ([]byte, error) {
	return []byte{}, fmt.Errorf("intentional failure")
}

func (v *FailsMarshal) UnmarshalJSON(valueByte []byte) error {
	v.Value = string(valueByte)
	return nil
}

func TestDeepCopyMarshalPanic(t *testing.T) {
	original := FailsMarshal{"value"}
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		} else {
			rStr := fmt.Sprintf("%s", r)
			assert.Contains(t, rStr, "error calling MarshalJSON for type util.FailsMarshal: intentional failure")
		}
	}()
	DeepCopy(original)

}

type FailsUnmarshal struct {
	Value string
}

func (v FailsUnmarshal) MarshalJSON() ([]byte, error) {
	return []byte(`"` + v.Value + `"`), nil
}

func (v *FailsUnmarshal) UnmarshalJSON(valueByte []byte) error {
	return fmt.Errorf("intentional failure")
}

func TestDeepCopyUnmarshalPanic(t *testing.T) {
	original := FailsUnmarshal{"value"}
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		} else {
			rStr := fmt.Sprintf("%s", r)
			assert.Contains(t, rStr, "intentional failure")
		}
	}()
	DeepCopy(original)

}

func TestCreateTempFileAndDirBadPath(t *testing.T) {

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		} else {
			rStr := fmt.Sprintf("%s", r)
			assert.Contains(t, rStr, "failed to make temp directory")
		}
	}()
	CreateTempFileAndDir("//////_invalid_dir_path", "unit_test_output.*.txt")
}

func TestCreateTempFileAndDirBadPattern(t *testing.T) {

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		} else {
			rStr := fmt.Sprintf("%s", r)
			assert.Contains(t, rStr, "failed to make temp file")
		}
	}()
	CreateTempFileAndDir("test_output/valid_dir_path", "//////invalid_pattern.*.txt")

}

func TestCreateTempFileAndDirNominal(t *testing.T) {
	outputDir := "test_output"
	testFileContents := "test output"

	file := CreateTempFileAndDir(outputDir, "unit_test_output.*.txt")

	//nominal path - write to the file and close it
	_, err := file.WriteString(testFileContents)
	assert.Nil(t, err)
	err = file.Close()
	assert.Nil(t, err)

	//make sure the folder exists:
	_, err = os.Stat(outputDir)
	assert.False(t, os.IsNotExist(err))

	//read back the file
	fileData, err := os.ReadFile(file.Name())
	assert.Nil(t, err)
	assert.Equal(t, testFileContents, string(fileData))

}
