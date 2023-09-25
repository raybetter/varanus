package config

import (
	"fmt"
	"slices"
	"strings"
)

// PathError allows a hierarchical config to annotate an error with the path as it is passed up the
// config.
type PathError struct {
	OriginalError error
	//PathTokens is a reverse order list of the path elements
	PathTokens []string
}

func (e PathError) Error() string {
	if len(e.PathTokens) > 0 {
		//reverse the slice
		revPath := make([]string, len(e.PathTokens))
		copy(revPath, e.PathTokens)
		slices.Reverse(revPath)
		return fmt.Errorf("at path %s an error occured: %w", strings.Join(revPath, "."), e.OriginalError).Error()
	} else {
		return e.OriginalError.Error()
	}
}

func (e *PathError) AddPathToken(token string) {
	e.PathTokens = append(e.PathTokens, token)
}

// AddTokenToPathError tries to convert err to a PathError and call AddPathToken with the supplied
// token.  Returns the original error with the path token added if this operation succeeded,
// or a new PathError object if the operation failed.
func AddTokenToPathError(err error, token string) error {
	if err == nil {
		return nil
	}
	pathError, ok := err.(PathError)
	if ok {
		pathError.AddPathToken(token)
		return pathError
	} else {
		newPathError := PathError{
			OriginalError: err,
			PathTokens:    []string{token},
		}
		return newPathError
	}

}

func AddTokenToPathErrorList(errList []error, token string) []error {
	newErrList := make([]error, len(errList))
	for index, err := range errList {
		newErrList[index] = AddTokenToPathError(err, token)
	}
	return newErrList
}
