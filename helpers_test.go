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

func TestGetParamsFromTag(t *testing.T) {
	tags, err := getParamsFromTag(`openapiDesc:"foo" openapiExample:"22" openapiEnum:"1,2,3" openapi:"required" openapiExt:"x-test=test"`)
	require.Nil(t, err)
	require.Equal(t, 1, len(tags.Openapi))
	require.Equal(t, "foo", tags.Description)
	require.Equal(t, "22", tags.Example)
	require.Equal(t, "1,2,3", tags.Enum)
	require.Equal(t, "", tags.Openapi["required"])
	require.Equal(t, "test", tags.Extensions["x-test"])
}

func TestParseParams(t *testing.T) {
	params, err := parseParams("required, type=string, example=1")
	require.Nil(t, err)
	require.Equal(t, 3, len(params))
	require.Equal(t, "string", params["type"])
	require.Equal(t, "1", params["example"])
	require.Equal(t, "", params["required"])

	// Test json example
	params, err = parseParams(`required, type=string, example={'foo': 'bar'}`)
	require.Nil(t, err)
	require.Equal(t, 3, len(params))
	require.Equal(t, "string", params["type"])
	require.Equal(t, `{"foo": "bar"}`, params["example"])
	require.Equal(t, "", params["required"])

	params, err = parseParams(`example={'foo': 'bar', 'id': 1}`)
	require.Nil(t, err)
	require.Equal(t, `{"foo": "bar", "id": 1}`, params["example"])
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
	require.Equal(t, "test.bar", addPkg("foo", "test.bar"))
}
