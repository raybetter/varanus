package walker

import (
	"fmt"
	"reflect"
	"strings"
)

var verbose = false

func SetVerbose(newval bool) {
	verbose = newval
}

func getYamlNameFromTag(tag reflect.StructTag) string {
	yamlTag := tag.Get("yaml")
	if yamlTag == "" {
		return ""
	}
	tokens := strings.Split(yamlTag, ",")
	return tokens[0]
}

func getFieldPathName(field reflect.StructField, currentType reflect.Type, currentValue reflect.Value) string {
	yamlString := getYamlNameFromTag(field.Tag)
	if yamlString != "" {
		return yamlString
	}
	//TODO support JSON, other field specs
	//default to the field name if no other name matchs
	return field.Name
}

func addFieldToPath(path string, field string) string {
	if path == "" {
		return field
	} else {
		return path + "." + field
	}
}

// WalkObjectMutable uses reflection to walk the supplied haystack object looking for objects with
// the same type as needle.  When a needle-typed object is found, the callback is called with a
// **pointer** to the needle object and a path string describing the objects location in the
// heirarchy.
//
// The haystack must be a pointer to an object.  The function retuns an error if any part of the
// structure is not settable.  Any error returned by the callback function will abort the walk and
// return the error.
//
// WalkObjectMutable should be used when the callback wants to modify the structure being walked.
// In particular, if the object heirarchy contains a map, the map values will be replace with new
// values (this is the only way to get a settable value out of a map).
//
// If the walk is for a read operation, WalkObjectImmutable can be used instead and will not modify
// the structure.
func WalkObjectMutable(
	haystack interface{},
	needle reflect.Type,
	callback func(interface{}, string) error,
) error {
	return walkObjectImplementation(haystack, needle, callback, true)
}

// WalkObjectImmutable uses reflection to walk the supplied haystack object looking for objects with
// the same type as needle.  When a needle-typed object is found, the callback is called with a copy
// of the needle object and a path string describing the objects location in the heirarchy.
//
// This walk should be used for read operations only.  Modifications to the needle object by the
// callback function will be lost.
//
// Any error returned by the callback function will abort the walk and return the error.
//
// If the walk is to modify the structure, then use WalkObjectMutable instead.
func WalkObjectImmutable(
	haystack interface{},
	needle reflect.Type,
	callback func(interface{}, string) error,
) error {
	return walkObjectImplementation(haystack, needle, callback, false)
}

func isNeedleType(currentType reflect.Type, currentValue reflect.Value, needleType reflect.Type, isMutable bool) bool {
	if currentType == needleType {
		return true
	}
	if needleType.Kind() == reflect.Interface {
		//for interface types, try testing if the value implements the interface
		// if verbose {
		// 	fmt.Println("For isNeedleType with interface needle, current type is", currentType, "interface type is", needleType)
		// }
		if currentType.Implements(needleType) {
			return true
		}
	}
	return false
}

// walkObjectImplementation provides the flow for both the immutable and mutable walk calls.
// Since the logic is complex and similar for the two cases, it's easier to maintain on
// implementation and switch behaviors on isMutable.
func walkObjectImplementation(
	haystack interface{},
	needleType reflect.Type,
	callback func(interface{}, string) error,
	isMutable bool,
) error {

	//declare recursive function
	var process func(reflect.Value, string) error

	//define recursive function
	process = func(currentValue reflect.Value, path string) error {

		currentType := currentValue.Type()
		if verbose {
			fmt.Printf("path: %s type: %s value %s\n", path, currentType, currentValue)
		}

		if isNeedleType(currentType, currentValue, needleType, isMutable) {
			// found the object type we were looking for
			var err error

			//if our needle type is a pointer, skip null pointers
			if currentType.Kind() == reflect.Pointer {
				if currentValue.IsNil() {
					return nil
				}
			}

			if isMutable {
				//if mutable, expect the needle value to be settable
				if !currentValue.CanSet() && currentType.Kind() != reflect.Pointer {
					return fmt.Errorf(
						"at path=%s, value=%s, found a value of type %s, but it is not settable",
						path, currentValue, needleType,
					)
				}
				if needleType.Kind() == reflect.Interface {
					//if the needle type is already an interface, don't get the pointer
					err = callback(currentValue.Interface(), path)
				} else {
					// otherwise, this is a mutable walk, so pass a pointer to the currentValue and path to the callback
					err = callback(currentValue.Addr().Interface(), path)
				}
			} else {
				// immutable, so pass the currentValue and path to the callback
				err = callback(currentValue.Interface(), path)
			}
			if err != nil {
				return fmt.Errorf("callback error at path=%s, type=%s, value=%s: %w",
					path, currentType, currentValue, err)
			}

			//if we get here, the needle callback finished.
			// If the current type is a pointer, derefernce it here so that we skip the pointer
			// process logic, otherwise we might call the callback a second time on the Elem
			//value of the pointer.
			if currentType.Kind() == reflect.Pointer {
				currentValue = currentValue.Elem()
				currentType = currentValue.Type()
				// fmt.Printf("after ptr deref for needle type, path: %s type: %s value %s\n", path, currentType, currentValue)
			}
		}
		if currentType.Kind() == reflect.Pointer {
			//stop at nil pointers
			if currentValue.IsNil() {
				return nil
			}

			innerValue := currentValue.Elem()
			// innerType := innerValue.Type()

			err := process(innerValue, path)
			if err != nil {
				return err
			}

		}
		if currentType.Kind() == reflect.Struct {
			for _, field := range reflect.VisibleFields(currentType) {
				//skip exported fields
				if !field.IsExported() {
					continue
				}
				fieldValue := currentValue.FieldByName(field.Name)
				fieldName := getFieldPathName(field, currentType, currentValue)

				if isMutable && fieldValue.CanAddr() {
					//for mutable walks, pass the pointer to field values
					err := process(fieldValue.Addr(), addFieldToPath(path, fieldName))
					if err != nil {
						return err
					}
				} else {
					err := process(fieldValue, addFieldToPath(path, fieldName))
					if err != nil {
						return err
					}
				}
			}
			return nil
		}
		if currentType.Kind() == reflect.Slice {
			for index := 0; index < currentValue.Len(); index++ {
				err := process(
					currentValue.Index(index),
					fmt.Sprintf("%s[%d]", path, index),
				)
				if err != nil {
					return err
				}
			}
			return nil
		}
		if currentType.Kind() == reflect.Map {
			iter := currentValue.MapRange()
			for iter.Next() {
				k := iter.Key()
				v := iter.Value()

				if isMutable {
					//since map values are not addressable, we make a new object of the value type
					//Note that reflect.New() already returns a pointer to v.Type()
					newVPtr := reflect.New(v.Type())
					//set the value from map onto our new value
					newVPtr.Elem().Set(v)

					//call process with our addressable copy
					err := process(newVPtr, fmt.Sprintf("%s[%s]", path, k))
					if err != nil {
						return err
					}

					//assign the copy back to the map
					currentValue.SetMapIndex(k, newVPtr.Elem())
				} else {
					//for immutable walks, use the map value directly because we will not modify it
					err := process(v, fmt.Sprintf("%s[%s]", path, k))

					if err != nil {
						return err
					}
				}
			}
			return nil
		}

		//if we get here, it's a non-matching single value type we don't need to walk further
		return nil

	}

	// fmt.Println("Needle type:", needle)

	err := process(reflect.ValueOf(haystack), "")
	if err != nil {
		return err
	}
	return nil

}
