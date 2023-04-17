package poly

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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

	a := &TypeInt{
		ValueC: 42,
		index:  1,
	}

	var aa any
	aa = a

	aaa := aa.(IndexGettable)
	fmt.Println(aaa.GetIndex())
}
