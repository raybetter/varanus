package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUrlHost(t *testing.T) {
	testCases := map[string]bool{
		"google.com":            true,
		"smtp.varanus.org":      true,
		"smtp.red-and-blue.org": true,
		"poop.wtf":              true,
		"foo.com/argyle/socks":  false,
		"http://foo.com":        false,
		":::foo.com":            false,
		"foo,bar.com":           false,
		"foo__bar.com":          false,
		".leading.dot":          false,
	}

	for candidate, expectedResult := range testCases {
		actualResult := IsUrlHost(candidate)
		assert.Equalf(t, expectedResult, actualResult,
			"expected candidate value '%s' to be %t but it is %t",
			candidate, expectedResult, actualResult)
	}

}
