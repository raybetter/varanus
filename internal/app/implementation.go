package app

import (
	"fmt"
)

type varanusAppImpl struct {
}

func CreateApp() VaranusApp {
	return varanusAppImpl{}
}

func newApplicationError(format string, args ...interface{}) error {
	return ApplicationError{
		theError: fmt.Errorf(format, args...),
	}
}
