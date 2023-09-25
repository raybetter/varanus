package secrets

import (
	"testing"
	"varanus/internal/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSealedItemSealed(t *testing.T) {

	//item is already sealed
	si := CreateSealedItem("sealed(+abcdef+jk==)")

	assert.Equal(t, true, si.IsValueSealed())
	assert.Equal(t, "sealed(+abcdef+jk==)", si.GetValue())

	//should have no validation failures
	vp := validation.ValidationProcess{}

	err := si.Validate(&vp)
	assert.Nil(t, err)

	err = vp.GetFinalValidationError()
	assert.Nil(t, err)

	// check string representations
	expectedString := `secrets.SealedItem{value:"+abcdef+jk==", isSealed:true}`
	assert.Equal(t, expectedString, si.String())
	assert.Equal(t, expectedString, si.GoString())

	//check marshal values
	yamlValue, err := si.MarshalYAML()
	assert.Nil(t, err)
	assert.Equal(t, "sealed(+abcdef+jk==)", yamlValue)

	jsonValue, err := si.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, []byte(`"sealed(+abcdef+jk==)"`), jsonValue)
}

func TestSealedItemUnsealed(t *testing.T) {
	//item is not sealed
	si := CreateSealedItem("some text")

	assert.Equal(t, false, si.IsValueSealed())
	assert.Equal(t, "some text", si.GetValue())

	//should have no validation failures
	vp := validation.ValidationProcess{}

	err := si.Validate(&vp)
	assert.Nil(t, err)

	err = vp.GetFinalValidationError()
	assert.Nil(t, err)

	// check string representations
	expectedString := `secrets.SealedItem{value:"<unsealed value redacted>", isSealed:false}`
	assert.Equal(t, expectedString, si.String())
	assert.Equal(t, expectedString, si.GoString())

	//check marshal values
	yamlValue, err := si.MarshalYAML()
	assert.Nil(t, err)
	assert.Equal(t, "some text", yamlValue)

	jsonValue, err := si.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, `"some text"`, string(jsonValue))
}

func TestUnsafeCreation(t *testing.T) {
	//unsafe creation lets you do anything you want with the values
	{
		//nominal unsealed case
		si := CreateUnsafeSealedItem("foo", false)
		assert.Equal(t, false, si.isSealed)
		assert.Equal(t, "foo", si.value)
	}
	{
		//nominal sealed case
		si := CreateUnsafeSealedItem("foo", true)
		assert.Equal(t, true, si.isSealed)
		assert.Equal(t, "foo", si.value)
	}
}

func TestSealedItemValidationErrors(t *testing.T) {
	type TestCase struct {
		CreateStr string
		VError    string
	}

	testCases := []TestCase{
		{
			CreateStr: "",
			VError:    "SealedItem with an unsealed value should not be empty",
		},
		{
			CreateStr: "sealed(foobar)",
			VError:    "value does not match the expected format for an encrypted, encoded string",
		},
	}

	for index, testCase := range testCases {

		si := CreateSealedItem(testCase.CreateStr)

		//should have a validation error because the value is empty
		vp := validation.ValidationProcess{}

		err := si.Validate(&vp)
		assert.Nilf(t, err, "for test index %d", index)

		err = vp.GetFinalValidationError()
		require.NotNilf(t, err, "for test index %d", index)

		assert.ErrorContainsf(t, err, testCase.VError, "for test index %d", index)
	}

}

func TestSealedItemUnmarshaling(t *testing.T) {
	//sealed yaml
	{
		si := SealedItem{}

		yamlNode := yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "sealed(+abcde==)",
		}

		err := si.UnmarshalYAML(&yamlNode)
		assert.Nil(t, err)

		assert.Equal(t, true, si.isSealed)
		assert.Equal(t, "+abcde==", si.value)
	}
	//unsealed yaml
	{
		si := SealedItem{}

		yamlNode := yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "foo",
		}

		err := si.UnmarshalYAML(&yamlNode)
		assert.Nil(t, err)

		assert.Equal(t, false, si.isSealed)
		assert.Equal(t, "foo", si.value)
	}
	//invalid yaml node
	{
		si := SealedItem{}

		yamlNode := yaml.Node{
			Kind:  yaml.SequenceNode, //wrong node type
			Value: "sealed(+abcde==)",
		}

		err := si.UnmarshalYAML(&yamlNode)
		assert.ErrorContains(t, err, "expected a scalar value")
	}
	//sealed json
	{
		si := SealedItem{}
		si.UnmarshalJSON([]byte(`"sealed(+abcde==)"`))
		assert.Equal(t, true, si.isSealed)
		assert.Equal(t, "+abcde==", si.value)
	}
	//unsealed json
	{
		si := SealedItem{}
		si.UnmarshalJSON([]byte(`"bar"`))
		assert.Equal(t, false, si.isSealed)
		assert.Equal(t, "bar", si.value)
	}

}

func TestSealedItemSealAndUnseal(t *testing.T) {

	secretValue := "unsealed value -- it's a secret!"

	//make an unsealed value
	si := CreateSealedItem(secretValue)

	assert.Equal(t, false, si.IsValueSealed())
	assert.Equal(t, secretValue, si.GetValue())

	{
		//check the item with no key
		sealCheckResult := si.CheckSeals(nil)
		assert.Equal(t, 1, sealCheckResult.UnsealedCount)
		assert.Equal(t, 0, sealCheckResult.SealedCount)
		assert.Len(t, sealCheckResult.UnsealErrors, 0)
	}

	//get a sealer to seal it
	sealer := MakeSecretSealer()
	err := sealer.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_4096_FILENAME)
	assert.Nil(t, err)

	err = si.Seal(sealer)
	assert.Nil(t, err)

	assert.Equal(t, true, si.IsValueSealed())

	{
		//validate the sealed value to make sure our validation works on an actual sealed value
		//should have no validation failures
		vp := validation.ValidationProcess{}
		err = si.Validate(&vp)
		assert.Nil(t, err)
		err = vp.GetFinalValidationError()
		assert.Nil(t, err)
	}

	{
		//check the sealed item with no key
		sealCheckResult := si.CheckSeals(nil)
		assert.Equal(t, 0, sealCheckResult.UnsealedCount)
		assert.Equal(t, 1, sealCheckResult.SealedCount)
		assert.Len(t, sealCheckResult.UnsealErrors, 0)
	}

	//a second sealing should succeed but not change state
	oldValue := si.value
	err = si.Seal(sealer)
	assert.Nil(t, err)

	assert.Equal(t, true, si.IsValueSealed())
	assert.Equal(t, oldValue, si.value)

	//make an unsealer with a private key
	unsealer := MakeSecretUnsealer()
	unsealer.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+PRIVATE_KEY_4096_FILENAME, "")

	{
		//check the sealed item with the key
		sealCheckResult := si.CheckSeals(unsealer)
		assert.Equal(t, 0, sealCheckResult.UnsealedCount)
		assert.Equal(t, 1, sealCheckResult.SealedCount)
		assert.Len(t, sealCheckResult.UnsealErrors, 0)
	}

	//now unseal it
	err = si.Unseal(unsealer)
	assert.Nil(t, err)

	assert.False(t, si.IsValueSealed())
	assert.Equal(t, secretValue, si.GetValue())

	{
		//check the item with the key
		sealCheckResult := si.CheckSeals(unsealer)
		assert.Equal(t, 1, sealCheckResult.UnsealedCount)
		assert.Equal(t, 0, sealCheckResult.SealedCount)
		assert.Len(t, sealCheckResult.UnsealErrors, 0)
	}

	//now unseal it again -- should have no error and no effect
	err = si.Unseal(unsealer)
	assert.Nil(t, err)

	assert.False(t, si.IsValueSealed())
	assert.Equal(t, secretValue, si.GetValue())

}

func TestSealedItemSealAndUnsealeWithErrors(t *testing.T) {

	secretValue := "unsealed value -- it's a secret!"

	//make an unsealed value
	si := CreateSealedItem(secretValue)

	assert.Equal(t, false, si.IsValueSealed())

	//get a sealer to seal it
	sealer := MakeSecretSealer()

	//try to seal without loading the public key
	err := si.Seal(sealer)
	assert.ErrorContains(t, err, "failed to seal secret: no public key set")

	//now load a key in the sealer
	err = sealer.LoadPublicKeyFromFile(TEST_FILE_PREFIX + PUBLIC_KEY_4096_FILENAME)
	assert.Nil(t, err)

	//seal the valid with the valid sealer
	err = si.Seal(sealer)
	assert.Nil(t, err)

	assert.Equal(t, true, si.IsValueSealed())

	//now corrupt the sealed value
	corruptSealedValue := "+aaaaaaa=="
	si.value = corruptSealedValue

	//make an unsealer with a private key
	unsealer := MakeSecretUnsealer()
	unsealer.LoadPrivateKeyFromFile(TEST_FILE_PREFIX+PRIVATE_KEY_4096_FILENAME, "")

	{
		//check the sealed item with the key
		sealCheckResult := si.CheckSeals(unsealer)
		assert.Equal(t, 0, sealCheckResult.UnsealedCount)
		assert.Equal(t, 1, sealCheckResult.SealedCount)
		assert.Len(t, sealCheckResult.UnsealErrors, 1)
		assert.ErrorContains(t, sealCheckResult.UnsealErrors[0], "crypto/rsa: decryption error")
	}

	//now try to unseal it
	err = si.Unseal(unsealer)
	assert.ErrorContains(t, err, "crypto/rsa: decryption error")

	//it should still be sealed and the corrupted value unchanged
	assert.True(t, si.IsValueSealed())
	assert.Equal(t, corruptSealedValue, si.value)

	{
		//check the item with the key again
		sealCheckResult := si.CheckSeals(unsealer)
		assert.Equal(t, 0, sealCheckResult.UnsealedCount)
		assert.Equal(t, 1, sealCheckResult.SealedCount)
		assert.Len(t, sealCheckResult.UnsealErrors, 1)
		assert.ErrorContains(t, sealCheckResult.UnsealErrors[0], "crypto/rsa: decryption error")
	}

}
