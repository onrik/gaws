package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePath(t *testing.T) {
	parser := NewParser(nil, nil)
	method, path, err := parser.parsePath("@openapi GET /api/v1/test ")
	require.Nil(t, err)
	require.Equal(t, "get", method)
	require.Equal(t, "/api/v1/test", path)
}

func TestParseParam(t *testing.T) {
	parser := NewParser(nil, nil)
	param, err := parser.parseParam("@openapiParam id in=path, type=int, example=11")

	require.Nil(t, err)
	require.Equal(t, "id", param.Name)
	require.Equal(t, "path", param.In)
	require.NotNil(t, param.Schema)
	require.Equal(t, "integer", param.Schema.Type)
	require.Equal(t, "11", param.Schema.Example)
}

func TestParseRequest(t *testing.T) {
	parser := NewParser(nil, []Struct{Struct{
		Name: "User",
		Fields: []StructField{
			StructField{
				Name:       "Name",
				Type:       "string",
				Tag:        `json:"name"`,
				IsExported: true,
				IsPointer:  false,
			},
		},
	}})

	// Test invalid schema
	body, err := parser.parseRequest(`@openapiRequest application/json {"foo", "bar"}`)
	require.NotNil(t, err)
	require.Equal(t, "Invalid JSON schema", err.Error())

	// Test unsupported content type
	body, err = parser.parseRequest(`@openapiRequest text/plain {}`)
	require.NotNil(t, err)
	require.Equal(t, "Unsupported Content-Type", err.Error())

	// Test json example
	body, err = parser.parseRequest(`@openapiRequest application/json {"foo": "bar"}`)
	require.Nil(t, err)

	content, e := body.Content["application/json"]
	require.True(t, e)
	require.Equal(t, `{"foo": "bar"}`, content.Example)

	// Test struct
	body, err = parser.parseRequest(`@openapiRequest application/json User`)
	require.Nil(t, err)

	content, e = body.Content["application/json"]
	require.True(t, e)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "string", content.Schema.Properties["name"].Type)

	// Test json schema
	body, err = parser.parseRequest(`@openapiRequest application/json {"user": User, "id": int}`)
	require.Nil(t, err)

	content, e = body.Content["application/json"]
	require.True(t, e)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "integer", content.Schema.Properties["id"].Type)
	require.Equal(t, "object", content.Schema.Properties["user"].Type)
	require.Equal(t, "string", content.Schema.Properties["user"].Properties["name"].Type)
}

func TestParseResponse(t *testing.T) {
	parser := NewParser(nil, []Struct{Struct{
		Name: "User",
		Fields: []StructField{
			StructField{
				Name:       "Name",
				Type:       "string",
				Tag:        `json:"name"`,
				IsExported: true,
				IsPointer:  false,
			},
		},
	}})

	// Test invalid schema
	status, contentType, content, err := parser.parseResponse(`@openapiResponse 200 application/json {"foo", "bar"}`)
	require.NotNil(t, err)
	require.Equal(t, "Invalid JSON schema", err.Error())

	// Test unsupported content type
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 text/xml {"foo": "bar"}`)
	require.NotNil(t, err)
	require.Equal(t, "Unsupported Content-Type", err.Error())

	// Test invalid status code
	status, contentType, content, err = parser.parseResponse(`@openapiResponse ddd application/json {"foo": "bar"}`)
	require.NotNil(t, err)
	require.Equal(t, "Invalid HTTP status code", err.Error())

	// Test json example
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/json {"foo": "bar"}`)
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/json", contentType)
	require.Equal(t, `{"foo": "bar"}`, content.Example)

	// Test struct
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/json User`)
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/json", contentType)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "string", content.Schema.Properties["name"].Type)

	// Test json schema
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/json {"user": User, "id": int}`)
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/json", contentType)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "integer", content.Schema.Properties["id"].Type)
	require.Equal(t, "object", content.Schema.Properties["user"].Type)
	require.Equal(t, "string", content.Schema.Properties["user"].Properties["name"].Type)

	// Test application/octet-stream
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/octet-stream`)
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/octet-stream", contentType)
	require.Equal(t, "string", content.Schema.Type)
	require.Equal(t, "binary", content.Schema.Format)
}

func TestParseStruct(t *testing.T) {
	parser := NewParser(nil, []Struct{Struct{
		Name: "User",
		Fields: []StructField{
			StructField{
				Name:       "Name",
				Type:       "string",
				Tag:        `json:"name"`,
				IsExported: true,
				IsPointer:  false,
			},
		},
	}})

	s, err := parser.parseStruct("Test")
	require.NotNil(t, err)
	require.Equal(t, "Unknown type: Test", err.Error())

	s, err = parser.parseStruct("User")
	require.Nil(t, err)
	require.Equal(t, "object", s.Type)

	s, err = parser.parseStruct("[]User")
	require.Nil(t, err)
	require.Equal(t, "array", s.Type)
}

func TestTypeToProperty(t *testing.T) {
	parser := NewParser(nil, nil)

	p, err := parser.typeToProperty("", "int")
	require.Nil(t, err)
	require.Equal(t, "integer", p.Type)
	require.Equal(t, "", p.Format)

	p, err = parser.typeToProperty("", "*int")
	require.Nil(t, err)
	require.Equal(t, "integer", p.Type)
	require.Equal(t, "", p.Format)

	p, err = parser.typeToProperty("", "string")
	require.Nil(t, err)
	require.Equal(t, "string", p.Type)
	require.Equal(t, "", p.Format)

	p, err = parser.typeToProperty("", "time.Time")
	require.Nil(t, err)
	require.Equal(t, "string", p.Type)
	require.Equal(t, "date-time", p.Format)

	p, err = parser.typeToProperty("", "*time.Time")
	require.Nil(t, err)
	require.Equal(t, "string", p.Type)
	require.Equal(t, "date-time", p.Format)

	p, err = parser.typeToProperty("", "[]string")
	require.Nil(t, err)
	require.Equal(t, "array", p.Type)
	require.NotNil(t, p.Items)
	require.Equal(t, "string", p.Items.Type)

	p, err = parser.typeToProperty("", "User")
	require.NotNil(t, err)
	require.Equal(t, "Unknown type: User", err.Error())

	parser.structs = append(parser.structs, Struct{
		Name: "User",
		Fields: []StructField{
			StructField{
				Name:       "Name",
				Type:       "string",
				Tag:        `json:"name"`,
				IsExported: true,
				IsPointer:  false,
			},
		},
	})

	p, err = parser.typeToProperty("", "User")
	require.Nil(t, err)
	require.Equal(t, "object", p.Type)
	require.Equal(t, "string", p.Properties["name"].Type)
}

func TestParseTags(t *testing.T) {
	tags := parseTags("@openapiTags foo,  bar")
	require.Equal(t, []string{"foo", "bar"}, tags)
}
