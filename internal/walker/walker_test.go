package walker

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

//**************************************************************************************************
//**************************************************************************************************
// Heirarchical evaluation of struct needle
//**************************************************************************************************
//**************************************************************************************************

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

func indexOf(haystack [][]string, needle []string) int {
	for index, value := range haystack {
		if slices.Equal(value, needle) {
			return index
		}
	}
	return -1
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

	//two possible expected sequences because the map values can be read in any order
	expectedCallbackSequence := [][]string{
		{
			"a1,a2,a3",
			"ba1-1,ba1-2",
			"ba2-1,ba2-2",
			"ba-foo-1",
			"ba-bar-1",
			"cd-so1-1,cd-so1-2,cd-so1-3",
			"cd-so2-1,cd-so2-2,cd-so2-3",
			"cd-so3-1,cd-so3-2,cd-so3-3",
			"ce-so-1,ce-so-2,ce-so-3",
		},
		{
			"a1,a2,a3",
			"ba1-1,ba1-2",
			"ba2-1,ba2-2",
			"ba-bar-1",
			"ba-foo-1",
			"cd-so1-1,cd-so1-2,cd-so1-3",
			"cd-so2-1,cd-so2-2,cd-so2-3",
			"cd-so3-1,cd-so3-2,cd-so3-3",
			"ce-so-1,ce-so-2,ce-so-3",
		},
	}

	expectedPathSequence := [][]string{
		// note that this and the next few are SOVal because that's the struct field name and there
		// is no YAML tag to override it.
		{
			"field_a.SOVal",
			"field_b.a_list[0].SOVal",
			"field_b.a_list[1].SOVal",
			"field_b.a_map[foo].SOVal",
			"field_b.a_map[bar].SOVal",
			"field_c.field_d.so_val_list[0]",
			"field_c.field_d.so_val_list[1]",
			"field_c.field_d.so_val_list[2]",
			"field_c.field_e.so_val",
		},
		{
			"field_a.SOVal",
			"field_b.a_list[0].SOVal",
			"field_b.a_list[1].SOVal",
			"field_b.a_map[bar].SOVal",
			"field_b.a_map[foo].SOVal",
			"field_c.field_d.so_val_list[0]",
			"field_c.field_d.so_val_list[1]",
			"field_c.field_d.so_val_list[2]",
			"field_c.field_e.so_val",
		},
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
	assert.Contains(t, expectedCallbackSequence, callbackSequence)
	assert.Contains(t, expectedPathSequence, pathSequence)
	assert.Equal(t,
		indexOf(expectedCallbackSequence, callbackSequence),
		indexOf(expectedPathSequence, pathSequence))

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

	//two possible expected sequences because the map values can be read in any order
	expectedCallbackSequence := [][]string{
		{
			"a1,a2,a3",
			"ba1-1,ba1-2",
			"ba2-1,ba2-2",
			"ba-foo-1",
			"ba-bar-1",
			"cd-so1-1,cd-so1-2,cd-so1-3",
			"cd-so2-1,cd-so2-2,cd-so2-3",
			"cd-so3-1,cd-so3-2,cd-so3-3",
			"ce-so-1,ce-so-2,ce-so-3",
		},
		{
			"a1,a2,a3",
			"ba1-1,ba1-2",
			"ba2-1,ba2-2",
			"ba-bar-1",
			"ba-foo-1",
			"cd-so1-1,cd-so1-2,cd-so1-3",
			"cd-so2-1,cd-so2-2,cd-so2-3",
			"cd-so3-1,cd-so3-2,cd-so3-3",
			"ce-so-1,ce-so-2,ce-so-3",
		},
	}

	expectedPathSequence := [][]string{
		{
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
		},
		{
			// note that this and the next few are SOVal because that's the struct field name and there
			// is no YAML tag to override it.
			"field_a.SOVal",
			"field_b.a_list[0].SOVal",
			"field_b.a_list[1].SOVal",
			"field_b.a_map[bar].SOVal",
			"field_b.a_map[foo].SOVal",
			"field_c.field_d.so_val_list[0]",
			"field_c.field_d.so_val_list[1]",
			"field_c.field_d.so_val_list[2]",
			"field_c.field_e.so_val",
		},
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
	assert.Contains(t, expectedCallbackSequence, callbackSequence)
	assert.Contains(t, expectedPathSequence, pathSequence)
	assert.Equal(t,
		indexOf(expectedCallbackSequence, callbackSequence),
		indexOf(expectedPathSequence, pathSequence))

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

	//this test hits each of the callbacks in sequence to ensure the error return through all the
	//different types in the walk

	targetPath := ""

	//setup the callback for the test
	testCallback := func(needle interface{}, path string) error {
		if path == targetPath {
			return fmt.Errorf("test error")
		}
		return nil
	}

	targetPaths := []string{
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

	for _, nextPath := range targetPaths {
		targetPath = nextPath

		//load the config object for the test
		object := loadTestObject(t, yamlStr)

		{
			//see the propagating error in WalkObjectMutable
			err := WalkObjectMutable(object, reflect.TypeOf(SpecialObject{}), testCallback)
			assert.ErrorContains(t, err, "callback error at path="+nextPath)
			assert.ErrorContains(t, err, ": test error")
		}

		{
			//see the propagating error in WalkObjectImmutable
			err := WalkObjectImmutable(*object, reflect.TypeOf(SpecialObject{}), testCallback)
			assert.ErrorContains(t, err, "callback error at path="+nextPath)
			assert.ErrorContains(t, err, ": test error")
		}

	}

}

//**************************************************************************************************
//**************************************************************************************************
// Basic interface needle
//**************************************************************************************************
//**************************************************************************************************

type SpecialObjectInterface interface {
	GetValues() string
}

type InterfaceContainer struct {
	SOI1 SpecialObjectInterface
	SOI2 SpecialObjectInterface
	SOI3 SpecialObjectInterface
	OSO1 OtherSpecialObject
	OSO2 *OtherSpecialObject
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
	oso1 := OtherSpecialObject{[]string{"foo-oso1", "bar-oso1"}}
	oso2 := OtherSpecialObject{[]string{"foo-oso2", "bar-oso2"}}

	ic := InterfaceContainer{
		SOI1: &so1,
		SOI2: so2,
		SOI3: &oso3,
		OSO1: oso1,
		OSO2: &oso2,
	}

	expectedCallbackSequence := []string{
		"foo1,bar1",
		"foo2,bar2",
		"oso::foo3,bar3",
		"oso::foo-oso1,bar-oso1",
		"oso::foo-oso2,bar-oso2",
	}
	expectedPathSequence := []string{
		"SOI1",
		"SOI2",
		"SOI3",
		"OSO1",
		"OSO2",
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

//**************************************************************************************************
//**************************************************************************************************
// Heirarchical evaluation of interface needle
//**************************************************************************************************
//**************************************************************************************************

type Checkable interface {
	Check() string
}

type TopCheckable struct {
	Middle1    MiddleCheckable
	Middle2    MiddleCheckable
	Middle3    MiddleUncheckable
	checkValue string
}

func (c TopCheckable) Check() string {
	return c.checkValue
}

type MiddleCheckable struct {
	Bottom1    BottomCheckable
	Bottom2    BottomUncheckable
	checkValue string
}

func (c MiddleCheckable) Check() string {
	return c.checkValue
}

type MiddleUncheckable struct {
	Bottom1 BottomCheckable
	Bottom2 BottomUncheckable
}

type BottomCheckable struct {
	checkValue string
}

func (c BottomCheckable) Check() string {
	return c.checkValue
}

type BottomUncheckable struct {
}

func TestHeirarchicalInterfaceNeedleImmutable(t *testing.T) {
	top := TopCheckable{
		Middle1: MiddleCheckable{
			Bottom1: BottomCheckable{
				checkValue: "bottom 11",
			},
			Bottom2:    BottomUncheckable{},
			checkValue: "middle 1",
		},
		Middle2: MiddleCheckable{
			Bottom1: BottomCheckable{
				checkValue: "bottom 21",
			},
			Bottom2:    BottomUncheckable{},
			checkValue: "middle 2",
		},
		Middle3: MiddleUncheckable{
			Bottom1: BottomCheckable{
				checkValue: "bottom 31",
			},
			Bottom2: BottomUncheckable{},
		},
		checkValue: "top",
	}

	expectedCheckSequence := []string{
		"top",
		"middle 1",
		"bottom 11",
		"middle 2",
		"bottom 21",
		"bottom 31",
	}
	expectedPathSequence := []string{
		"",
		"Middle1",
		"Middle1.Bottom1",
		"Middle2",
		"Middle2.Bottom1",
		"Middle3.Bottom1",
	}

	callbackSequence := []string{}
	pathSequence := []string{}

	//setup the callback for the test
	testCallback := func(needle interface{}, path string) error {
		pathSequence = append(pathSequence, path)

		val, ok := needle.(Checkable)
		require.True(t, ok)
		callbackSequence = append(callbackSequence, val.Check())

		return nil
	}

	checkableType := reflect.TypeOf((*Checkable)(nil)).Elem()

	err := WalkObjectImmutable(top, checkableType, testCallback)
	assert.Nil(t, err)

	//check results
	assert.Equal(t, expectedCheckSequence, callbackSequence)
	assert.Equal(t, expectedPathSequence, pathSequence)
}

func TestHeirarchicalInterfaceNeedleMutable(t *testing.T) {
	top := TopCheckable{
		Middle1: MiddleCheckable{
			Bottom1: BottomCheckable{
				checkValue: "bottom 11",
			},
			Bottom2:    BottomUncheckable{},
			checkValue: "middle 1",
		},
		Middle2: MiddleCheckable{
			Bottom1: BottomCheckable{
				checkValue: "bottom 21",
			},
			Bottom2:    BottomUncheckable{},
			checkValue: "middle 2",
		},
		Middle3: MiddleUncheckable{
			Bottom1: BottomCheckable{
				checkValue: "bottom 31",
			},
			Bottom2: BottomUncheckable{},
		},
		checkValue: "top",
	}

	expectedCheckSequence := []string{
		"top",
		"middle 1",
		"bottom 11",
		"middle 2",
		"bottom 21",
		"bottom 31",
	}
	expectedPathSequence := []string{
		"",
		"Middle1",
		"Middle1.Bottom1",
		"Middle2",
		"Middle2.Bottom1",
		"Middle3.Bottom1",
	}

	callbackSequence := []string{}
	pathSequence := []string{}

	//setup the callback for the test
	testCallback := func(needle interface{}, path string) error {
		pathSequence = append(pathSequence, path)

		val, ok := needle.(Checkable)
		require.True(t, ok)
		callbackSequence = append(callbackSequence, val.Check())

		return nil
	}

	checkableType := reflect.TypeOf((*Checkable)(nil)).Elem()

	err := WalkObjectMutable(&top, checkableType, testCallback)
	assert.Nil(t, err)

	//check results
	assert.Equal(t, expectedCheckSequence, callbackSequence)
	assert.Equal(t, expectedPathSequence, pathSequence)
}
