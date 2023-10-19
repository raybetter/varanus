package mail

import (
	"fmt"
	"time"
)

type WaitError struct {
	waitTime time.Duration
}

func (we WaitError) Error() string {
	return fmt.Sprintf("wait for %s", we.waitTime)
}

func (we WaitError) GetWaitTime() time.Duration {
	return we.waitTime
}
