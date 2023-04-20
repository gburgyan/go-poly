# go-poly

[![Build Status](https://github.com/gburgyan/go-poly/actions/workflows/go.yml/badge.svg)](https://github.com/gburgyan/go-poly/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/gburgyan/go-poly)](https://goreportcard.com/report/github.com/gburgyan/go-poly)
[![GoDoc](https://pkg.go.dev/badge/github.com/gburgyan/go-poly)](https://pkg.go.dev/github.com/gburgyan/go-poly)
[![License](https://img.shields.io/github/license/gburgyan/go-poly)](LICENSE)

`go-poly` is a GoLang library that provides functionality for marshalling and unmarshalling polymorphic JSON. It is designed to simplify the handling of JSON objects with varying structures in Go applications.

## Features

- Easy marshalling and unmarshalling of polymorphic JSON
- Support for custom types
- Minimal dependencies and efficient performance
- Flexible usage patterns
- Does not lose any context from the original, even indexes are preserved

## Installation

To install `go-poly`, use the following command:

```sh
go get -u github.com/gburgyan/go-poly
```

## Usage

## Documentation

Polymorphic JSON exists due to the dynamic nature and flexibility of JSON data structures, allowing objects to have varying structures within a single collection. This can be useful when dealing with diverse data sources, evolving APIs, or when trying to maintain backward compatibility. Go isn't good at dealing with these JSON structures because it doesn't have a native handling of polymorphism.

This library aims to solve that by allowing the marshalling and unmarshalling of these JSONs easily without writing custom functions to deal with this.

Take this JSON for example that might define a residence:

```json
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
```

Using `go-poly` you can marshall and unmarshall this:

```go
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
```



Documentation
For more detailed usage instructions and examples, please refer to the [GoDoc documentation](https://pkg.go.dev/github.com/gburgyan/go-poly)

## License

`go-poly` is licensed under the [MIT License](LICENSE).