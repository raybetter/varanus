package secrets

import (
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var randomRegex = regexp.MustCompile(`^[A-Za-z0-9 ~!@#$%^&*()-=_+[\]{};':",.<>/?\x60|\\]+$`)

const (
	TEST_FILE_PREFIX                    = "tests/"
	PUBLIC_KEY_4096_WITH_PASS_FILENAME  = "key-4096-with-pw.pub"
	PRIVATE_KEY_4096_WITH_PASS_FILENAME = "key-4096-with-pw.pem"
	PUBLIC_KEY_4096_FILENAME            = "key-4096.pub"
	PRIVATE_KEY_4096_FILENAME           = "key-4096.pem"
	PUBLIC_KEY_2048_FILENAME            = "key-2048.pub"
	PRIVATE_KEY_2048_FILENAME           = "key-2048.pem"
	PUBLIC_KEY_512_FILENAME             = "key-512.pub"
	PRIVATE_KEY_512_FILENAME            = "key-512.pem"
	PUBLIC_KEY_EC_FILENAME              = "key-EC.pub"
	PRIVATE_KEY_EC_FILENAME             = "key-EC.pem"
	PUBLIC_KEY_UNSUPPORTED_FILENAME     = "key-EC-unsupported.pub"
	PRIVATE_KEY_UNSUPPORTED_FILENAME    = "key-EC-unsupported.pem"
	PUBLIC_KEY_EC_WITH_PW_FILENAME      = "key-EC-with-pw.pub"
	PRIVATE_KEY_EC_WITH_PW_FILENAME     = "key-EC-with-pw.pem"
	NOT_A_KEY_FILENAME                  = "not-a-key.txt"
	EMPTY_KEY_FILE_FILENAME             = "empty_key_file.txt"
	TEST_KEY_PASSPHRASE                 = "testpassword!"
)

func makeRandomString(length int) string {
	message := make([]byte, length)
	for index := 0; index < length; index++ {
		message[index] = byte(rand.Intn(95) + 32) //sets an ascii value
	}
	return string(message)
}

func TestMakeRandomString(t *testing.T) {
	for i := 0; i < 50; i++ {
		randomString := makeRandomString(50 + i)
		assert.Len(t, randomString, 50+i)
		assert.Truef(t, randomRegex.Match([]byte(randomString)), "for string '%s'", randomString)
	}

}

func TestSealAndUnsealWithPassphrase(t *testing.T) {

	sealer := secretSealerImpl{}
	err := sealer.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_4096_WITH_PASS_FILENAME)
	assert.Nil(t, err)

	unsealer := secretUnsealerImpl{}
	err = unsealer.LoadPrivateKeyFromFile(
		TEST_FILE_PREFIX+PRIVATE_KEY_4096_WITH_PASS_FILENAME,
		TEST_KEY_PASSPHRASE)
	assert.Nil(t, err)

	secretMessage := `I am a very model of a modern major general`

	sealedSecret, err := sealer.SealSecret(secretMessage)
	assert.Nil(t, err)

	assert.True(t, SealedValueRegex.Match([]byte(sealedSecret)))

	recoveredSecret, err := unsealer.UnsealSecret(sealedSecret)
	assert.Nil(t, err)

	assert.Equal(t, secretMessage, recoveredSecret)

}

func TestBadKeyPassphrase(t *testing.T) {

	unsealer := secretUnsealerImpl{}
	err := unsealer.LoadPrivateKeyFromFile(
		TEST_FILE_PREFIX+PRIVATE_KEY_4096_WITH_PASS_FILENAME,
		"NOT THE RIGHT PASSPHRASE")
	assert.ErrorContains(t, err, "incorrect password or unsupported key type")

}

func TestSealAndUnsealWithoutPassphrase(t *testing.T) {

	sealer := secretSealerImpl{}
	sealer.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_4096_FILENAME)

	unsealer := secretUnsealerImpl{}
	unsealer.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+PRIVATE_KEY_4096_FILENAME, "")

	secretMessage := `I am a very model of a modern major general`

	sealedSecret, err := sealer.SealSecret(secretMessage)
	assert.Nil(t, err)

	assert.True(t, SealedValueRegex.Match([]byte(sealedSecret)))

	recoveredSecret, err := unsealer.UnsealSecret(sealedSecret)
	assert.Nil(t, err)

	assert.Equal(t, secretMessage, recoveredSecret)

}

func TestUninitializedErrors(t *testing.T) {

	sealerLong := secretSealerImpl{}
	unsealerLong := secretUnsealerImpl{}

	maxLen, err := sealerLong.GetMaximumSecretSize()
	assert.Equal(t, 0, maxLen)
	assert.ErrorContains(t, err, "no public key set")

	failedSealStr, err := sealerLong.SealSecret("doesn't matter")
	assert.Equal(t, "", failedSealStr)
	assert.ErrorContains(t, err, "no public key set")

	failedUnsealStr, err := unsealerLong.UnsealSecret("doesn't matter")
	assert.Equal(t, "", failedUnsealStr)
	assert.ErrorContains(t, err, "no private key set")

}

func TestBadFileErrors(t *testing.T) {

	//invalid key files
	{
		sealer := secretSealerImpl{}
		err := sealer.LoadPublicKeyFromFile(TEST_FILE_PREFIX + NOT_A_KEY_FILENAME)
		assert.ErrorContains(t, err, "error decoding key")
	}
	{
		unsealer := secretUnsealerImpl{}
		err := unsealer.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+NOT_A_KEY_FILENAME, "")
		assert.ErrorContains(t, err, "error decoding key")
	}

	//empty key files
	{
		sealer := secretSealerImpl{}
		err := sealer.LoadPublicKeyFromFile(TEST_FILE_PREFIX + EMPTY_KEY_FILE_FILENAME)
		assert.ErrorContains(t, err, "pem decoding failed")
	}
	{
		unsealer := secretUnsealerImpl{}
		err := unsealer.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+EMPTY_KEY_FILE_FILENAME, "")
		assert.ErrorContains(t, err, "pem decoding failed")
	}

	//nonexistent filenames
	{
		sealer := secretSealerImpl{}
		err := sealer.LoadPublicKeyFromFile("nonexistent/path/and/file.pem")
		//for unknown reason, sometimes the error is different, so allow either error
		assert.True(t,
			(strings.Contains(err.Error(), "The system cannot find the path specified") ||
				strings.Contains(err.Error(), "no such file or directory")))
	}

	{
		unsealer := secretUnsealerImpl{}
		err := unsealer.LoadPrivateKeyFromFile("nonexistent/path/and/file.pub", "")
		//for unknown reason, sometimes the error is different, so allow either error
		assert.True(t,
			(strings.Contains(err.Error(), "The system cannot find the path specified") ||
				strings.Contains(err.Error(), "no such file or directory")))
	}

	//unsupported key type
	{
		sealer := secretSealerImpl{}
		err := sealer.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_EC_FILENAME)
		assert.ErrorContains(t, err, "the key is not an RSA public key")
	}

	{
		unsealer := secretUnsealerImpl{}
		err := unsealer.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+PRIVATE_KEY_EC_FILENAME, "")
		assert.ErrorContains(t, err, "not a supported key type.  Use an RSA key")
	}

	//keys with unsupported ciphers
	{
		sealer := secretSealerImpl{}
		err := sealer.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_UNSUPPORTED_FILENAME)
		assert.ErrorContains(t, err, "unsupported elliptic curve")
	}

	{
		unsealer := secretUnsealerImpl{}
		err := unsealer.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+PRIVATE_KEY_UNSUPPORTED_FILENAME, "")
		assert.ErrorContains(t, err, "not a supported key type.  Use an RSA key")
	}

	{
		sealer := secretSealerImpl{}
		err := sealer.LoadPublicKeyFromFile(
			TEST_FILE_PREFIX + PUBLIC_KEY_EC_WITH_PW_FILENAME,
		)
		assert.ErrorContains(t, err, "the key is not an RSA public key")
	}

	{
		unsealer := secretUnsealerImpl{}
		err := unsealer.LoadPrivateKeyFromFile(
			TEST_FILE_PREFIX+PRIVATE_KEY_EC_WITH_PW_FILENAME,
			TEST_KEY_PASSPHRASE,
		)
		assert.ErrorContains(t, err, "key block is not of type RSA")
	}

}

func TestMaxLengths(t *testing.T) {

	const EXPECTED_4096_MESSAGE_SIZE int = 446
	const EXPECTED_2048_MESSAGE_SIZE int = 190

	sealer4096 := secretSealerImpl{}
	unsealer4096 := secretUnsealerImpl{}
	sealer2048 := secretSealerImpl{}
	unsealer2048 := secretUnsealerImpl{}
	sealer512 := secretSealerImpl{}
	unsealer512 := secretUnsealerImpl{}

	//load 4096 keys
	{
		err := sealer4096.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_4096_FILENAME)
		assert.Nil(t, err)
		assert.True(t, sealer4096.HasKey())

		err = unsealer4096.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+PRIVATE_KEY_4096_FILENAME, "")
		assert.Nil(t, err)
		assert.True(t, unsealer4096.HasKey())
	}

	//load 2048 keys
	{
		err := sealer2048.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_2048_FILENAME)
		assert.Nil(t, err)
		assert.True(t, sealer2048.HasKey())

		err = unsealer2048.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+PRIVATE_KEY_2048_FILENAME, "")
		assert.Nil(t, err)
		assert.True(t, unsealer2048.HasKey())
	}

	//this sealer should fail to load because the key is too short
	{
		err := sealer512.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_512_FILENAME)
		assert.ErrorContains(t, err, "the key is too small to effectively encrypt secrets")
		assert.False(t, sealer512.HasKey())

		err = unsealer512.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+PRIVATE_KEY_512_FILENAME, "")
		assert.ErrorContains(t, err, "the key is too small to effectively encrypt secrets")
		assert.False(t, sealer512.HasKey())
	}

	//check maximum sizes
	{
		maxSize4096, err := sealer4096.GetMaximumSecretSize()
		assert.Nil(t, err)
		assert.Equal(t, EXPECTED_4096_MESSAGE_SIZE, maxSize4096)
	}

	{
		maxSize2048, err := sealer2048.GetMaximumSecretSize()
		assert.Nil(t, err)
		assert.Equal(t, EXPECTED_2048_MESSAGE_SIZE, maxSize2048)
	}

	//nil call to helper function
	{
		maxSize := computeMaximumSecretSize(nil)
		assert.Equal(t, maxSize, 0)
	}

	//seal and unseal things of different lengths to validate the lengths
	seal_and_unseal := func(bufferLen int, context string, maxLength int, sealer *secretSealerImpl, unsealer *secretUnsealerImpl) {

		//make a random message of a certain length
		messageStr := makeRandomString(bufferLen)

		sealedMessage, err := sealer.SealSecret(messageStr)

		if bufferLen > maxLength {
			//expect failure if too long
			assert.Equalf(t, "", sealedMessage, "for message length %d with %s", bufferLen, context)
			assert.ErrorContains(t, err, "message too long for RSA key size", "for message length %d with %s", bufferLen, context)
		} else {
			assert.Nilf(t, err, "for message length %d with %s", bufferLen, context)

			//not too long, so expect success
			assert.Truef(t, SealedValueRegex.Match([]byte(sealedMessage)), "for message length %d with %s", bufferLen, context)

			recoveredMessage, err := unsealer.UnsealSecret(sealedMessage)
			assert.Nilf(t, err, "for message length %d with %s", bufferLen, context)
			assert.Equalf(t, messageStr, recoveredMessage, "for message length %d with %s", bufferLen, context)
		}

	}

	for buf_length := 0; buf_length < 450; buf_length += 50 {
		seal_and_unseal(buf_length, "2048-key", EXPECTED_2048_MESSAGE_SIZE, &sealer2048, &unsealer2048)
		seal_and_unseal(buf_length, "4096-key", EXPECTED_4096_MESSAGE_SIZE, &sealer4096, &unsealer4096)
	}

	//clear the keys
	{
		assert.True(t, sealer2048.HasKey())
		sealer2048.ClearKeys()
		assert.False(t, sealer2048.HasKey())

		assert.True(t, unsealer2048.HasKey())
		unsealer2048.ClearKeys()
		assert.False(t, unsealer2048.HasKey())

		assert.True(t, sealer4096.HasKey())
		sealer4096.ClearKeys()
		assert.False(t, sealer4096.HasKey())

		assert.True(t, unsealer4096.HasKey())
		unsealer4096.ClearKeys()
		assert.False(t, unsealer4096.HasKey())
	}

}

func TestInvalidCiphertext(t *testing.T) {

	sealer4096 := secretSealerImpl{}
	unsealer4096 := secretUnsealerImpl{}
	sealer2048 := secretSealerImpl{}
	unsealer2048 := secretUnsealerImpl{}

	{
		err := sealer4096.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_4096_FILENAME)
		assert.Nil(t, err)
		assert.True(t, sealer4096.HasKey())
	}

	{
		err := unsealer4096.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+PRIVATE_KEY_4096_FILENAME, "")
		assert.Nil(t, err)
		assert.True(t, unsealer4096.HasKey())
	}

	{
		err := sealer2048.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_2048_FILENAME)
		assert.Nil(t, err)
		assert.True(t, sealer2048.HasKey())
	}

	{
		err := unsealer2048.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+PRIVATE_KEY_2048_FILENAME, "")
		assert.Nil(t, err)
		assert.True(t, unsealer2048.HasKey())
	}

	message := "This is a test message."

	sealedMessage4096, err := sealer4096.SealSecret(message)
	assert.Nil(t, err)

	//can unseal with our own unsealer
	{
		recoveredMessage4096, err := unsealer4096.UnsealSecret(sealedMessage4096)
		assert.Nil(t, err)
		assert.Equal(t, message, recoveredMessage4096)
	}

	//can't unseal with a different unsealer
	{
		recoveredMessage2048, err := unsealer2048.UnsealSecret(sealedMessage4096)
		assert.Equal(t, "", recoveredMessage2048)
		assert.ErrorContains(t, err, "crypto/rsa: decryption error")
	}

	//can't unseal a random string of bytes
	{
		invalidCipherText := makeRandomString(200)
		recoveredInvalidCipherText, err := unsealer2048.UnsealSecret(invalidCipherText)
		assert.Equal(t, "", recoveredInvalidCipherText)
		assert.ErrorContains(t, err, "crypto/rsa: decryption error")
	}

}

// MockSecretHolder provides a SecretHolder interface wrapper around a sealed item for testing
type MockSecretHolder struct {
	si SealedItem
}

func (msh *MockSecretHolder) Seal(sealer SecretSealer) error {
	return msh.si.Seal(sealer)
}
func (msh *MockSecretHolder) Unseal(unsealer SecretUnsealer) error {
	return msh.si.Unseal(unsealer)
}
func (msh *MockSecretHolder) CheckSeals(unsealer SecretUnsealer) SealCheckResult {
	return msh.si.CheckSeals(unsealer)
}

func TestSecretHolderMethods(t *testing.T) {

	sealer := secretSealerImpl{}
	unsealer := secretUnsealerImpl{}

	//make a sealer and unsealer
	err := sealer.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_4096_FILENAME)
	assert.Nil(t, err)

	err = unsealer.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+PRIVATE_KEY_4096_FILENAME, "")
	assert.Nil(t, err)

	//make a mock secret holder
	secretValue := "it's a secret."
	msh := MockSecretHolder{
		si: CreateSealedItem(secretValue),
	}

	//check the initial, unsealed state
	{
		check := msh.CheckSeals(&unsealer)
		assert.Equal(t, 0, check.SealedCount)
		assert.Equal(t, 1, check.UnsealedCount)
		assert.Len(t, check.UnsealErrors, 0)
	}

	//seal and check
	sealer.SealSecretHolder(&msh)
	{
		check := msh.CheckSeals(&unsealer)
		assert.Equal(t, 1, check.SealedCount)
		assert.Equal(t, 0, check.UnsealedCount)
		assert.Len(t, check.UnsealErrors, 0)
	}

	//unseal and check
	unsealer.UnsealSecretHolder(&msh)
	{
		check := msh.CheckSeals(&unsealer)
		assert.Equal(t, 0, check.SealedCount)
		assert.Equal(t, 1, check.UnsealedCount)
		assert.Len(t, check.UnsealErrors, 0)
	}

	//corrupt and check
	msh.si.isSealed = true
	msh.si.value = "+aaaaaa=="
	{
		check := msh.CheckSeals(&unsealer)
		assert.Equal(t, 1, check.SealedCount)
		assert.Equal(t, 0, check.UnsealedCount)
		assert.Len(t, check.UnsealErrors, 1)
		assert.ErrorContains(t, check.UnsealErrors[0], "crypto/rsa: decryption error")
	}

}
