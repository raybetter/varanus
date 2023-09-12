package secrets

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"hash"
	"os"
)

//--------------------------------------------------------------------------------------------------
// Helper functions shared by SecretSealer and SecretUnsealer
//--------------------------------------------------------------------------------------------------

const MIN_SECRET_EFFECTIVE_LENGTH int = 50

// GetSealerHash returns an instance of the hash used when sealing and unsealing the secrets
func getSealerHash() hash.Hash {
	return sha256.New()
}

func computeMaximumSecretSize(publicKey *rsa.PublicKey) int {
	if publicKey == nil {
		return 0
	}

	// per https://go.dev/src/crypto/rsa/rsa.go#L514
	// "The message must be no longer than the length of the public modulus minus twice the hash
	//    length, minus a further 2."
	return publicKey.Size() - (2 * getSealerHash().Size()) - 2
}

//--------------------------------------------------------------------------------------------------
// Unsealer
//--------------------------------------------------------------------------------------------------

// SecretUnsealer opens secrets sealed by SecretSealer for use by the system
type SecretUnsealer struct {
	privateKey *rsa.PrivateKey
}

func (su *SecretUnsealer) HasKey() bool {
	return su.privateKey != nil
}

func (su *SecretUnsealer) LoadPrivateKeyFromFile(filename string) error {
	keyBuffer, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading '%s': %w", filename, err)
	}
	err = su.LoadPrivateKey(keyBuffer)
	if err != nil {
		return fmt.Errorf("error reading key from file '%s': %w", filename, err)
	}
	return nil
}

func (su *SecretUnsealer) LoadPrivateKey(rawBytes []byte) error {
	privateBlock, _ := pem.Decode([]byte(rawBytes))
	genericPrivateKey, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	if err != nil {
		return fmt.Errorf("error decoding key: %w", err)
	}
	rsaPrivateKey, ok := genericPrivateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("the key is not an RSA private key")
	}

	//make sure the key is long enough to be useful
	if computeMaximumSecretSize(&rsaPrivateKey.PublicKey) < MIN_SECRET_EFFECTIVE_LENGTH {
		return fmt.Errorf("the key is too small to effectively encrypt secrets")
	}

	su.privateKey = rsaPrivateKey
	return nil
}

func (su *SecretUnsealer) ClearKeys() {
	su.privateKey = nil
}

func (su SecretUnsealer) UnsealSecret(cipherText string) (string, error) {
	if su.privateKey == nil {
		return "", fmt.Errorf("no private key set")
	}

	ct, _ := base64.StdEncoding.DecodeString(cipherText)
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	plaintext, err := rsa.DecryptOAEP(getSealerHash(), rng, su.privateKey, ct, label)

	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

//--------------------------------------------------------------------------------------------------
// Sealer
//--------------------------------------------------------------------------------------------------

// SecretSealer seal secrets in a way that they can be unsealed later by the SecretUnsealer
type SecretSealer struct {
	publicKey *rsa.PublicKey
}

func (ss *SecretSealer) HasKey() bool {
	return ss.publicKey != nil
}

func (ss *SecretSealer) LoadPublicKeyFromFile(filename string) error {
	keyBuffer, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading '%s': %w", filename, err)
	}
	err = ss.LoadPublicKey(keyBuffer)
	if err != nil {
		return fmt.Errorf("error reading key from file '%s': %w", filename, err)
	}
	return nil
}

func (ss *SecretSealer) LoadPublicKey(rawBytes []byte) error {
	publicBlock, _ := pem.Decode([]byte(rawBytes))
	genericPublicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		return fmt.Errorf("error decoding key: %w", err)
	}
	rsaPublicKey, ok := genericPublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("the key is not an RSA public key")
	}

	//make sure the key is long enough to be useful
	if computeMaximumSecretSize(rsaPublicKey) < MIN_SECRET_EFFECTIVE_LENGTH {
		return fmt.Errorf("the key is too small to effectively encrypt secrets")
	}

	ss.publicKey = rsaPublicKey
	return nil
}

func (ss *SecretSealer) ClearKeys() {
	ss.publicKey = nil
}

func (ss SecretSealer) GetMaximumSecretSize() (int, error) {
	if ss.publicKey == nil {
		return 0, fmt.Errorf("no public key set")
	}
	return computeMaximumSecretSize(ss.publicKey), nil
}

func (ss SecretSealer) SealSecret(secretToSeal string) (string, error) {
	if ss.publicKey == nil {
		return "", fmt.Errorf("no public key set")
	}

	// maxSecretSize := computeMaximumSecretSize(ss.publicKey)
	// if len(secretToSeal) > maxSecretSize {
	// 	return "", fmt.Errorf(
	// 		"the secret length (%d) exceeds the maximum secret length for this key (%d)",
	// 		len(secretToSeal), maxSecretSize,
	// 	)
	// }

	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(getSealerHash(), rng, ss.publicKey, []byte(secretToSeal), label)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
