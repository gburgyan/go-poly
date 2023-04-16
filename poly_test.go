package poly

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TypeString struct {
	ValueA string
}

type TypeFloat struct {
	ValueB float32
}

type TypeInt struct {
	ValueC int
	Index  int
}

func (t *TypeInt) SetIndex(i int) {
	t.Index = i
}

type SlicesAB struct {
	TypeString []TypeString
	TypeBravo  []TypeFloat `poly:"TypeFloat"`
	TypeInt    TypeInt
	TypeIntP   *TypeInt
}

func TestUnmarshallPoly(t *testing.T) {
	in := `
[
	{
		"type": "TypeString",
		"ValueA": "ValueString"
	},
	{
		"@type": "TypeFloat",
		"ValueB": 42.23
	},
	{
		"@type": "TypeInt",
		"ValueC": 105
	},
	{
		"@type": "TypeIntP",
		"ValueC": 123
	}
]`
	var result SlicesAB

	err := UnmarshallPoly([]byte(in), &result)
	assert.NoError(t, err)

	assert.Len(t, result.TypeString, 1)
	assert.Len(t, result.TypeBravo, 1)
}
