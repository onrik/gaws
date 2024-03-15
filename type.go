package main

type ParsedTypeKind int

const (
	baseType ParsedTypeKind = iota + 1
	timeType
	mapType
	arrayType
	structType
)

// ParsedType represents parsed type of value
type ParsedType struct {
	Kind   ParsedTypeKind
	Name   string
	File   File
	Nested *ParsedType
}
