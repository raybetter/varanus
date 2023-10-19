package mail

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitError(t *testing.T) {
	err := WaitError{10 * time.Second}
	assert.Equal(t, "wait for 10s", err.Error())
	assert.Equal(t, time.Second*10, err.GetWaitTime())
}
