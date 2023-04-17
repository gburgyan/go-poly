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

		//if fieldType.Kind() == reflect.Struct {
		//	fieldType = reflect.PointerTo(fieldType)
		//	ptrValue := reflect.New(fieldType)
		//	ptrValue.Set(fieldValue.Addr())
		//	fieldValue = ptrValue
		//}

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
