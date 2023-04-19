package poly

import (
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

	bytes, err := Marshall(in)
	assert.NoError(t, err)
	assert.Equal(t, `[{"ValueC":105},{"ValueC":23},{"ValueA":"A"},{"ValueA":"B"},{"ValueB":42},{"ValueB":43}]`, string(bytes))

	bytes, err = Marshall(&in) // Try with pointer
	assert.NoError(t, err)
	assert.Equal(t, `[{"ValueC":105},{"ValueC":23},{"ValueA":"A"},{"ValueA":"B"},{"ValueB":42},{"ValueB":43}]`, string(bytes))
}
