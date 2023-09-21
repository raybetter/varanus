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
	assert.Equal(t, "+abcdef+jk==", si.GetValue())

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
	//should not get any text from GetValue because we're not supposed to read unsealed secrets
	assert.Equal(t, "", si.GetValue())

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
	assert.ErrorContains(t, err, "attempt to marshal an unsealed SealedItem is not allowed")
	assert.Nil(t, yamlValue)

	jsonValue, err := si.MarshalJSON()
	assert.ErrorContains(t, err, "attempt to marshal an unsealed SealedItem is not allowed")
	assert.Equal(t, []byte{}, jsonValue)
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

var testKey string = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAybHNVmB0+9tNyZyKiMCJ
Qz0gUW/QtNdQd/XX26w6SCQtJVoM+o6r5vbz9YQoKHYe8vfeUdEmE79UorMMmndB
S8v8lMwWuCEy9MfcMRsSWnz8u9yRyhVfjaqEnMJHu2Pw05GLhQVf7fD+45eSTUZa
EenaOZUXQX9RVA2MEA4TuaIIAM7uMEbU2ta0zM8A9WathRkqxqNN/2l24Y3AjWek
xA1thE7wHvGtvhAO3v1S1GFbH/bbBoLSm3Ry+dZV8Hw+CK+h/soXzEjg7uIR67gW
SRZ3CPOGK2/0pQTLDMxQ9zCAzgAArMFAtjEe0Os51NgK5r170s00EY4mNTSE7285
5dLg+vJ3dcT5R1rbvElE3HI0JpmACNCGTxumML5f2GMiRgPyLsAbrOxDIhYessrj
QkZixmITW5dDvltbB/Rc8yojR3qvSe5SRD0kH/R2wikJnFA/rlQHWKR37e0/uMOu
cQGgeQB5EVF9Kskljo9VyPk7laqCJMMoZc1Ka21QhSLRDbNuXrNcfaDGDMJ5uk+w
3rDktpcFb/4cv5Jc+noMym+MiEZemvQz9cJjlBGdov/tPvjzaJERtbjrzSXQpO5f
C6CO4UwI/B/OEswbmNxW50Lh1rGQUrrVVSxpT2Co18xaAJO144cqkMO+UcDxWcgr
PFX6vkNXMsPZ4hxALqlxYZUCAwEAAQ==
-----END PUBLIC KEY-----`

func TestSealedItemSealWithKey(t *testing.T) {
	//make an unsealed value
	si := CreateSealedItem("unsealed value -- it's a secret!")

	assert.Equal(t, false, si.IsValueSealed())

	//get a sealer to seal it
	sealer := SecretSealer{}
	err := sealer.LoadPublicKey([]byte(testKey))
	assert.Nil(t, err)

	err = si.SealValue(&sealer)
	assert.Nil(t, err)

	assert.Equal(t, true, si.IsValueSealed())

	//validate the sealed value to make sure our validation works on an actual sealed value
	//should have no validation failures
	vp := validation.ValidationProcess{}

	err = si.Validate(&vp)
	assert.Nil(t, err)

	err = vp.GetFinalValidationError()
	assert.Nil(t, err)

}

func TestSealedItemSealWithKeyErrors(t *testing.T) {
	//make an unsealed value
	si := CreateSealedItem("unsealed value -- it's a secret!")

	assert.Equal(t, false, si.IsValueSealed())

	//get a sealer to seal it
	sealer := SecretSealer{}

	//try to seal without loading the public key
	err := si.SealValue(&sealer)
	assert.ErrorContains(t, err, "failed to seal secret no public key set")

	//now load a key in the sealer
	err = sealer.LoadPublicKey([]byte(testKey))
	assert.Nil(t, err)

	//seal the valid with the valid sealer
	err = si.SealValue(&sealer)
	assert.Nil(t, err)

	assert.Equal(t, true, si.IsValueSealed())

	//a second sealing should fail
	err = si.SealValue(&sealer)
	assert.ErrorContains(t, err, "value is already sealed")

}
