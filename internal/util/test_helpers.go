package util

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// DeepCopyForTesting is an exceedingly lazy but compact way of deep copying the config structs
// with all their pointer objects also duplicated
//
// https://stackoverflow.com/questions/50269322/how-to-copy-struct-and-dereference-all-pointers
func DeepCopy(v interface{}) interface{} {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	vptr := reflect.New(reflect.TypeOf(v))
	err = json.Unmarshal(data, vptr.Interface())
	if err != nil {
		panic(err)
	}
	return vptr.Elem().Interface()
}

func CreateTempFileAndDir(dir string, pattern string) *os.File {
	//make the directory
	err := os.MkdirAll(dir, 0744)
	if err != nil {
		panic(fmt.Errorf("failed to make temp directory '%s': %w", dir, err))
	}

	//make the temp file
	tempFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		panic(fmt.Errorf("failed to make temp file in dir '%s' with pattern '%s': %w", dir, pattern, err))
	}
	return tempFile
}

// Ptr returns a pointer the input argument, allowing you to do Ptr("constant value") when you
// ~need~ **want** to set a *string in a struct definition in one line.
//
// https://stackoverflow.com/questions/30716354/how-do-i-do-a-literal-int64-in-go/30716481
func Ptr[T any](v T) *T {
	return &v
}

// Call within a test function so that the test is skipped unless VARANUS_TEST_FULL is set
func SlowTest(t *testing.T) {

	pc, file, line, ok := runtime.Caller(1)
	require.True(t, ok)

	if os.Getenv("VARANUS_TEST_FULL") == "" {
		t.Skipf("Skipping %v at %s:%d because VARANUS_TEST_FULL is not set", runtime.FuncForPC(pc).Name(), file, line)
	}
}
