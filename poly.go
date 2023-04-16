package poly

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type TypeLocator interface {
	TypeName() string
}

var TypeLocatorType = reflect.TypeOf([]TypeLocator{}).Elem()

type GenericTypeLocator struct {
	Type       string `json:"type,omitempty"`
	TypeAt     string `json:"@type,omitempty"`
	TypeCaps   string `json:"Type,omitempty"`
	TypeAtCaps string `json:"@Type,omitempty"`
}

var DefaultLocator = reflect.TypeOf(GenericTypeLocator{})

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

type Indexable interface {
	SetIndex(index int)
}

type fieldLookup struct {
	index     int
	fieldType reflect.Type
	rootType  reflect.Type
	kind      reflect.Kind
}

func UnmarshallPoly(rawJson []byte, target any) error {
	return UnmarshallPolyCustomType(rawJson, target, DefaultLocator)
}

func UnmarshallPolyCustomType(rawJson []byte, target any, typeLocator reflect.Type) error {
	// Verify that the typeLocator is suitable.
	fmt.Println(TypeLocatorType)
	if !reflect.PointerTo(typeLocator).AssignableTo(TypeLocatorType) {
		return fmt.Errorf("typeLocator not assignable to a TypeLocator")
	}

	fields := map[string]fieldLookup{}
	targetTypePtr := reflect.TypeOf(target)
	if targetTypePtr.Kind() != reflect.Pointer {
		return fmt.Errorf("target must be a pointer")
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
		var typeName string
		if tag, ok := f.Tag.Lookup("poly"); ok {
			typeName = tag
		} else {
			typeName = f.Name
		}
		fields[typeName] = fl
	}

	typeSliceType := reflect.SliceOf(reflect.PointerTo(typeLocator))
	fmt.Println(typeSliceType)
	slicePtr := reflect.New(typeSliceType)

	err := json.Unmarshal(rawJson, slicePtr.Interface())
	if err != nil {
		return err
	}

	subTypesSlice := slicePtr.Elem()

	var subJSONs []json.RawMessage
	err = json.Unmarshal(rawJson, &subJSONs)
	if err != nil {
		return err
	}

	targetValue := reflect.ValueOf(target).Elem()
	for i := 0; i < subTypesSlice.Len(); i++ {
		tc := subTypesSlice.Index(i).Interface().(TypeLocator)
		t := tc.TypeName()
		if len(t) == 0 {
			continue
		}
		if fl, ok := fields[t]; ok {
			newSub := reflect.New(fl.fieldType)
			newSubObj := newSub.Interface()
			err = json.Unmarshal(subJSONs[i], newSubObj)
			if err != nil {
				return err
			}
			if indexable, ok := newSubObj.(Indexable); ok {
				indexable.SetIndex(i)
			}
			if fl.kind == reflect.Slice {
				newSlice := reflect.Append(targetValue.Field(fl.index), newSub.Elem())
				targetValue.Field(fl.index).Set(newSlice)
			} else {
				targetValue.Field(fl.index).Set(newSub.Elem())
			}
		}
	}

	return nil
}
