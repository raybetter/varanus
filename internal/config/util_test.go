package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathError(t *testing.T) {

	err1 := fmt.Errorf("Original Error")

	//wrap the original error
	err2 := AddTokenToPathError(err1, "token1")
	assert.Equal(t, "at path token1 an error occured: Original Error", err2.Error())

	//add a path to err2 with wrapper func
	err3 := AddTokenToPathError(err2, "token3")
	assert.Equal(t, "at path token3.token1 an error occured: Original Error", err3.Error())

	//cast and manually add a token
	err2PathError, ok := err2.(PathError)
	assert.True(t, ok)

	err2PathError.AddPathToken("token2")
	assert.Equal(t, "at path token2.token1 an error occured: Original Error", err2PathError.Error())

	//add a path to err2 with wrapper func
	err4 := AddTokenToPathError(err2PathError, "token4")
	assert.Equal(t, "at path token4.token2.token1 an error occured: Original Error", err4.Error())

	//path error with no tokens
	err5 := PathError{
		OriginalError: fmt.Errorf("A test error"),
	}
	assert.Equal(t, "A test error", err5.Error())
}

func TestPathErrorList(t *testing.T) {
	errorList := []error{
		AddTokenToPathError(fmt.Errorf("Error A"), "A"),
		AddTokenToPathError(fmt.Errorf("Error B"), "B"),
		AddTokenToPathError(fmt.Errorf("Error C"), "C"),
	}
	newErrorList := AddTokenToPathErrorList(errorList, "Z")
	for index, newError := range newErrorList {

		newPathError, ok := newError.(PathError)
		assert.True(t, ok)
		oldPathError, ok := errorList[index].(PathError)
		assert.True(t, ok)

		assert.Equal(t, oldPathError.OriginalError.Error(), newPathError.OriginalError.Error())
		assert.Len(t, newPathError.PathTokens, 2)
		assert.Equal(t, oldPathError.PathTokens[0], newPathError.PathTokens[0])
		assert.Equal(t, "Z", newPathError.PathTokens[1])
	}
}
