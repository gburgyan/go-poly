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
	var result SlicesABC

	err := UnmarshallPoly([]byte(in), &result)
	assert.NoError(t, err)

	assert.Len(t, result.TypeString, 1)
	assert.Len(t, result.TypeBravo, 1)
}
