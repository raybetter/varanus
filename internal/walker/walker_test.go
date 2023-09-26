package walker

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type SpecialObjectInterface interface {
	AddValue(string)
	GetValues() string
}

type SpecialObject struct {
	values []string
}

func (so SpecialObject) GetValues() string {
	return strings.Join(so.values, ",")
}

func (so *SpecialObject) AddValue(newValue string) {
	so.values = append(so.values, newValue)
}

func (so *SpecialObject) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected a scalar value")
	}
	so.values = strings.Split(value.Value, ",")

	return nil
}

func (so SpecialObject) MarshalYAML() (interface{}, error) {
	return strings.Join(so.values, ","), nil
}

type TopObject struct {
	FieldA IntermediateA  `yaml:"field_a"`
	FieldB IntermediateB  `yaml:"field_b"`
	FieldC *IntermediateC `yaml:"field_c"`
}

type IntermediateA struct {
	StrVal string        `yaml:"str_val"`
	IntVal int           `yaml:"int_val"`
	SOVal  SpecialObject //note no YAML tag -- this gets serialized as soval
}

type IntermediateB struct {
	AList []IntermediateA          `yaml:"a_list"`
	AMap  map[string]IntermediateA `yaml:"a_map"`
}

type IntermediateC struct {
	FieldD IntermediateD  `yaml:"field_d"`
	FieldE *IntermediateE `yaml:"field_e"`
	EmptyE *IntermediateE //we will let this be empty by not putting a value in the YAML string
}

type IntermediateD struct {
	StrVal    string          `yaml:"str_val"`
	IntVal    int             `yaml:"int_val"`
	SOValList []SpecialObject `yaml:"so_val_list"`
}

type IntermediateE struct {
	StrVal string         `yaml:"str_val"`
	IntVal int            `yaml:"int_val"`
	SOVal  *SpecialObject `yaml:"so_val"`
}

func loadTestObject(t *testing.T, yamlStr string) *TopObject {
	object := &TopObject{}
	err := yaml.Unmarshal([]byte(yamlStr), &object)
	require.Nil(t, err)

	return object
}

const yamlStr = `---
field_a:
  str_val: string at field_a.str_val
  int_val: 1
  soval: "a1,a2,a3"
field_b:
  a_list:
    - str_val: string at field_b.a_list[0].str_val
      int_val: 2
      soval: "ba1-1,ba1-2"
    - str_val: string at field_b.a_list[1].str_val
      int_val: 3
      soval: "ba2-1,ba2-2"
  a_map:
    foo:
      str_val: string at field_b.a_map[foo].str_val
      int_val: 4
      soval: "ba-foo-1"
    bar:
      str_val: string at field_b.a_map[bar].str_val
      int_val: 4
      soval: "ba-bar-1"
field_c:
  field_d:
    str_val: string at field_c.field_d.str_val
    int_val: 1
    so_val_list: 
      - "cd-so1-1,cd-so1-2,cd-so1-3"
      - "cd-so2-1,cd-so2-2,cd-so2-3"
      - "cd-so3-1,cd-so3-2,cd-so3-3"
  field_e:
    str_val: string at field_c.field_e.str_val
    int_val: 1
    so_val: "ce-so-1,ce-so-2,ce-so-3"
`

func TestWalkerMutable(t *testing.T) {

	expectedCallbackSequence := []string{
		"a1,a2,a3",
		"ba1-1,ba1-2",
		"ba2-1,ba2-2",
		"ba-foo-1",
		"ba-bar-1",
		"cd-so1-1,cd-so1-2,cd-so1-3",
		"cd-so2-1,cd-so2-2,cd-so2-3",
		"cd-so3-1,cd-so3-2,cd-so3-3",
		"ce-so-1,ce-so-2,ce-so-3",
	}

	expectedPathSequence := []string{
		// note that this and the next few are SOVal because that's the struct field name and there
		// is no YAML tag to override it.
		"field_a.SOVal",
		"field_b.a_list[0].SOVal",
		"field_b.a_list[1].SOVal",
		"field_b.a_map[foo].SOVal",
		"field_b.a_map[bar].SOVal",
		"field_c.field_d.so_val_list[0]",
		"field_c.field_d.so_val_list[1]",
		"field_c.field_d.so_val_list[2]",
		"field_c.field_e.so_val",
	}

	callbackSequence := []string{}
	pathSequence := []string{}

	//setup the callback for the test
	testCallback := func(needle interface{}, path string) error {
		soVal, ok := needle.(*SpecialObject)
		require.True(t, ok)

		//record the sequence of callbacks
		callbackSequence = append(callbackSequence, soVal.GetValues())
		pathSequence = append(pathSequence, path)

		//modify the object by adding a value
		soVal.AddValue("final")

		return nil
	}

	//load the config object for the test
	object := loadTestObject(t, yamlStr)

	// fmt.Printf("%# v", pretty.Formatter(object))

	// walk the config object
	err := WalkObjectMutable(object, reflect.TypeOf(SpecialObject{}), testCallback)
	assert.Nil(t, err)

	//check results
	assert.Equal(t, expectedCallbackSequence, callbackSequence)
	assert.Equal(t, expectedPathSequence, pathSequence)

	// check for modified values
	hasFinalValue := func(so *SpecialObject, expected string) {
		lastValue := so.values[len(so.values)-1]
		assert.Equal(t, expected, lastValue, "in SO %s", so)
	}

	hasFinalValue(&object.FieldA.SOVal, "final")
	hasFinalValue(&object.FieldB.AList[0].SOVal, "final")
	hasFinalValue(&object.FieldB.AList[1].SOVal, "final")

	{
		mapVal := object.FieldB.AMap["foo"]
		hasFinalValue(&mapVal.SOVal, "final")
	}
	{
		mapVal := object.FieldB.AMap["bar"]
		hasFinalValue(&mapVal.SOVal, "final")
	}

	hasFinalValue(&object.FieldC.FieldD.SOValList[0], "final")
	hasFinalValue(&object.FieldC.FieldD.SOValList[1], "final")
	hasFinalValue(&object.FieldC.FieldD.SOValList[2], "final")
	hasFinalValue(object.FieldC.FieldE.SOVal, "final")

}

func TestWalkerImmutable(t *testing.T) {

	expectedCallbackSequence := []string{
		"a1,a2,a3",
		"ba1-1,ba1-2",
		"ba2-1,ba2-2",
		"ba-foo-1",
		"ba-bar-1",
		"cd-so1-1,cd-so1-2,cd-so1-3",
		"cd-so2-1,cd-so2-2,cd-so2-3",
		"cd-so3-1,cd-so3-2,cd-so3-3",
		"ce-so-1,ce-so-2,ce-so-3",
	}

	expectedPathSequence := []string{
		// note that this and the next few are SOVal because that's the struct field name and there
		// is no YAML tag to override it.
		"field_a.SOVal",
		"field_b.a_list[0].SOVal",
		"field_b.a_list[1].SOVal",
		"field_b.a_map[foo].SOVal",
		"field_b.a_map[bar].SOVal",
		"field_c.field_d.so_val_list[0]",
		"field_c.field_d.so_val_list[1]",
		"field_c.field_d.so_val_list[2]",
		"field_c.field_e.so_val",
	}

	callbackSequence := []string{}
	pathSequence := []string{}

	//setup the callback for the test
	testCallback := func(needle interface{}, path string) error {
		//since this is the immutable, cast to the object itself
		soVal, ok := needle.(SpecialObject)
		require.True(t, ok)

		//record the sequence of callbacks
		callbackSequence = append(callbackSequence, soVal.GetValues())
		pathSequence = append(pathSequence, path)

		//modify the object by adding a value -- this change should be lost
		soVal.AddValue("final")

		return nil
	}

	//load the config object for the test
	object := loadTestObject(t, yamlStr)

	// fmt.Printf("%# v", pretty.Formatter(object))

	// walk the config object (not the pointer)
	err := WalkObjectImmutable(*object, reflect.TypeOf(SpecialObject{}), testCallback)
	assert.Nil(t, err)

	//check results
	assert.Equal(t, expectedCallbackSequence, callbackSequence)
	assert.Equal(t, expectedPathSequence, pathSequence)

	// check for modified values
	hasNoFinalValue := func(so *SpecialObject, expected string) {
		lastValue := so.values[len(so.values)-1]
		assert.NotEqual(t, expected, lastValue, "in SO %s", so)
	}

	hasNoFinalValue(&object.FieldA.SOVal, "final")
	hasNoFinalValue(&object.FieldB.AList[0].SOVal, "final")
	hasNoFinalValue(&object.FieldB.AList[1].SOVal, "final")

	{
		mapVal := object.FieldB.AMap["foo"]
		hasNoFinalValue(&mapVal.SOVal, "final")
	}
	{
		mapVal := object.FieldB.AMap["bar"]
		hasNoFinalValue(&mapVal.SOVal, "final")
	}

	hasNoFinalValue(&object.FieldC.FieldD.SOValList[0], "final")
	hasNoFinalValue(&object.FieldC.FieldD.SOValList[1], "final")
	hasNoFinalValue(&object.FieldC.FieldD.SOValList[2], "final")
	hasNoFinalValue(object.FieldC.FieldE.SOVal, "final")

}

func TestWalkerMutableCallWithImmutableObject(t *testing.T) {

	//setup the callback for the test
	testCallback := func(needle interface{}, path string) error {
		return nil
	}

	//load the config object for the test
	object := loadTestObject(t, yamlStr)

	//call WalkObjectMutable with the object (not the pointer) should fail
	err := WalkObjectMutable(*object, reflect.TypeOf(SpecialObject{}), testCallback)
	assert.ErrorContains(t, err, "found a value of type walker.SpecialObject, but it is not settable")

}

func TestWalkerCallbackErrors(t *testing.T) {

	//setup the callback for the test
	testCallback := func(needle interface{}, path string) error {
		return fmt.Errorf("test error")
	}

	//load the config object for the test
	object := loadTestObject(t, yamlStr)

	{
		//see the propagating error in WalkObjectMutable
		err := WalkObjectMutable(object, reflect.TypeOf(SpecialObject{}), testCallback)
		assert.ErrorContains(t, err, "callback error at path")
		assert.ErrorContains(t, err, ": test error")
	}

	{
		//see the propagating error in WalkObjectImmutable
		err := WalkObjectImmutable(*object, reflect.TypeOf(SpecialObject{}), testCallback)
		assert.ErrorContains(t, err, "callback error at path")
		assert.ErrorContains(t, err, ": test error")
	}

}

type InterfaceContainer struct {
	SOI1 SpecialObjectInterface
	SOI2 SpecialObjectInterface
	SOI3 SpecialObjectInterface
}

type OtherSpecialObject struct {
	values []string
}

func (oso *OtherSpecialObject) AddValue(value string) {
	oso.values = append(oso.values, value)
}

func (oso OtherSpecialObject) GetValues() string {
	return "oso::" + strings.Join(oso.values, ",")
}

func TestInterfaceFields(t *testing.T) {
	so1 := SpecialObject{[]string{"foo1", "bar1"}}
	so2 := &SpecialObject{[]string{"foo2", "bar2"}}
	oso3 := OtherSpecialObject{[]string{"foo3", "bar3"}}

	ic := InterfaceContainer{
		SOI1: &so1,
		SOI2: so2,
		SOI3: &oso3,
	}

	expectedCallbackSequence := []string{
		"foo1,bar1",
		"foo2,bar2",
		"oso::foo3,bar3",
	}
	expectedPathSequence := []string{
		"SOI1",
		"SOI2",
		"SOI3",
	}

	callbackSequence := []string{}
	pathSequence := []string{}

	//setup the callback for the test
	testCallback := func(needle interface{}, path string) error {
		pathSequence = append(pathSequence, path)

		val, ok := needle.(SpecialObjectInterface)
		require.True(t, ok)
		callbackSequence = append(callbackSequence, val.GetValues())

		return nil
	}

	soiType := reflect.TypeOf((*SpecialObjectInterface)(nil)).Elem()

	err := WalkObjectImmutable(ic, soiType, testCallback)
	assert.Nil(t, err)

	//check results
	assert.Equal(t, expectedCallbackSequence, callbackSequence)
	assert.Equal(t, expectedPathSequence, pathSequence)
}
