package secrets

import (
	"fmt"
	"reflect"
	"varanus/internal/walker"
)

type SealCheckResult struct {
	UnsealedCount int
	SealedCount   int
	UnsealErrors  []error
}

var sealedItemInterfaceType = reflect.TypeOf(SealedItem{})

func (r SealCheckResult) HumanReadable() {
	fmt.Printf("Of %d total items, %d are sealed and %d are unsealed.\n",
		r.UnsealedCount+r.SealedCount, r.SealedCount, r.UnsealedCount)
	if len(r.UnsealErrors) > 0 {
		fmt.Printf("%d seal errors were detected:\n", len(r.UnsealErrors))
		for _, se := range r.UnsealErrors {
			fmt.Println("  ", se)
		}
	}
}

func CheckSealsOnObject(objectToSeal interface{}, unsealer SecretUnsealer) (SealCheckResult, error) {
	result := SealCheckResult{}

	sealedItemWorker := func(needle interface{}, path string) error {
		si := needle.(SealedItem)

		if !si.IsValueSealed() {
			result.UnsealedCount += 1
		} else {
			result.SealedCount += 1
			err := si.Check(unsealer)
			if err != nil {
				err = fmt.Errorf("error at path %s: %w", path, err)
				result.UnsealErrors = append(result.UnsealErrors, err)
			}
		}
		return nil
	}

	err := walker.WalkObjectImmutable(objectToSeal, sealedItemInterfaceType, sealedItemWorker)
	if err != nil {
		// unreachable because there are no inducable errors in the worker callback
		return SealCheckResult{}, err
	}
	return result, nil
}

type SealResult struct {
	TotalUnsealedCount int
	TotalSealedCount   int
	NumberSealed       int
	SealErrors         []error
}

func (r SealResult) Dump() {
	fmt.Printf("The seal operation sealed %d items.\n", r.NumberSealed)
	fmt.Printf("After the seal operation, of %d total items, %d are sealed and %d are unsealed.\n",
		r.TotalUnsealedCount+r.TotalSealedCount, r.TotalSealedCount, r.TotalUnsealedCount)
	if len(r.SealErrors) > 0 {
		fmt.Printf("%d seal errors were detected:\n", len(r.SealErrors))
		for _, se := range r.SealErrors {
			fmt.Println("  ", se)
		}
	}
}

func SealObject(objectToSeal interface{}, sealer SecretSealer) (SealResult, error) {
	result := SealResult{}

	sealedItemWorker := func(needle interface{}, path string) error {
		si := needle.(*SealedItem)

		//already sealed
		if si.IsValueSealed() {
			result.TotalSealedCount += 1
			return nil
		}

		//unsealed, so try to seal the item
		err := si.Seal(sealer)
		if err == nil {
			//seal successful
			result.TotalSealedCount += 1
			result.NumberSealed += 1
		} else {
			//seal failed
			result.TotalUnsealedCount += 1
			err = fmt.Errorf("error at path %s: %w", path, err)
			result.SealErrors = append(result.SealErrors, err)
		}
		return nil
	}

	err := walker.WalkObjectMutable(objectToSeal, sealedItemInterfaceType, sealedItemWorker)
	if err != nil {
		// unreachable because there are no inducable errors in the worker callback
		return SealResult{}, err
	}
	return result, nil
}

type UnsealResult struct {
	TotalUnsealedCount int
	TotalSealedCount   int
	NumberUnsealed     int
	UnsealErrors       []error
}

func (r UnsealResult) Dump() {
	fmt.Printf("The unseal operation unsealed %d items.\n", r.NumberUnsealed)
	fmt.Printf("After the unseal operation, of %d total items, %d are sealed and %d are unsealed.\n",
		r.TotalUnsealedCount+r.TotalSealedCount, r.TotalSealedCount, r.TotalUnsealedCount)
	if len(r.UnsealErrors) > 0 {
		fmt.Printf("%d seal errors were detected:\n", len(r.UnsealErrors))
		for _, se := range r.UnsealErrors {
			fmt.Println("  ", se)
		}
	}
}

func UnsealObject(objectToSeal interface{}, unsealer SecretUnsealer) (UnsealResult, error) {
	result := UnsealResult{}

	sealedItemWorker := func(needle interface{}, path string) error {
		si := needle.(*SealedItem)

		//already unsealed
		if !si.IsValueSealed() {
			result.TotalUnsealedCount += 1
			return nil
		}

		//sealed, so try to unseal the item
		err := si.Unseal(unsealer)
		if err == nil {
			//unseal successful
			result.TotalUnsealedCount += 1
			result.NumberUnsealed += 1
		} else {
			//unseal failed
			result.TotalSealedCount += 1
			err = fmt.Errorf("error at path %s: %w", path, err)
			result.UnsealErrors = append(result.UnsealErrors, err)
		}
		return nil
	}

	err := walker.WalkObjectMutable(objectToSeal, sealedItemInterfaceType, sealedItemWorker)
	if err != nil {
		// unreachable because there are no inducable errors in the worker callback
		return UnsealResult{}, err
	}
	return result, nil
}
