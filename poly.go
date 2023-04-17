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

// Indexable is an interface that you should implement if you need to know the
// index into the array of JSON sub-objects. This should be implemented by the
// types that are referred to by the TypeLocator.
type Indexable interface {
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

func UnmarshallPoly(rawJson []byte, target any) error {
	return UnmarshallPolyCustomType(rawJson, target, DefaultLocator)
}

func UnmarshallPolyCustomType(rawJson []byte, target any, typeLocator reflect.Type) error {
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

			// If that object implements the Indexable interface, let it know the
			// index from which it was read from.
			if indexable, ok := newSubObj.(Indexable); ok {
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

func unmarshallSubArrays(rawJson []byte) ([]json.RawMessage, error) {
	var subJSONs []json.RawMessage
	err := json.Unmarshal(rawJson, &subJSONs)
	if err != nil {
		return nil, err
	}
	return subJSONs, nil
}
