try making the top level config object a map of things that are registered from the modules, rather than known to the config at comiple time

config.go

```go

package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ConfigMap map[string]interface{}

type ConfigGenerator func() interface{}

var configGeneratorMap = map[string]ConfigGenerator{}

// func (c *ConfigMap) MarshalYAML() (interface{}, error) {
// 	return si.GetValue(), nil
// }

func (c *ConfigMap) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping value")
	}
	

	return nil
}

func RegisterConfigGenerator(name string, generator ConfigGenerator) error {
	_, keyExists := configGeneratorMap[name]
	if keyExists {
		return fmt.Errorf("a config named '%s' has already been registered", name)
	}
	configGeneratorMap[name] = generator
	return nil
}

// ReadConfigFromFile creates a config object from the file at filename.
func ReadConfig(data []byte) (*ConfigMap, error) {
	configMap := ConfigMap{}

	for name, generator := range configGeneratorMap {
		configMap[name] = generator()
	}

	//now that we have a config map, load it with yaml data from the data source
	err := yaml.Unmarshal(data, &configMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return &configMap, nil
}


```

config_test.go

```go

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type ConfigTop1 struct {
	MiddleA *ConfigMiddle1 `yaml:"middle_a"`
	MiddleB *ConfigMiddle1 `yaml:"middle_b"`
}

type ConfigMiddle1 struct {
	BottomA ConfigBottom1 `yaml:"bottom_a"`
	BottomB ConfigBottom1 `yaml:"bottom_b"`
}

type ConfigBottom1 struct {
	SVal string `yaml:"sval"`
	IVal int    `yaml:"ival"`
}

type ConfigTop2 struct {
	MiddleC *ConfigMiddle2 `yaml:"middle_c"`
	MiddleD *ConfigMiddle2 `yaml:"middle_d"`
}

type ConfigMiddle2 struct {
	BottomC ConfigBottom2 `yaml:"bottom_c"`
	BottomD ConfigBottom2 `yaml:"bottom_d"`
}

type ConfigBottom2 struct {
	FVal float64 `yaml:"fval"`
	SVal string  `yaml:"sval"`
}

func TestConfigMapE2E(t *testing.T) {

	func1 := func() interface{} {
		return &ConfigTop1{}
	}
	func2 := func() interface{} {
		return &ConfigTop2{}
	}

	RegisterConfigGenerator("foo", func1)
	RegisterConfigGenerator("bar", func2)

	ydata := `---
foo:
  middle_a:
    bottom_a:
      ival: 10
      sval: foo aa string
    bottom_b:
      ival: 20
      sval: foo ab string
  middle_b:
    bottom_a:
      ival: 30
      sval: foo ba string
    bottom_b:
      ival: 40
      sval: foo bb string
bar:
  middle_c:
    bottom_c:
      fval: 50.5
      sval: bar cc string
    bottom_d:
      fval: 60.5
      sval: bar cd string
  middle_d:
    bottom_c:
      fval: 70.5
      sval: bar dc string
    bottom_d:
      fval: 80.5
      sval: bar dd string
`
	config, err := ReadConfig([]byte(ydata))
	assert.Nil(t, err)
	assert.Len(t, config, 2)

}

```