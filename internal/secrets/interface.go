package secrets

type SecretHolder interface {
	Seal(sealer SecretSealer) error
	Unseal(unsealer SecretUnsealer) error
	CheckSeals(unsealer SecretUnsealer) SealCheckResult
}

type SecretSealer interface {
	LoadPublicKeyFromFile(filename string) error
	LoadPublicKey(rawBytes []byte) error
	ClearKeys()
	GetMaximumSecretSize() (int, error)
	SealSecret(secretToSeal string) (string, error)
	SealSecretHolder(holder SecretHolder)
}

type SecretUnsealer interface {
	LoadPrivateKeyFromFile(filename string, passphrase string) error
	LoadPrivateKey(rawBytes []byte, passphrase string) error
	ClearKeys()
	UnsealSecret(cipherText string) (string, error)
	UnsealSecretHolder(holder SecretHolder)
}

func MakeSecretSealer() SecretSealer {
	return &secretSealerImpl{}
}
func MakeSecretUnsealer() SecretUnsealer {
	return &secretUnsealerImpl{}
}
