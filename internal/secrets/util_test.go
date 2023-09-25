package secrets

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSealCheckResult(t *testing.T) {
	scr1 := SealCheckResult{10, 20, []error{fmt.Errorf("an error")}}

	scr2 := SealCheckResult{1, 2, []error{fmt.Errorf("another error")}}

	assert.Equal(t, 10, scr1.UnsealedCount)
	assert.Equal(t, 20, scr1.SealedCount)
	assert.Len(t, scr1.UnsealErrors, 1)
	assert.ErrorContains(t, scr1.UnsealErrors[0], "an error")

	assert.Equal(t, 1, scr2.UnsealedCount)
	assert.Equal(t, 2, scr2.SealedCount)
	assert.Len(t, scr2.UnsealErrors, 1)
	assert.ErrorContains(t, scr2.UnsealErrors[0], "another error")

	scr1.Append(scr2)

	assert.Equal(t, 11, scr1.UnsealedCount)
	assert.Equal(t, 22, scr1.SealedCount)
	assert.Len(t, scr1.UnsealErrors, 2)
	assert.ErrorContains(t, scr1.UnsealErrors[0], "an error")
	assert.ErrorContains(t, scr1.UnsealErrors[1], "another error")

	assert.Equal(t, 1, scr2.UnsealedCount)
	assert.Equal(t, 2, scr2.SealedCount)
	assert.Len(t, scr2.UnsealErrors, 1)
	assert.ErrorContains(t, scr2.UnsealErrors[0], "another error")

}
