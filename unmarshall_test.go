package poly

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshallPoly(t *testing.T) {
	in := `
[
	{
		"type": "TypeString",
		"ValueA": "ValueString"
	},
	{
		"type": "TypeString",
		"ValueA": "ValueString2"
	},
	{
		"@type": "TypeFloat",
		"ValueB": 42.23
	},
	{
		"Type": "TypeInt",
		"ValueC": 105
	},
	{
		"@Type": "TypeIntP",
		"ValueC": 123
	}
]`
	var result SlicesABC

	err := UnmarshallPoly([]byte(in), &result)
	assert.NoError(t, err)

	assert.Len(t, result.TypeString, 2)
	assert.Equal(t, "ValueString", result.TypeString[0].ValueA)
	assert.Equal(t, "ValueString2", result.TypeString[1].ValueA)
	assert.Len(t, result.TypeBravo, 1)
	assert.Equal(t, float32(42.23), result.TypeBravo[0].ValueB)
	assert.Equal(t, 105, result.TypeInt.ValueC)
	assert.Equal(t, 3, result.TypeInt.index)
	assert.Equal(t, 123, result.TypeIntP.ValueC)
	assert.Equal(t, 4, result.TypeIntP.index)
}
