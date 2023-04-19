package poly

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// TypeLocator needs to be implemented by whatever pre-deserializing type that
// is used to determine what is the actual type that we should create for each
// sub-object of whatever JSON array we're unmarshalling.
// Whatever `struct` you make that implements this should have the correct
// JSON members that will be used to determine which object to create to properly
// unmarshall the object.
type TypeLocator interface {
	// TypeName returns the name of the generic type needed to satisfy the
	// polymorphic unmarshalling.
	TypeName() string
}

// typeLocatorType is the type of the above interface.
var typeLocatorType = reflect.TypeOf([]TypeLocator{}).Elem()

// GenericTypeLocator provides a default implementation of the TypeLocator that
// handles common cases.
type GenericTypeLocator struct {
	Type       string `json:"type,omitempty"`
	TypeAt     string `json:"@type,omitempty"`
	TypeCaps   string `json:"Type,omitempty"`
	TypeAtCaps string `json:"@Type,omitempty"`
}

// DefaultLocator is the type of the default TypeLocator that is used in the simpler
// implementation of the unmarshaller.
var DefaultLocator = reflect.TypeOf(GenericTypeLocator{})

// TypeName returns the name of the generic type represented by the receiver.
func (t *GenericTypeLocator) TypeName() string {
	if len(t.Type) > 0 {
		return t.Type
	}
	if len(t.TypeAt) > 0 {
		return t.TypeAt
	}
	if len(t.TypeCaps) > 0 {
		return t.TypeCaps
	}
	if len(t.TypeAtCaps) > 0 {
		return t.TypeAtCaps
	}
	return ""
}

// IndexSettable is an interface that you should implement if you need to know the
// index into the array of JSON sub-objects. This should be implemented by the
// types that are referred to by the TypeLocator.
type IndexSettable interface {
	// SetIndex is called with the zero-based index into the JSON sub-object.
	// In cases where there are sub-objects that cannot be unmarshalled, those
	// still count in the indexing.
	SetIndex(index int)
}

type fieldLookup struct {
	index     int
	fieldType reflect.Type
	rootType  reflect.Type
	kind      reflect.Kind
	ptr       bool
}

// Unmarshall is a convenience function that takes a raw JSON byte slice and a
// target any type variable, and unmarshalls the JSON into the target variable
// based on the default polymorphism rules. The target variable should be a
// struct with fields tagged with their respective polymorphic type names. The
// default polymorphism rules are defined by the DefaultLocator, which implements
// some common polymorphic type resolutions using "type", "@type", "Type", and
// "@Type" as the keys to determine the type of object based on the JSON. If your
// needs are different, you can define a custom TypeLocator which handles the
// type resolution in whatever way is needed for your application.
//
// This function is a wrapper around UnmarshallCustom, using the
// DefaultLocator for type resolution. If an error occurs during unmarshalling, it
// returns an error.
//
// Example usage:
//
//		type Dog struct { ... }
//		type Cat struct { ... }
//		type Owner struct { ... }
//
//		type Result struct {
//		    Dogs  []Dog `poly:"dog"`
//		    Cats  []Cat `poly:"cat"`
//	     	Owner Owner `poly:"owner"`
//		}
//
//		var result Result
//		err := Unmarshall(jsonData, &result)
//
// In this example, the Unmarshall function would unmarshall the JSON into the
// Result struct, populating the Dogs and Cats slices based on the polymorphic type
// names defined in the DefaultLocator struct.
func Unmarshall(rawJson []byte, target any) error {
	return UnmarshallCustom(rawJson, target, DefaultLocator)
}

// UnmarshallCustom takes a raw JSON byte slice, a target any type variable, and
// a typeLocator of type reflect.Type. It unmarshalls the JSON into the target
// variable based on the custom type polymorphism rules defined by the
// typeLocator. The target variable should be a struct with fields tagged with
// their respective polymorphic type names. The typeLocator should be a struct
// implementing the TypeLocator interface which returns a type name for the
// current object. If an error occurs during unmarshalling, it returns an error.
//
// Example usage:
//
//		type Dog struct { ... }
//		type Cat struct { ... }
//		type Owner struct { ... }
//
//		type AnimalTypeLocator struct { ... }
//		func (tl *AnimalTypeLocator) TypeName() string { ... }
//
//		type Result struct {
//		    Dogs  []Dog `poly:"dog"`
//		    Cats  []Cat `poly:"cat"`
//	  	    Owner Owner `poly:"owner"`
//		}
//
//		var result Result
//		err := UnmarshallCustom(jsonData, &result, reflect.Type(AnimalTypeLocator{})
//
// In this example, the UnmarshallCustom function would unmarshall the JSON
// into the Result struct, populating the Dogs and Cats slices based on the polymorphic
// type names defined in the TypeLocator struct.
func UnmarshallCustom(rawJson []byte, target any, typeLocator reflect.Type) error {
	if len(rawJson) == 0 {
		return nil
	}

	targetFields, err := makeTargetFieldLookup(target)
	if err != nil {
		return err
	}

	subTypesSlice, err := unmarshallTypeMap(rawJson, typeLocator)
	if err != nil {
		return err
	}

	subJSONs, err := unmarshallSubArrays(rawJson)
	if err != nil {
		// We should never hit this because we've previously unmarshalled the type map above.
		return err
	}

	targetValue := reflect.ValueOf(target).Elem()
	for i := 0; i < subTypesSlice.Len(); i++ {
		// Figure out what type of object we need to make to satisfy the polymorphic
		// needs for *this* sub-object.
		tc, ok := subTypesSlice.Index(i).Interface().(TypeLocator)
		if !ok {
			// This should be impossible to get to as we've already checked.
			return fmt.Errorf("could not convert object to a TypeLocator")
		}
		t := tc.TypeName()
		if len(t) == 0 {
			// If nothing is returned, that's the signal that we are not interested in
			// this sub-object.
			continue
		}
		if fl, ok := targetFields[t]; ok {
			// We have a matching field we should unmarshall into.

			// Create an instance of that object and unmarshall the sub-JSON into
			// this object.
			newSub := reflect.New(fl.fieldType)
			newSubObj := newSub.Interface()
			err = json.Unmarshal(subJSONs[i], newSubObj)
			if err != nil {
				return err
			}

			// If that object implements the IndexSettable interface, let it know the
			// index from which it was read from.
			if indexable, ok := newSubObj.(IndexSettable); ok {
				indexable.SetIndex(i)
			}

			// If the actual target isn't a pointer, unwrap the Value into the object itself.
			if !fl.ptr {
				newSub = newSub.Elem()
			}

			// Finally figure out how to save it.
			if fl.kind == reflect.Slice {
				// A slice gets appended to.
				newSlice := reflect.Append(targetValue.Field(fl.index), newSub)
				targetValue.Field(fl.index).Set(newSlice)
			} else {
				// A value just gets set.
				targetValue.Field(fl.index).Set(newSub)
			}
		}
	}

	return nil
}

// makeTargetFieldLookup is a helper function that takes a target any type
// variable and returns a map of fieldLookup structs keyed by the polymorphic
// type names. The target variable should be a struct with fields optionally
// tagged with their respective polymorphic type names or using the field name as
// the default type name if no tag is provided. If the target variable is not a
// pointer, the function returns an error along with an empty map.
//
// This function is used internally by UnmarshallCustom to create a lookup
// table for target struct fields, allowing it to efficiently match and unmarshal
// JSON objects into the appropriate target fields based on their polymorphic type names.
//
// Example usage:
//
//		type Result struct {
//		    Dogs  []Dog `poly:"dog"`
//		    Cats  []Cat `poly:"cat"`
//	     	Owner Owner `poly:"owner"`
//		}
//
//		fields, err := makeTargetFieldLookup(&Result{})
//		// fields is a map containing fieldLookup structs for the "dog," "cat," and "owner" types.
//
// The returned map would have two entries, one for the "dog" type and one for the "cat"
// type. Each entry would contain a fieldLookup struct with information about the
// corresponding field in the target struct, such as the field index, field type,
// whether it is a pointer, and the kind of the field (e.g., slice or value).
func makeTargetFieldLookup(target any) (map[string]fieldLookup, error) {
	fields := map[string]fieldLookup{}
	targetTypePtr := reflect.TypeOf(target)
	if targetTypePtr.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("target must be a pointer")
	}
	targetType := targetTypePtr.Elem()
	for i := 0; i < targetType.NumField(); i++ {
		f := targetType.Field(i)

		fl := fieldLookup{
			index:     i,
			fieldType: f.Type,
			kind:      f.Type.Kind(),
		}

		if f.Type.Kind() == reflect.Slice {
			fl.fieldType = f.Type.Elem()
		} else {
			fl.fieldType = f.Type
		}
		if fl.fieldType.Kind() == reflect.Pointer {
			fl.ptr = true
			fl.fieldType = fl.fieldType.Elem()
		}

		var typeName string
		if tag, ok := f.Tag.Lookup("poly"); ok {
			typeName = tag
		} else {
			typeName = f.Name
		}
		fields[typeName] = fl
	}
	return fields, nil
}

// unmarshallTypeMap is a helper function that takes a raw JSON byte slice and a
// typeLocator of type reflect.Type. It unmarshalls the JSON into a slice of
// typeLocator instances, one for each object in the input JSON. The typeLocator
// should be a reflect.Type implementing the TypeLocator interface. If an error occurs
// during unmarshalling, it returns an error along with an empty reflect.Value.
//
// This function is used internally by UnmarshallCustom to determine the
// polymorphic type names for each object in the JSON.
func unmarshallTypeMap(rawJson []byte, typeLocator reflect.Type) (reflect.Value, error) {
	// Verify that the typeLocator is suitable.
	if !reflect.PointerTo(typeLocator).AssignableTo(typeLocatorType) {
		return reflect.Value{}, fmt.Errorf("typeLocator not assignable to a TypeLocator")
	}

	typeSliceType := reflect.SliceOf(reflect.PointerTo(typeLocator))
	slicePtr := reflect.New(typeSliceType)

	err := json.Unmarshal(rawJson, slicePtr.Interface())
	if err != nil {
		return reflect.Value{}, err
	}

	return slicePtr.Elem(), nil
}

// unmarshallSubArrays is a helper function that takes a raw JSON byte slice and
// returns a slice of json.RawMessage objects, where each json.RawMessage corresponds
// to an object in the input JSON array. If an error occurs during unmarshalling, it
// returns an error along with an empty slice.
//
// This function is used internally by UnmarshallCustom to extract the JSON
// objects for each sub-object, which will later be unmarshalled into the appropriate
// target fields based on their polymorphic type names.
func unmarshallSubArrays(rawJson []byte) ([]json.RawMessage, error) {
	var subJSONs []json.RawMessage
	err := json.Unmarshal(rawJson, &subJSONs)
	if err != nil {
		// We should never get here because the code flow would have already unmarshalled this
		// in a different way earlier.
		return nil, err
	}
	return subJSONs, nil
}
