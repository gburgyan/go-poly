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

Polymorphic JSON exists due to the dynamic nature and flexibility of JSON data structures, allowing objects to have varying structures within a single collection. This can be useful when dealing with diverse data sources, evolving APIs, or when trying to maintain backward compatibility. Go isn't good at dealing with these JSON structures because it doesn't have a native handling of polymorphism.

This library aims to solve that by allowing the marshalling and unmarshalling of these JSONs easily without writing custom functions to deal with this.

### Unmarshalling

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

This library knows how to deal with slices of objects, in which case the newly unmarshalled object is appended to the slice. If it's a struct type, or a pointer to a struct type, then it's simply assigned. If there are multiple instances that are unmarshalled of a scalar type, the last one wins as it will simply overwrite the earlier ones.

#### Type Lookups

The default implementation uses the `GenericTypeLocator` which looks for common type discriminators:

* type
* Type
* @type
* @Type

If you need to do something special, pass in a `Type` that implements `TypeLocator` to `UnmarshallCustom`. What will happen is the JSON will first be unmarshalled into a slice of your type, then each instance of that will be used to determine the type of object to actually unmarshall.

#### Finding the correct target field

The returned type name is used to figure out what field in the target object will get filled. If there is no `poly` tag on a field, the name of the field is used verbatim. If the field has a `poly tag, then that is used to find the correct field.

#### Indexing

While not typical, sometimes the ordering of the elements in the JSON array is important. In that case, implement the `IndexSettable` interface by the types that are being deserialized.

The `SetIndex(index int)` will be called after the object is unmarshalled with the zero-based index of the JSON array from which it was unmarshalled from.

### Marshalling

You can call the `poly.Marshall` function to marshall everything back to JSON. This will flatten everything back to a slice that can be marshalled back into a JSON array.

#### Indexing

Similar to the unmarshalling, sometimes the order in which the elements of the JSON array appear is important. If this is important to you, simply implement the `IndexGettable` interface with your object. The `GetIndex()` will be called to determine the relative index.

If there are multiple elements that return the same index, they will all be grouped together. The order in which they were encountered will be preserved within the same returned index.

If there are some objects that don't implement the `IndexGettable` interface, they will simply be sorted to the end.

#### Limitation

There is no magic to emit any sort of type field in the marshalled JSON. If you need to have a type discriminator in the JSON, you must have that field on the objects that are being marshalled with the value set appropriately.

```go
type JsonObject struct {
    Type string `json:"type"`
    Value string
}
```

Without the `Type` field, or something like it, there will not be a type marshalled in the JSON. 

## License

`go-poly` is licensed under the [MIT License](LICENSE).