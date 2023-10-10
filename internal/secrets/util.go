package secrets

import (
	"fmt"
	"reflect"
	"strings"
	"varanus/internal/walker"
)

type SealCheckResult struct {
	UnsealedCount int
	SealedCount   int
	UnsealErrors  []error
}

func (r SealCheckResult) HumanReadable() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Of %d total items, %d are sealed and %d are unsealed.\n",
		r.UnsealedCount+r.SealedCount, r.SealedCount, r.UnsealedCount)
	if len(r.UnsealErrors) > 0 {
		fmt.Fprintf(&sb, "%d seal check errors were detected:\n", len(r.UnsealErrors))
		for _, se := range r.UnsealErrors {
			fmt.Fprintf(&sb, "  %s\n", se)
		}
	}
	return sb.String()
}

func CheckSealsOnObject(objectToSeal interface{}, unsealer SecretUnsealer) SealCheckResult {
	result := SealCheckResult{}

	sealedItemWorker := func(needle interface{}, path string) error {
		si := needle.(SealableReader)

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

	var sealedItemInterfaceType = reflect.TypeOf((*SealableReader)(nil)).Elem()
	err := walker.WalkObjectImmutable(objectToSeal, sealedItemInterfaceType, sealedItemWorker)
	if err != nil {
		// unreachable because the worker never returns an error
		panic(err)
	}
	return result
}

type SealResult struct {
	TotalUnsealedCount int
	TotalSealedCount   int
	NumberSealed       int
	SealErrors         []error
}

func (r SealResult) HumanReadable() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "The seal operation sealed %d items.\n", r.NumberSealed)
	fmt.Fprintf(&sb, "After the seal operation, of %d total items, %d are sealed and %d are unsealed.\n",
		r.TotalUnsealedCount+r.TotalSealedCount, r.TotalSealedCount, r.TotalUnsealedCount)
	if len(r.SealErrors) > 0 {
		fmt.Fprintf(&sb, "%d seal errors were detected:\n", len(r.SealErrors))
		for _, se := range r.SealErrors {
			fmt.Fprintln(&sb, "  ", se)
		}
	}

	return sb.String()
}

func SealObject(objectToSeal interface{}, sealer SecretSealer) SealResult {
	result := SealResult{}

	sealedItemWorker := func(needle interface{}, path string) error {
		si := needle.(SealableWriter)

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

	var sealedItemInterfaceType = reflect.TypeOf((*SealableWriter)(nil)).Elem()
	err := walker.WalkObjectMutable(objectToSeal, sealedItemInterfaceType, sealedItemWorker)
	if err != nil {
		// unreachable because the worker never returns an error
		panic(err)

	}
	return result
}

type UnsealResult struct {
	TotalUnsealedCount int
	TotalSealedCount   int
	NumberUnsealed     int
	UnsealErrors       []error
}

func (r UnsealResult) HumanReadable() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "The unseal operation unsealed %d items.\n", r.NumberUnsealed)
	fmt.Fprintf(&sb, "After the unseal operation, of %d total items, %d are sealed and %d are unsealed.\n",
		r.TotalUnsealedCount+r.TotalSealedCount, r.TotalSealedCount, r.TotalUnsealedCount)
	if len(r.UnsealErrors) > 0 {
		fmt.Fprintf(&sb, "%d unseal errors were detected:\n", len(r.UnsealErrors))
		for _, se := range r.UnsealErrors {
			fmt.Fprintln(&sb, "  ", se)
		}
	}

	return sb.String()
}

func UnsealObject(objectToSeal interface{}, unsealer SecretUnsealer) UnsealResult {
	result := UnsealResult{}

	sealedItemWorker := func(needle interface{}, path string) error {
		si := needle.(SealableWriter)

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

	var sealedItemInterfaceType = reflect.TypeOf((*SealableWriter)(nil)).Elem()
	err := walker.WalkObjectMutable(objectToSeal, sealedItemInterfaceType, sealedItemWorker)
	if err != nil {
		// unreachable because the worker never returns an error
		panic(err)
	}
	return result
}
