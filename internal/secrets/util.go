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
