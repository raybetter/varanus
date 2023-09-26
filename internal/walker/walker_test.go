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
}

type SpecialObject struct {
	values []string
}

func (so *SpecialObject) String() string {
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

func TestWalker(t *testing.T) {

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
		soVal, ok := needle.(SpecialObject)
		require.True(t, ok)

		//record the sequence of callbacks
		callbackSequence = append(callbackSequence, soVal.String())
		pathSequence = append(pathSequence, path)

		//modify the object by adding a value
		soVal.AddValue("final")

		return nil
	}

	//load the config object for the test
	object := loadTestObject(t, yamlStr)

	// fmt.Printf("%# v", pretty.Formatter(object))

	// walk the config object
	err := WalkObject(object, reflect.TypeOf(SpecialObject{}), testCallback)
	assert.Nil(t, err)

	//check results
	assert.Equal(t, expectedCallbackSequence, callbackSequence)
	assert.Equal(t, expectedPathSequence, pathSequence)

	// check for modified values
	hasLastValue := func(so *SpecialObject, expected string) bool {
		return so.values[len(so.values)-1] == expected
	}

	assert.True(t, hasLastValue(&object.FieldA.SOVal, "final"))
	assert.True(t, hasLastValue(&object.FieldB.AList[0].SOVal, "final"))
	assert.True(t, hasLastValue(&object.FieldB.AList[1].SOVal, "final"))

	{
		mapVal := object.FieldB.AMap["foo"]
		assert.True(t, hasLastValue(&mapVal.SOVal, "final"))
	}
	{
		mapVal := object.FieldB.AMap["bar"]
		assert.True(t, hasLastValue(&mapVal.SOVal, "final"))
	}

	assert.True(t, hasLastValue(&object.FieldC.FieldD.SOValList[0], "final"))
	assert.True(t, hasLastValue(&object.FieldC.FieldD.SOValList[1], "final"))
	assert.True(t, hasLastValue(&object.FieldC.FieldD.SOValList[2], "final"))
	assert.True(t, hasLastValue(object.FieldC.FieldE.SOVal, "final"))

}

// func TestWalkerWithNilPtr(t *testing.T) {

// 	//this test omits FieldE in the yamlStr so that the pointer value will be nil

// 	const yamlStr = `---
//     field_a:
//       str_val: string at field_a.str_val
//       int_val: 1
//       soval: "a1,a2,a3"
//     field_b:
//       a_list:
//         - str_val: string at field_b.a_list[0].str_val
//           int_val: 2
//           soval: "ba1-1,ba1-2"
//         - str_val: string at field_b.a_list[1].str_val
//           int_val: 3
//           soval: "ba2-1,ba2-2"
//       a_map:
//         foo:
//           str_val: string at field_b.a_map[foo].str_val
//           int_val: 4
//           soval: "ba-foo-1"
//         bar:
//           str_val: string at field_b.a_map[bar].str_val
//           int_val: 4
//           soval: "ba-bar-1"
//     field_c:
//       field_d:
//         str_val: string at field_c.field_d.str_val
//         int_val: 1
//         so_val_list:
//           - "cd-so1-1,cd-so1-2,cd-so1-3"
//           - "cd-so2-1,cd-so2-2,cd-so2-3"
//           - "cd-so3-1,cd-so3-2,cd-so3-3"
// `

// 	expectedCallbackSequence := []string{
// 		"a1,a2,a3",
// 		"ba1-1,ba1-2",
// 		"ba2-1,ba2-2",
// 		"ba-foo-1",
// 		"ba-bar-1",
// 		"cd-so1-1,cd-so1-2,cd-so1-3",
// 		"cd-so2-1,cd-so2-2,cd-so2-3",
// 		"cd-so3-1,cd-so3-2,cd-so3-3",
// 	}

// 	expectedPathSequence := []string{
// 		// note that this and the next few are SOVal because that's the struct field name and there
// 		// is no YAML tag to override it.
// 		"field_a.SOVal",
// 		"field_b.a_list[0].SOVal",
// 		"field_b.a_list[1].SOVal",
// 		"field_b.a_map[foo].SOVal",
// 		"field_b.a_map[bar].SOVal",
// 		"field_c.field_d.so_val_list[0]",
// 		"field_c.field_d.so_val_list[1]",
// 		"field_c.field_d.so_val_list[2]",
// 	}

// 	callbackSequence := []string{}
// 	pathSequence := []string{}

// 	//setup the callback for the test
// 	testCallback := func(needle interface{}, path string) error {
// 		soVal, ok := needle.(SpecialObject)
// 		require.True(t, ok)

// 		callbackSequence = append(callbackSequence, soVal.String())
// 		pathSequence = append(pathSequence, path)

// 		// fmt.Println("**********************************")
// 		// fmt.Println(needle)
// 		// fmt.Println("**********************************")
// 		return nil
// 	}

// 	//load the config object for the test
// 	object := loadTestObject(t, yamlStr)

// 	// fmt.Printf("%# v", pretty.Formatter(object))

// 	// walk the config object
// 	err := WalkObject(object, reflect.TypeOf(SpecialObject{}), testCallback)
// 	assert.Nil(t, err)

// 	//check results
// 	assert.Equal(t, expectedCallbackSequence, callbackSequence)
// 	assert.Equal(t, expectedPathSequence, pathSequence)
// }

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

func TestInterfaceFields(t *testing.T) {
	// so1 := SpecialObject{[]string{"foo1", "bar1"}}
	// so2 := &SpecialObject{[]string{"foo2", "bar2"}}
	// oso3 := OtherSpecialObject{[]string{"foo3", "bar3"}}

	// ic := InterfaceContainer{
	// 	SOI1: &so1,
	// 	SOI2: so2,
	// 	SOI3: &oso3,
	// }
}
