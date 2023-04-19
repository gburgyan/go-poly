package poly

import (
	"encoding/json"
	"fmt"
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
	return Unmarshall(rawJson, r)
}

func (r Residence) MarshalJSON() ([]byte, error) {
	return Marshall(r)
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
]
`

	// First do it manually using the library Unmarshall function
	r := Residence{}
	err := Unmarshall([]byte(in), &r)
	assert.NoError(t, err)

	// Now, see how this works with the custom implementation of the json.UnmarshallJSON function
	r2 := Residence{}
	err = json.Unmarshal([]byte(in), &r2)
	assert.NoError(t, err)

	m2bytes, err := json.Marshal(r2)
	assert.NoError(t, err)
	fmt.Println(string(m2bytes))
}
