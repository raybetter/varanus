package secrets

type SecretsManager interface {
	SetPrivateKey(key []byte) error
	ClearPrivateKey() error

	SealSecret(unsealed string) (string, error)
	UnsealSecret(sealed string) (string, error)
}
