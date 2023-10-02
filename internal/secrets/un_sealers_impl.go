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
	"regexp"
	"strings"

	"github.com/youmark/pkcs8"
)

//--------------------------------------------------------------------------------------------------
// Helper functions shared by SecretSealer and SecretUnsealer
//--------------------------------------------------------------------------------------------------

// SealedValueRegex matches the expected encoding produced by the sealing function
var SealedValueRegex = regexp.MustCompile(`^[A-Za-z0-9+/]+=+$`)

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

// secretUnsealerImpl opens secrets sealed by SecretSealer for use by the system
type secretUnsealerImpl struct {
	privateKey *rsa.PrivateKey
}

func (su *secretUnsealerImpl) HasKey() bool {
	return su.privateKey != nil
}

func (su *secretUnsealerImpl) LoadPrivateKeyFromFile(filename string, passphrase string) error {
	keyBuffer, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading '%s': %w", filename, err)
	}
	err = su.LoadPrivateKey(keyBuffer, passphrase)
	if err != nil {
		return fmt.Errorf("error reading key from file '%s': %w", filename, err)
	}
	return nil
}

func (su *secretUnsealerImpl) LoadPrivateKey(rawBytes []byte, passphrase string) error {
	privateBlock, _ := pem.Decode([]byte(rawBytes))
	if privateBlock == nil {
		return fmt.Errorf("pem decoding failed")
	}
	rsaPrivateKey, err := pkcs8.ParsePKCS8PrivateKeyRSA(privateBlock.Bytes, []byte(passphrase))
	if err != nil {
		if strings.Contains(err.Error(), "use ParseECPrivateKey instead") {
			return fmt.Errorf("not a supported key type.  Use an RSA key")
		}
		if strings.Contains(err.Error(), "incorrect password") {
			return fmt.Errorf("incorrect password or unsupported key type")
		}

		return fmt.Errorf("error decoding key: %w", err)
	}

	//make sure the key is long enough to be useful
	if computeMaximumSecretSize(&rsaPrivateKey.PublicKey) < MIN_SECRET_EFFECTIVE_LENGTH {
		return fmt.Errorf("the key is too small to effectively encrypt secrets")
	}

	su.privateKey = rsaPrivateKey
	return nil
}

func (su *secretUnsealerImpl) ClearKeys() {
	su.privateKey = nil
}

func (su secretUnsealerImpl) UnsealSecret(cipherText string) (string, error) {
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

func (su *secretUnsealerImpl) UnsealObject(objectToUnseal interface{}) UnsealResult {
	return UnsealObject(objectToUnseal, su)
}

func (su *secretUnsealerImpl) CheckSeals(objectToCheck interface{}) SealCheckResult {
	return CheckSealsOnObject(objectToCheck, su)
}

//--------------------------------------------------------------------------------------------------
// Sealer
//--------------------------------------------------------------------------------------------------

// secretSealerImpl seal secrets in a way that they can be unsealed later by the SecretUnsealer
type secretSealerImpl struct {
	publicKey *rsa.PublicKey
}

func (ss *secretSealerImpl) HasKey() bool {
	return ss.publicKey != nil
}

func (ss *secretSealerImpl) LoadPublicKeyFromFile(filename string) error {
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

func (ss *secretSealerImpl) LoadPublicKey(rawBytes []byte) error {
	publicBlock, _ := pem.Decode([]byte(rawBytes))
	if publicBlock == nil {
		return fmt.Errorf("pem decoding failed")
	}
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

func (ss *secretSealerImpl) ClearKeys() {
	ss.publicKey = nil
}

func (ss secretSealerImpl) GetMaximumSecretSize() (int, error) {
	if ss.publicKey == nil {
		return 0, fmt.Errorf("no public key set")
	}
	return computeMaximumSecretSize(ss.publicKey), nil
}

func (ss secretSealerImpl) SealSecret(secretToSeal string) (string, error) {
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

func (su *secretSealerImpl) SealObject(objectToSeal interface{}) SealResult {
	return SealObject(objectToSeal, su)
}
