package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseStructs(t *testing.T) {
	p, err := parseStructs("", "./tests/", true)

	require.Nil(t, err)
	require.Equal(t, 58, len(p.structs))

	s := structByName(p.structs, "User")
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

	s = structByName(p.structs, "Group")
	require.NotNil(t, s)
	require.Equal(t, 3, len(s.Fields))
	require.Equal(t, "Data", s.Fields[1].Name)
	require.Equal(t, "json2.RawMessage", s.Fields[1].Type)
	require.Equal(t, "Admin", s.Fields[2].Name)
	require.Equal(t, "User", s.Fields[2].Type)

	s = structByName(p.structs, "Struct2")
	require.NotNil(t, s)
	require.Equal(t, 2, len(s.Fields))
	require.Equal(t, "Data", s.Fields[1].Name)
	require.Equal(t, "json3.RawMessage", s.Fields[1].Type)

	require.NotNil(t, structByName(p.structs, "json2.RawMessage"))
	require.NotNil(t, structByName(p.structs, "json3.RawMessage"))
}
