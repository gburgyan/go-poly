# go-poly

[![Build Status](https://github.com/gburgyan/go-poly/actions/workflows/go.yml/badge.svg)](https://github.com/gburgyan/go-poly/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/gburgyan/go-poly)](https://goreportcard.com/report/github.com/gburgyan/go-poly)
[![GoDoc](https://pkg.go.dev/badge/github.com/gburgyan/go-poly)](https://pkg.go.dev/github.com/gburgyan/go-poly)
[![License](https://img.shields.io/github/license/gburgyan/go-poly)](LICENSE)

`go-poly` is a GoLang library that provides functionality for marshalling and unmarshalling polymorphic JSON. It is designed to simplify the handling of JSON objects with varying structures in Go applications.

## Features

* Simplified marshalling and unmarshalling of polymorphic JSON
* Support for custom types
* Minimal dependencies and efficient performance
* Flexible usage patterns
* Preservation of original context, including indexes

## Installation

To install `go-poly`, use the following command:

```sh
go get -u github.com/gburgyan/go-poly
```

## Usage

Polymorphic JSON allows objects to have varying structures within a single collection due to the dynamic nature and flexibility of JSON data structures. This is useful when dealing with diverse data sources, evolving APIs, or maintaining backward compatibility. However, Go lacks native support for handling polymorphic JSON structures.

`go-poly` aims to address this limitation by enabling easy marshalling and unmarshalling of polymorphic JSONs without the need for custom functions.

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

Using `go-poly` you can marshall and unmarshall like this:

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

You can then unmarshall the JSON into your object by:

```go
var residence Residence
err := json.Unmarshall(input, &residence)
```

Since you've implemented the `json.Unmarshaler` interface, your function `UnmarshalJSON` will get called to handle the unmarshalling. This will happen even this is buried within a larger JSON document.

Alternately, you can call the polymorphic unmarshalling directly:

```go
var residence Residence
err := poly.Unmarshall(input, &residence)
```

This library handles slices of objects by appending newly unmarshalled objects to the slice. For struct types or pointers to struct types, they are simply assigned. If multiple instances of a scalar type are unmarshalled, the last instance will overwrite earlier ones.

#### Type Lookups

The default implementation uses the `GenericTypeLocator` which looks for common type discriminators:

* type
* Type
* @type
* @Type

For custom implementations, provide a `Type` that implements the `TypeLocator` interface and pass it to `UnmarshalCustom`. During the unmarshalling process, the JSON will first be converted into a slice of your custom type. Subsequently, each instance in the slice will be used to determine the actual object type for unmarshalling. This approach offers flexibility, allowing your implementation to perform any necessary actions to identify the correct type. For example, if you need to examine multiple JSON fields to determine the concrete type, your custom implementation can handle that.

#### Finding the correct target field

The returned type name is used to figure out what field in the target object will get filled. If there is no `poly` tag on a field, the name of the field is used verbatim. If the field has a `poly` tag, then that is used to find the correct field.

#### Indexing

In cases where the order of elements in the JSON array is important, implement the `IndexSettable` interface for the types being deserialized.

After unmarshalling, the `SetIndex(index int)` function will be called with the zero-based index of the JSON array from which the object was unmarshalled.

### Marshalling

As with unmarshalling, implementing the `json.Marshaler` interface will trigger the `MarshalJSON` function during the marshalling process. When calling `json.Marshal`, your function will handle marshalling, and the polymorphic JSON will be emitted.

You can also manually call `poly.Marshal` without implementing any interface, but this method will not participate in the standard json.Marshal behavior.

#### Indexing

Similar to unmarshalling, the order of elements in the JSON array may be important during marshalling. To maintain the desired order, implement the `IndexGettable` interface for your object. The `GetIndex()` function will be called to determine the relative index.

If multiple elements return the same index, they will be grouped together, and the order in which they were encountered will be preserved within the same returned index.

Objects that do not implement the IndexGettable interface will be sorted to the end of the array.

#### Limitation

The library does not automatically emit a type field in the marshalled JSON. If you require a type discriminator in the JSON, include an appropriate field with the correct value in the objects being marshalled, as shown below:

```go
type JsonObject struct {
    Type string `json:"type"`
    Value string
}
```

Without the `Type` field, or a similar field, the type will not be marshalled in the JSON.

## License

`go-poly` is licensed under the [MIT License](LICENSE).