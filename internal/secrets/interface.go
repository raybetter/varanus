package secrets

type SecretSealer interface {
	LoadPublicKeyFromFile(filename string) error
	LoadPublicKey(rawBytes []byte) error
	ClearKeys()
	GetMaximumSecretSize() (int, error)
	SealSecret(secretToSeal string) (string, error)
	SealObject(objectToSeal interface{}) SealResult
}

type SecretUnsealer interface {
	LoadPrivateKeyFromFile(filename string, passphrase string) error
	LoadPrivateKey(rawBytes []byte, passphrase string) error
	ClearKeys()
	UnsealSecret(cipherText string) (string, error)
	UnsealObject(objectToUnseal interface{}) UnsealResult
	CheckSeals(objectToCheck interface{}) SealCheckResult
}

// CreateUnsafeSealedItem creates a sealed item from raw data with no checks -- primarily used for testing.
func CreateUnsafeSealedItem(value string, isSealed bool) SealedItem {
	return SealedItem{value, isSealed}
}

func CreateSealedItem(value string) SealedItem {
	processedValue, isSealed := processSealedItemString(value)
	return SealedItem{processedValue, isSealed}
}

func MakeSecretSealer() SecretSealer {
	return &secretSealerImpl{}
}
func MakeSecretUnsealer() SecretUnsealer {
	return &secretUnsealerImpl{}
}

type SealableReader interface {
	IsValueSealed() bool
	GetValue() string
	Check(unsealer SecretUnsealer) error
}

type SealableWriter interface {
	SealableReader
	Seal(sealer SecretSealer) error
	Unseal(unsealer SecretUnsealer) error
}
