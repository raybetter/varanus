package secrets

type SealCheckResult struct {
	UnsealedCount int
	SealedCount   int
	UnsealErrors  []error
}

func (scr *SealCheckResult) Append(other SealCheckResult) {
	scr.UnsealedCount += other.UnsealedCount
	scr.SealedCount += other.SealedCount
	scr.UnsealErrors = append(scr.UnsealErrors, other.UnsealErrors...)
}

type SealResult struct {
	TotalUnsealedCount int
	TotalSealedCount   int
	NumberSealed       int
	SealErrors         []error
}

func SealObject(target interface{}, sealer SecretSealer) (SealResult, error) {
	result := SealResult{}

	sealedItemWorker := func(si *SealedItem) error {

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
			result.SealErrors = append(result.SealErrors, err)
		}
		return nil
	}

	err := iterateSealedItems(target, sealedItemWorker)
	if err != nil {
		return SealResult{}, err
	}
	return result, nil
}

// iterateSealedItems uses reflection to work through the fields and subfields of target,
// identifying any members that are SealedItems and passing each item to the worker for some
// operation.
func iterateSealedItems(target interface{}, worker func(*SealedItem) error) error {

	// targetType := reflect.TypeOf(target)
	// switch targetType.Kind {

	// }
	return nil
}
