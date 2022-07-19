package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetStr(t *testing.T) {
	require.Equal(t, "2", getStr([]string{"1", "2"}, 1))
	require.Equal(t, "", getStr([]string{"1", "2"}, 2))
}

func TestAtoi(t *testing.T) {
	require.Equal(t, 0, atoi("t"))
	require.Equal(t, 201, atoi("201"))
}

func TestStrIn(t *testing.T) {
	require.True(t, strIn("1", []string{"1", "2", "3"}))
	require.False(t, strIn("4", []string{"1", "2", "3"}))
}

func TestParseParams(t *testing.T) {

}

func TestParseJSONSchema(t *testing.T) {
	schema, err := parseJSONSchema(`{"foo", bar}`)
	require.NotNil(t, err)

	schema, err = parseJSONSchema(`{
	"foo": bar,
}`)
	require.Nil(t, err)
	require.Equal(t, map[string]string{"foo": "bar"}, schema)
}

func TestGetPkg(t *testing.T) {
	require.Equal(t, "", getPkg("test"))
	require.Equal(t, "test", getPkg("test.Test"))
	require.Equal(t, "json", getPkg("json.F"))
}

func TestAddPkg(t *testing.T) {
	require.Equal(t, "test", addPkg("", "test"))
	require.Equal(t, "foo.bar", addPkg("foo", "bar"))
	require.Equal(t, "test.bar", addPkg("test", "test.bar"))
}
