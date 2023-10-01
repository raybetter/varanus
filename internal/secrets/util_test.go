package secrets

import (
	"fmt"
)

func ExampleSealCheckResult() {
	result := SealCheckResult{10, 20, []error{fmt.Errorf("an error")}}
	result.HumanReadable()
	// Output:
	// Of 30 total items, 20 are sealed and 10 are unsealed.
	// 1 seal errors were detected:
	//    an error
}

func ExampleSealResult() {
	result := SealResult{10, 20, 5, []error{fmt.Errorf("an error"), fmt.Errorf("another error")}}
	result.Dump()
	// Output:
	// The seal operation sealed 5 items.
	// After the seal operation, of 30 total items, 20 are sealed and 10 are unsealed.
	// 2 seal errors were detected:
	//    an error
	//    another error
}

func ExampleUnsealResult() {
	result := UnsealResult{1, 2, 15, []error{fmt.Errorf("an error"), fmt.Errorf("another error")}}
	result.Dump()
	// Output:
	// The unseal operation unsealed 15 items.
	// After the unseal operation, of 3 total items, 2 are sealed and 1 are unsealed.
	// 2 seal errors were detected:
	//    an error
	//    another error
}
