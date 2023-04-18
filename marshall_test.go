package poly

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestMarshallPoly(t *testing.T) {
	in := SlicesABC{
		TypeString: []TypeString{
			{
				ValueA: "A",
			},
			{
				ValueA: "B",
			},
		},
		TypeBravo: []TypeFloat{
			{
				ValueB: 42,
			},
			{
				ValueB: 43,
			},
		},
		TypeInt: TypeInt{
			ValueC: 23,
			index:  2,
		},
		TypeIntP: &TypeInt{
			ValueC: 105,
			index:  1,
		},
	}

	bytes, err := MarshallPoly(in)
	assert.NoError(t, err)
	s := string(bytes)
	fmt.Println(s)
}

func TestReflectNonPointerInterface(t *testing.T) {
	a := TypeInt{
		ValueC: 42,
		index:  1,
	}

	at := reflect.TypeOf(a)
	av := reflect.ValueOf(a)

	pav := reflect.New(at)
	pav.Elem().Set(av)

	igv := pav.Convert(indexGettableType)
	i := igv.Interface().(IndexGettable)

	fmt.Println(i.GetIndex())
}
