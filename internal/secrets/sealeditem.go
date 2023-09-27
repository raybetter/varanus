package secrets

import (
	"fmt"
	"regexp"
	"strings"
	"varanus/internal/validation"

	"gopkg.in/yaml.v3"
)

// note that inner group corresponds to internal/secrets/secretmgr.SealedValueRegex
var sealedWrapperRegex = regexp.MustCompile(`^sealed\((.*)\)$`)

func wrapSealedString(value string) string {
	return fmt.Sprintf("sealed(%s)", value)
}

func processSealedItemString(value string) (string, bool) {
	matchVals := sealedWrapperRegex.FindStringSubmatch(value)
	if matchVals == nil {
		return value, false
	} else {
		//the value should be the first group out of the regex
		//make sure it matches the secret form first though
		return matchVals[1], true
	}
}

type SealedItem struct {
	value    string
	isSealed bool
}

func createSealedItemImpl(value string) SealedItem {
	processedValue, isSealed := processSealedItemString(value)
	return SealedItem{processedValue, isSealed}
}

func (si SealedItem) IsValueSealed() bool {
	return si.isSealed
}

func (si SealedItem) GetValue() string {
	if si.isSealed {
		return wrapSealedString(si.value)
	}
	return si.value
}

func (si SealedItem) checkRawValue() error {
	if si.isSealed {
		if !SealedValueRegex.MatchString(si.value) {
			return fmt.Errorf("value does not match the expected format for an encrypted, encoded string")
		}
	} else {
		//unsealed cannot be empty
		if len(si.value) == 0 {
			return fmt.Errorf("SealedItem with an unsealed value should not be empty")
		}
	}
	return nil
}

// implement validation.Validatable
func (si SealedItem) Validate(vp *validation.ValidationProcess) error {
	err := si.checkRawValue()
	if err != nil {
		vp.AddValidationError(si, err.Error())
	}
	return nil
}

// Seal seals the secret in the sealed value using the supplied SecretSealer.  Calls to Seal are
// idempotent -- if the item is already sealed, then nothing happens.
func (si *SealedItem) Seal(sealer SecretSealer) error {
	if si.isSealed {
		return nil
	}
	sealedValue, err := sealer.SealSecret(si.value)
	if err != nil {
		return fmt.Errorf("failed to seal secret: %w", err)
	}
	si.value = sealedValue
	si.isSealed = true
	return nil
}

// SealValue seals the secret in the sealed value using the supplied SecretSealer.  Calls to SealValue are
// idempotent -- if the item is already sealed, then nothing happens.
func (si *SealedItem) Unseal(unsealer SecretUnsealer) error {
	if !si.isSealed {
		return nil
	}
	unsealedValue, err := unsealer.UnsealSecret(si.value)
	if err != nil {
		return fmt.Errorf("failed to unseal secret %w", err)
	}
	si.value = unsealedValue
	si.isSealed = false
	return nil
}

// CheckSeal returns an error if the item is sealed and cannot be unsealed by unsealer.  If the
// unsealer is nil or the item is not sealed, then nothing is returned.
func (si SealedItem) Check(unsealer SecretUnsealer) error {

	if unsealer != nil {
		if si.isSealed {
			//check that we can unseal the secret
			_, err := unsealer.UnsealSecret(si.value)
			return err
		}
	} else {
		//if no unsealer provided, fall back to raw item checks
		return si.checkRawValue()
	}
	return nil
}

func (si SealedItem) String() string {
	return si.GoString()
}

func (si SealedItem) GoString() string {
	if si.isSealed {
		return fmt.Sprintf(`secrets.SealedItem{value:"%s", isSealed:%t}`, si.value, si.isSealed)
	} else {
		return fmt.Sprintf(`secrets.SealedItem{value:"<unsealed value redacted>", isSealed:%t}`, si.isSealed)
	}
}

// ------------------------------- YAML marshaling and unmarshaling --------------------------------

func (si SealedItem) MarshalYAML() (interface{}, error) {
	return si.GetValue(), nil
}

func (si *SealedItem) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected a scalar value")
	}
	processedValue, isSealed := processSealedItemString(value.Value)
	si.value = processedValue
	si.isSealed = isSealed

	return nil
}

// ------------------------------- JSON marshaling and unmarshaling --------------------------------

func (si SealedItem) MarshalJSON() ([]byte, error) {
	return []byte(`"` + si.GetValue() + `"`), nil
}

func (si *SealedItem) UnmarshalJSON(valueByte []byte) error {
	value := string(valueByte)
	value = strings.Trim(value, `"`)

	processedValue, isSealed := processSealedItemString(value)
	si.value = processedValue
	si.isSealed = isSealed

	return nil
}
