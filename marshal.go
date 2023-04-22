package poly

import (
	"encoding/json"
	"math"
	"reflect"
	"sort"
)

// IndexGettable is an interface that can optionally be implemented by objects
// that can provide an index that will be used to control the ordering of the
// objects in the JSON array.
type IndexGettable interface {
	GetIndex() int
}

var indexGettableType = reflect.TypeOf([]IndexGettable{}).Elem()

// indexedObject is a wrapper around the value type that also contains
// the index of the object. This is used to sort the objects based on the index
// provided by the IndexGettable interface.
type indexedObject struct {
	Index int
	Value any
}

// Marshal takes an input object of any type and serializes it into a JSON
// byte array. The function flattens the input object by extracting its fields
// and appending them to a slice. For fields of slice types, the function appends
// individual non-zero elements of the slice to the resulting flattened slice.
// The function also sorts the flattened slice based on the index of the elements
// if they implement the IndexGettable interface which can be used to control the
// ordering of the resultant JSON objects. If there are multiple objects that
// have the same index, they are sorted together with the internal order based on
// when they were first encountered. Any object that does not implement the
// IndexGettable interface will be sorted to the end using the same rules.
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
func Marshal(obj any) ([]byte, error) {
	flattenedObjs := Flatten(obj)

	return json.Marshal(flattenedObjs)
}

// Flatten takes an input object of any type and flattens the input object by
// extracting its fields and appending them to a slice. For fields of slice
// types, the function appends individual non-zero elements of the slice to the
// resulting flattened slice. The function also sorts the flattened slice based
// on the index of the elements if they implement the IndexGettable interface
// which can be used to control the ordering of the resultant JSON objects. If
// there are multiple objects that have the same index, they are sorted together
// with the internal order based on when they were first encountered. Any object
// that does not implement the IndexGettable interface will be sorted to the end
// using the same rules.
//
// This does not marshal them into JSON, unlike Marshal, and can be used
// if there is a need to do any custom JSON serialization by your own code.
//
// As in Marshal, the objects, when serialized to JSON by your code will
// need to have a field indicating the polymorphic type if you need that represented
// in the output JSON.
//
// Parameters:
// - obj (any): The input object to be serialized. This can be a
// value or a pointer of any type.
//
// Returns:
// - ([]any): A flattened representation of the input object with all the
// fields of the original object returned as a slice.
func Flatten(obj any) []any {

	sourceType := reflect.TypeOf(obj)
	sourceValue := reflect.ValueOf(obj)

	if sourceType.Kind() == reflect.Pointer {
		sourceType = sourceType.Elem()
		sourceValue = sourceValue.Elem()
	}

	needToSort := false
	indexedObjects := make([]indexedObject, 0)

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
					indexedObject, itemSortable := indexedObjectForValue(sliceVal)
					needToSort = needToSort || itemSortable
					indexedObjects = append(indexedObjects, indexedObject)
				}
			}
		} else {
			if !zeroObj {
				indexedObject, itemSortable := indexedObjectForValue(fieldValue)
				needToSort = needToSort || itemSortable
				indexedObjects = append(indexedObjects, indexedObject)
			}
		}
	}

	if needToSort {
		sort.SliceStable(indexedObjects, func(i, j int) bool {
			return indexedObjects[i].Index < indexedObjects[j].Index
		})
	}

	var flattenedObjs []any
	for _, item := range indexedObjects {
		flattenedObjs = append(flattenedObjs, item.Value)
	}

	return flattenedObjs
}

// indexedObjectForValue takes a reflect.Value and returns a
// indexedObject object with the value and index of the object. If the
// object does not implement the IndexGettable interface, the index is set to
// MaxInt and the needToSort flag is set to false.
func indexedObjectForValue(sliceVal reflect.Value) (indexedObject, bool) {
	sortItem := indexedObject{
		Index: math.MaxInt,
		Value: sliceVal.Interface(),
	}
	needToSort := false
	if sliceVal.CanConvert(indexGettableType) {
		needToSort = true
		sortItem.Index = sliceVal.Convert(indexGettableType).Interface().(IndexGettable).GetIndex()
	}
	return sortItem, needToSort
}
