package tests

import "github.com/onrik/gaws/tests/nested"

type SliceAlias []string

type StructForAlias struct {
	Name string `json:"name"`
}

type Alias StructForAlias

type NestedAlias nested.NestedStruct
