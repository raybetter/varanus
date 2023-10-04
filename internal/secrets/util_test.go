package secrets

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSealCheckResult(t *testing.T) {
	result := SealCheckResult{10, 20, []error{fmt.Errorf("an error")}}
	output := result.HumanReadable()
	expectedOutput := `Of 30 total items, 20 are sealed and 10 are unsealed.
1 seal check errors were detected:
  an error
`
	assert.Equal(t, expectedOutput, output)
}

func TestSealResult(t *testing.T) {
	result := SealResult{10, 20, 5, []error{fmt.Errorf("an error"), fmt.Errorf("another error")}}
	output := result.HumanReadable()
	expectedOutput := `The seal operation sealed 5 items.
After the seal operation, of 30 total items, 20 are sealed and 10 are unsealed.
2 seal errors were detected:
   an error
   another error
`
	assert.Equal(t, expectedOutput, output)
}

func TestUnsealResult(t *testing.T) {
	result := UnsealResult{1, 2, 15, []error{fmt.Errorf("an error"), fmt.Errorf("another error")}}
	output := result.HumanReadable()
	expectedOutput := `The unseal operation unsealed 15 items.
After the unseal operation, of 3 total items, 2 are sealed and 1 are unsealed.
2 unseal errors were detected:
   an error
   another error
`
	assert.Equal(t, expectedOutput, output)
}

func TestEnsureSealedItemIsSealable(t *testing.T) {
	readerFun := func(s SealableReader) {
	}
	writerFun := func(s SealableWriter) {
	}

	si := CreateSealedItem("foo")

	//this test will not compile if SealedItem does not implement SealableReader
	readerFun(si)
	//this test will not compile if *SealedItem does not implement SealableWriter
	writerFun(&si)
}
