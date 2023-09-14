package config

import (
	"fmt"
	"regexp"
	"varanus/internal/secrets"

	"gopkg.in/yaml.v3"
)

type SealedItem struct {
	SealedValue string
}

func (si SealedItem) Validate(vp *ValidationProcess) error {
	if !secrets.SealedValueRegex.MatchString(si.SealedValue) {
		vp.AddValidationError(si, "value does not match the expected format for an encrypted, encoded string")
	}
	return nil
}

func (si *SealedItem) MarshalYAML() (interface{}, error) {
	return fmt.Sprintf("sealed(%s)", si.SealedValue), nil
}

// note that inner group corresponds to internal/secrets/secretmgr.SealedValueRegex
var sealedRegex = regexp.MustCompile(`^sealed\(([A-Za-z0-9+/]+=+)\)$`)

func (si *SealedItem) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected a scalar value")
	}
	matchVals := sealedRegex.FindStringSubmatch(value.Value)
	if matchVals == nil {
		return fmt.Errorf("value does not have the form of: 'sealed(<encrypted_encoded_string>)'.  Was it sealed with the Varanus sealing tool?")
	}
	//if we get here, get the sealed value out of the regex
	si.SealedValue = matchVals[1]
	return nil
}
