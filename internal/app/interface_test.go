package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeepCopy(t *testing.T) {
	err := newApplicationError("test error %s %d", "some string", 55)
	assert.Equal(t, "test error some string 55", err.Error())

	appError, ok := err.(ApplicationError)
	assert.True(t, ok)

	assert.Equal(t, "test error some string 55", appError.Unwrap().Error())
}
