package poly

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Residence struct {
	Location Location      `poly:"location"`
	People   []Person      `poly:"person"`
	Pets     []Pet         `poly:"pet"`
	Water    *WaterService `poly:"water"`
}

func (r *Residence) UnmarshalJSON(rawJson []byte) error {
	return Unmarshal(rawJson, r)
}

func (r Residence) MarshalJSON() ([]byte, error) {
	return Marshal(r)
}

type Location struct {
	Address string `json:"address"`
}

type Person struct {
	Name       string `json:"name,omitempty"`
	Occupation string `json:"occupation,omitempty"`
	Age        int    `json:"age,omitempty"`
}

type Pet struct {
	Name    string `json:"name,omitempty"`
	Species string `json:"species,omitempty"`
}

type WaterService struct {
	Provider string `json:"provider,omitempty"`
}

func TestExampleUnmarshall(t *testing.T) {
	in := `
[
  {
    "type": "location",
    "address": "123 Main"
  },
  {
    "type": "person",
    "name": "John",
    "occupation": "Teacher",
    "age": 35
  },
  {
    "type": "person",
    "name": "Mary",
    "occupation": "Programmer",
    "age": 33
  },
  {
    "type": "pet",
    "name": "Rover",
    "species": "dog"
  },
  {
    "type": "pet",
    "name": "Fluffy",
    "species": "cat"
  },
  {
    "type": "water",
    "provider": "Public City Water"
  }
]`

	// First do it manually using the library Unmarshal function
	r := Residence{}
	err := Unmarshal([]byte(in), &r)
	assert.NoError(t, err)
	assert.Equal(t, "123 Main", r.Location.Address)
	assert.Len(t, r.People, 2)
	assert.Equal(t, "John", r.People[0].Name)
	assert.Equal(t, "Mary", r.People[1].Name)
	assert.Len(t, r.Pets, 2)
	assert.Equal(t, "Rover", r.Pets[0].Name)
	assert.Equal(t, "Fluffy", r.Pets[1].Name)
	assert.Equal(t, "Public City Water", r.Water.Provider)

	// Now, see how this works with the custom implementation of the json.UnmarshallJSON function
	r2 := Residence{}
	err = json.Unmarshal([]byte(in), &r2)
	assert.NoError(t, err)
	assert.Equal(t, "123 Main", r2.Location.Address)
	assert.Len(t, r2.People, 2)
	assert.Equal(t, "John", r2.People[0].Name)
	assert.Equal(t, "Mary", r2.People[1].Name)
	assert.Len(t, r2.Pets, 2)
	assert.Equal(t, "Rover", r2.Pets[0].Name)
	assert.Equal(t, "Fluffy", r2.Pets[1].Name)
	assert.Equal(t, "Public City Water", r2.Water.Provider)

	// Now use the json.MarshalIndent to marshall the r2 object back to JSON. This should
	// invoke the MarshalJSON receiver function above to apply the custom polymorphic
	// marshalling.
	marshalledBytes, err := json.MarshalIndent(r2, "", "  ")
	assert.NoError(t, err)

	// Note that this does *not* include the type fields in the result JSON. If you need
	// the types emitted, then you need to add fields to the structs and set them appropriately.
	expected := `[
  {
    "address": "123 Main"
  },
  {
    "name": "John",
    "occupation": "Teacher",
    "age": 35
  },
  {
    "name": "Mary",
    "occupation": "Programmer",
    "age": 33
  },
  {
    "name": "Rover",
    "species": "dog"
  },
  {
    "name": "Fluffy",
    "species": "cat"
  },
  {
    "provider": "Public City Water"
  }
]`
	assert.Equal(t, expected, string(marshalledBytes))

}
