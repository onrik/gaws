package main

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseStructs(t *testing.T) {
	p := newStructsParser()
	st, err := p.parse(Package{FSPath: "./tests/", ImportPath: ""})
	require.NoError(t, err)
	require.Nil(t, err)
	require.Equal(t, 12, len(st))

	s, ok := st["User"]
	require.True(t, ok)
	require.NotNil(t, s)
	require.Equal(t, 4, len(s.Fields))
	require.Equal(t, "ID", s.Fields[0].Name)
	require.Equal(t, "int", s.Fields[0].Type)
	require.Equal(t, "Name", s.Fields[1].Name)
	require.Equal(t, "string", s.Fields[1].Type)
	require.Equal(t, "CreatedAt", s.Fields[2].Name)
	require.Equal(t, "*time.Time", s.Fields[2].Type)
	require.Equal(t, true, s.Fields[2].IsPointer)
	require.Equal(t, "Groups", s.Fields[3].Name)
	require.Equal(t, "[]Group", s.Fields[3].Type)
	require.Equal(t, "User", s.Name)
	require.Equal(t, "", s.Pkg)

	s, ok = st["Group"]
	require.True(t, ok)
	require.NotNil(t, s)
	require.Equal(t, 3, len(s.Fields))
	require.Equal(t, "Data", s.Fields[1].Name)
	require.Equal(t, "json2.RawMessage", s.Fields[1].Type)
	require.Equal(t, "Admin", s.Fields[2].Name)
	require.Equal(t, "User", s.Fields[2].Type)

	s, ok = st["Struct2"]
	require.True(t, ok)
	require.NotNil(t, s)
	require.Equal(t, 2, len(s.Fields))
	require.Equal(t, "Data", s.Fields[1].Name)
	require.Equal(t, "json3.RawMessage", s.Fields[1].Type)

	// alias structs
	s, ok = st["SliceAlias"]
	require.True(t, ok)
	require.NotNil(t, s)
	require.Equal(t, "", s.Pkg)
	require.Equal(t, []StructField(nil), s.Fields)
	require.Equal(t, "[]string", s.Origin)

	// parse types from std lib
	pkgs, err := parser.ParseDir(token.NewFileSet(), "./tests/", nil, parser.ParseComments)
	require.NoError(t, err)

	f := NewFile(pkgs["tests"].Files["tests/structs.go"], "./tests/", "")
	f, err = f.ParseImport("json2", "RawMessage")
	require.NoError(t, err)

	st, err = p.parse(f.Pkg)
	require.NoError(t, err)
	require.NotNil(t, st["RawMessage"])

	// parsed package should be cached
	require.NotNil(t, p.structs["encoding/json"]["RawMessage"])

	//s = structByName(p.structs, "NestedAlias", pkgs["tests"].Files["tests/alias_structs.go"])
	//require.NotNil(t, s)
	//require.Equal(t, "", s.Pkg)
	//require.Equal(t, []StructField(nil), s.Fields)
	//require.Equal(t, "nested.NestedStruct", s.Origin)
	//
	//// nested struct
	//s = structByName(p.structs, "nested.NestedStruct", pkgs["tests"].Files["tests/structs4.go"])
	//require.NotNil(t, s)
	//require.Equal(t, "github.com/onrik/gaws/tests/nested", s.Pkg)

	// TODO: uncomment this test when deep parsing will be enabled
	// even more nested struct with duplicate module name
	//pkgs, err = parser.ParseDir(token.NewFileSet(), "./tests/nested", nil, parser.ParseComments)
	//if err != nil {
	//	require.Nil(t, err)
	//}
	//
	//s = structByName(p.structs, "nested.NestedStruct", pkgs["nested"].Files["tests/nested/nested.go"])
	//require.NotNil(t, s)
	//require.Equal(t, "github.com/onrik/gaws/tests/nested/nested", s.Pkg)
}
