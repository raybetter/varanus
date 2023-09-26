package walker

import (
	"fmt"
	"reflect"
	"strings"
)

func getYamlNameFromTag(tag reflect.StructTag) string {
	yamlTag := tag.Get("yaml")
	if yamlTag == "" {
		return ""
	}
	tokens := strings.Split(yamlTag, ",")
	return tokens[0]
}

func getFieldPathName(fieldIndex int, targetType reflect.Type, targetValue reflect.Value) string {
	yamlString := getYamlNameFromTag(targetType.Field(fieldIndex).Tag)
	if yamlString != "" {
		return yamlString
	}
	//TODO support JSON, other field specs
	//default to the field name if no other name matchs
	return targetType.Field(fieldIndex).Name
}

func addFieldToPath(path string, field string) string {
	if path == "" {
		return field
	} else {
		return path + "." + field
	}
}

func WalkObject(haystack interface{}, needle reflect.Type, callback func(interface{}, string) error) error {

	p := reflect.ValueOf(haystack)
	fmt.Println("type of p:", p.Type())
	fmt.Println("settability of p:", p.CanSet())
	v := p.Elem()
	fmt.Println("settability of v:", v.CanSet())

	//declare recursive function
	var process func(interface{}, string) error

	//define recursive function
	process = func(haystack interface{}, path string) error {
		fmt.Println(path)

		targetType := reflect.TypeOf(haystack)
		targetValue := reflect.ValueOf(haystack)

		fmt.Println("path/CanSet/CanAddr", path, targetValue.CanSet(), targetValue.CanAddr())

		if targetType == needle {
			err := callback(targetValue.Interface(), path)
			if err != nil {
				return err
			}
			return nil
		}
		if targetType.Kind() == reflect.Pointer {
			if targetValue.IsNil() {
				//skip nil pointers
				return nil
			} else {
				//pointer doesn't add any path information, so just pass the path through
				innerValue := targetValue.Elem()
				return process(innerValue.Interface(), path)
			}
		}
		if targetType.Kind() == reflect.Struct {
			for index := 0; index < targetValue.NumField(); index++ {
				f := targetValue.Field(index)
				fieldName := getFieldPathName(index, targetType, targetValue)

				err := process(f.Interface(), addFieldToPath(path, fieldName))
				if err != nil {
					return err
				}
			}
			return nil
		}
		if targetType.Kind() == reflect.Slice {
			for index := 0; index < targetValue.Len(); index++ {
				err := process(
					targetValue.Index(index).Interface(),
					fmt.Sprintf("%s[%d]", path, index),
				)
				if err != nil {
					return err
				}
			}
			return nil
		}
		if targetType.Kind() == reflect.Map {
			iter := targetValue.MapRange()
			for iter.Next() {
				k := iter.Key()
				v := iter.Value()
				err := process(
					v.Interface(),
					fmt.Sprintf("%s[%s]", path, k),
				)
				if err != nil {
					return err
				}
			}
			return nil
		}

		//else it's a single value type we don't care about
		return nil

	}

	err := process(haystack, "")
	if err != nil {
		return err
	}
	return nil

}
