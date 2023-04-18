package poly

import (
	"encoding/json"
	"math"
	"reflect"
	"sort"
)

type IndexGettable interface {
	GetIndex() int
}

// MarshallPoly is a function that takes an input object of any type and
// serializes it into a JSON byte array. The function flattens the input object
// by extracting its fields and appending them to a slice. For fields of slice
// types, the function appends individual non-zero elements of the slice to the
// resulting flattened slice. The function also sorts the flattened slice based
// on the index of the elements if they implement the IndexGettable interface
// which can be used to control the ordering of the resultant JSON objects.
// If there are multiple objects that have the same index, they are sorted together
// with the internal order based on when they were first encountered. Any
// object that does not implement the IndexGettable interface will be sorted to
// the end using the same rules.
//
// Note that this only serialized the objects to JSON using the default JSON
// marshalling. This means that if you want to have a JSON value that can be used
// to determine the type of object when the JSON is being deserialized, the
// appropriate field must be present. This does not magically add any
// type resolution to the output JSON objects.
//
// Parameters:
// - obj (any): The input object to be serialized. This can be a
// value or a pointer of any type.
//
// Returns:
// - ([]byte, error): A JSON byte array representing the serialized
// input object, and an error if any occurs during the marshalling process.
//
// The function is designed to be flexible and support a wide range of input
// types. The resulting JSON byte array is a representation of the input object's
// fields and their values, with nested structures being flattened and sorted
// based on the index provided by the IndexGettable interface. This function is
// useful for situations where a more compact or custom JSON representation is
// desired for complex data structures.
func MarshallPoly(obj any) ([]byte, error) {
	var flattenedObjs []any

	sourceType := reflect.TypeOf(obj)
	sourceValue := reflect.ValueOf(obj)

	if sourceType.Kind() == reflect.Pointer {
		sourceType = sourceType.Elem()
		sourceValue = sourceValue.Elem()
	}

	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)
		fieldType := field.Type
		fieldValue := sourceValue.Field(i)

		zeroObj := false
		if fieldValue.IsZero() {
			zeroObj = true
		}

		if fieldType.Kind() == reflect.Struct {
			// If we have a concrete object, that may cause issues
			// for trying to convert that to a IndexGettable if the
			// function takes a receiver pointer. Convert this to
			// a pointer to object, which accounts for both pointer
			// and non-pointer receiver implementations.
			ptrValue := reflect.New(fieldType)
			ptrValue.Elem().Set(fieldValue)
			fieldValue = ptrValue
			fieldType = reflect.TypeOf(fieldValue)
		}

		if fieldType.Kind() == reflect.Slice {
			for i := 0; i < fieldValue.Len(); i++ {
				sliceVal := fieldValue.Index(i)
				if !sliceVal.IsZero() {
					flattenedObjs = append(flattenedObjs, sliceVal.Interface())
				}
			}
		} else {
			if !zeroObj {
				flattenedObjs = append(flattenedObjs, fieldValue.Interface())
			}
		}
	}

	sort.SliceStable(flattenedObjs, func(i, j int) bool {
		ii := math.MaxInt
		ij := math.MaxInt
		if indexer, ok := flattenedObjs[i].(IndexGettable); ok {
			ii = indexer.GetIndex()
		}
		if indexer, ok := flattenedObjs[j].(IndexGettable); ok {
			ii = indexer.GetIndex()
		}
		return ii < ij
	})

	return json.Marshal(flattenedObjs)
}
