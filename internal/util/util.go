package util

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
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
