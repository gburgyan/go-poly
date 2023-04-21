package poly

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
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

	err := Unmarshal([]byte(in), &result)
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

func TestUnmarshal_BadLocator(t *testing.T) {
	in := `
[
	{
		"type": "TypeString",
		"ValueA": "ValueString"
	}
]`
	var result SlicesABC

	// string doesn't implement the TypeLocator interface.
	err := UnmarshalCustom([]byte(in), &result, reflect.TypeOf(""))
	assert.Error(t, err)
}

func TestUnmarshal_JSONError(t *testing.T) {
	in := `
[
	{
		"type": "TypeString",
		"ValueA": 42
	}
]`
	var result SlicesABC

	// string doesn't implement the TypeLocator interface.
	err := Unmarshal([]byte(in), &result)
	assert.Error(t, err)
}

func TestUnmarshal_InvalidJSON(t *testing.T) {
	var result SlicesABC
	err := Unmarshal([]byte(`not valid JSON`), &result)

	assert.Error(t, err)
}

func TestUnmarshal_NoType(t *testing.T) {
	in := `
[
	{
		"ValueA": "ValueString"
	}
]`
	var result SlicesABC

	// string doesn't implement the TypeLocator interface.
	err := Unmarshal([]byte(in), &result)
	assert.NoError(t, err)
	assert.Len(t, result.TypeString, 0)
	assert.Len(t, result.TypeBravo, 0)
	assert.Nil(t, result.TypeIntP)
}

func TestUnmarshal_NilSON(t *testing.T) {
	var result SlicesABC
	err := Unmarshal(nil, &result)

	assert.NoError(t, err)
	assert.Len(t, result.TypeString, 0)
	assert.Len(t, result.TypeBravo, 0)
	assert.Nil(t, result.TypeIntP)
}

func TestUnmarshal_EmptyJSON(t *testing.T) {
	var result SlicesABC
	err := Unmarshal([]byte(`[]`), &result)

	assert.NoError(t, err)
	assert.Empty(t, result.TypeString)
	assert.Empty(t, result.TypeBravo)
	assert.Nil(t, result.TypeIntP)
}

func TestUnmarshal_NonPointer(t *testing.T) {
	var result SlicesABC
	err := Unmarshal([]byte(`[]`), result)

	assert.Error(t, err)
}
