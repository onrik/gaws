package main

import (
	goParser "go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func getFile(t *testing.T, pkg, fsPath, importPath string) File {
	pkgs, err := goParser.ParseDir(token.NewFileSet(), filepath.Dir(fsPath), nil, goParser.ParseComments)
	require.NoError(t, err)

	return NewFile(pkgs[pkg].Files[fsPath], filepath.Dir(fsPath), importPath)
}

func TestParsePath(t *testing.T) {
	parser := NewParser(nil, newStructsParser())
	method, path, deprecated, err := parser.parsePath("@openapi GET /api/v1/test ")
	require.Nil(t, err)
	require.Equal(t, "get", method)
	require.Equal(t, "/api/v1/test", path)
	require.False(t, deprecated)

	// Test deprecated
	method, path, deprecated, err = parser.parsePath("@openapi GET /api/v1/test deprecated ")
	require.Nil(t, err)
	require.Equal(t, "get", method)
	require.Equal(t, "/api/v1/test", path)
	require.True(t, deprecated)
}

func TestParseParam(t *testing.T) {
	parser := NewParser(nil, newStructsParser())
	param, err := parser.parseParam("@openapiParam id in=path, type=int, example=11")

	require.Nil(t, err)
	require.Equal(t, "id", param.Name)
	require.Equal(t, "path", param.In)
	require.NotNil(t, param.Schema)
	require.Equal(t, "integer", param.Schema.Type)
	require.Equal(t, "11", param.Schema.Example)
}

func TestParseRequest(t *testing.T) {
	parser := NewParser(&Doc{
		OpenAPI:    "3.0.0",
		Paths:      map[string]Path{},
		Components: Component{map[string]SecurityScheme{}, map[string]*Schema{}}},
		newStructsParser())

	// Test invalid schema
	body, err := parser.parseRequest(`@openapiRequest application/json {"foo", "bar"}`, File{})
	require.NotNil(t, err)
	require.Equal(t, "Invalid JSON schema", err.Error())

	// Test unsupported content type
	body, err = parser.parseRequest(`@openapiRequest text/plain {}`, File{})
	require.NotNil(t, err)
	require.Equal(t, "Unsupported Content-Type", err.Error())

	// Test json example
	body, err = parser.parseRequest(`@openapiRequest application/json {"foo": "bar"}`, File{})
	require.Nil(t, err)

	content, e := body.Content["application/json"]
	require.True(t, e)
	require.Equal(t, `{"foo": "bar"}`, content.Example)

	// Test struct
	body, err = parser.parseRequest(`@openapiRequest application/json User`, getFile(t, "tests", "tests/structs.go", ""))
	require.Nil(t, err)

	content, e = body.Content["application/json"]
	require.True(t, e)
	require.Equal(t, "", content.Example)
	require.Equal(t, "", content.Schema.Type)
	require.Equal(t, "#/components/schemas/User", content.Schema.Ref)

	// Test array of structs
	body, err = parser.parseRequest(`@openapiRequest application/json []User`, getFile(t, "tests", "tests/structs.go", ""))
	require.Nil(t, err)

	content, e = body.Content["application/json"]
	require.True(t, e)
	require.Equal(t, "", content.Example)
	require.Equal(t, "array", content.Schema.Type)
	require.Equal(t, "#/components/schemas/User", content.Schema.Items.Ref)

	// Test json schema
	body, err = parser.parseRequest(`@openapiRequest application/json {"user": User, "id": int}`, getFile(t, "tests", "tests/structs.go", ""))
	require.Nil(t, err)

	content, e = body.Content["application/json"]
	require.True(t, e)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "integer", content.Schema.Properties["id"].Type)
	require.Equal(t, "#/components/schemas/User", content.Schema.Properties["user"].Ref)

	// Test json schema with array
	body, err = parser.parseRequest(`@openapiRequest application/json {"user": []User, "id": int}`, getFile(t, "tests", "tests/structs.go", ""))
	require.Nil(t, err)

	content, e = body.Content["application/json"]
	require.True(t, e)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "array", content.Schema.Properties["user"].Type)
	require.Equal(t, "#/components/schemas/User", content.Schema.Properties["user"].Items.Ref)

	// Test json schema with nested struct
	body, err = parser.parseRequest(`@openapiRequest application/json {"user": nested.NestedStruct, "id": int}`, getFile(t, "tests", "tests/structs4.go", ""))
	require.Nil(t, err)

	content, e = body.Content["application/json"]
	require.True(t, e)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "#/components/schemas/NestedStruct", content.Schema.Properties["user"].Ref)

	// Test json schema with array of nested structs
	body, err = parser.parseRequest(`@openapiRequest application/json {"user": []nested.NestedStruct, "id": int}`, getFile(t, "tests", "tests/structs4.go", ""))
	require.Nil(t, err)

	content, e = body.Content["application/json"]
	require.True(t, e)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "array", content.Schema.Properties["user"].Type)
	require.Equal(t, "#/components/schemas/NestedStruct", content.Schema.Properties["user"].Items.Ref)
}

func TestParseResponse(t *testing.T) {
	parser := NewParser(&Doc{
		OpenAPI:    "3.0.0",
		Paths:      map[string]Path{},
		Components: Component{map[string]SecurityScheme{}, map[string]*Schema{}}},
		newStructsParser())

	// Test invalid schema
	status, contentType, content, err := parser.parseResponse(`@openapiResponse 200 application/json {"foo", "bar"}`, File{})
	require.NotNil(t, err)
	require.Equal(t, "Invalid JSON schema", err.Error())

	// Test unsupported content type
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 text/xml {"foo": "bar"}`, File{})
	require.NotNil(t, err)
	require.Equal(t, "Unsupported Content-Type", err.Error())

	// Test invalid status code
	status, contentType, content, err = parser.parseResponse(`@openapiResponse ddd application/json {"foo": "bar"}`, File{})
	require.NotNil(t, err)
	require.Equal(t, "Invalid HTTP status code", err.Error())

	// Test json example
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/json {"foo": "bar"}`, File{})
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/json", contentType)
	require.Equal(t, `{"foo": "bar"}`, content.Example)

	// Test struct
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/json User`, getFile(t, "tests", "tests/structs.go", ""))
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/json", contentType)
	require.Equal(t, "", content.Example)
	require.Equal(t, "", content.Schema.Type)
	require.Equal(t, "#/components/schemas/User", content.Schema.Ref)

	// Test array of structs
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/json []User`, getFile(t, "tests", "tests/structs.go", ""))
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/json", contentType)
	require.Equal(t, "array", content.Schema.Type)
	require.Equal(t, "#/components/schemas/User", content.Schema.Items.Ref)

	// Test json schema
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/json {"user": User, "id": int}`, File{})
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/json", contentType)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "integer", content.Schema.Properties["id"].Type)

	// Test json schema with array
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/json {"user": []User, "id": int}`, getFile(t, "tests", "tests/structs.go", ""))
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/json", contentType)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "array", content.Schema.Properties["user"].Type)
	require.Equal(t, "#/components/schemas/User", content.Schema.Properties["user"].Items.Ref)

	// Test json schema with nested struct
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/json {"user": nested.NestedStruct, "id": int}`, getFile(t, "tests", "tests/structs4.go", ""))
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/json", contentType)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "#/components/schemas/NestedStruct", content.Schema.Properties["user"].Ref)

	// Test json schema with array of nested structs
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/json {"user": []nested.NestedStruct, "id": int}`, getFile(t, "tests", "tests/structs4.go", ""))
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/json", contentType)
	require.Equal(t, "", content.Example)
	require.Equal(t, "object", content.Schema.Type)
	require.Equal(t, "array", content.Schema.Properties["user"].Type)
	require.Equal(t, "#/components/schemas/NestedStruct", content.Schema.Properties["user"].Items.Ref)

	// Test application/octet-stream
	status, contentType, content, err = parser.parseResponse(`@openapiResponse 200 application/octet-stream`, File{})
	require.Nil(t, err)
	require.Equal(t, "200", status)
	require.Equal(t, "application/octet-stream", contentType)
	require.Equal(t, "string", content.Schema.Type)
	require.Equal(t, "binary", content.Schema.Format)
}

func TestParseStruct(t *testing.T) {
	doc := Doc{
		OpenAPI:    "3.0.0",
		Paths:      map[string]Path{},
		Components: Component{map[string]SecurityScheme{}, map[string]*Schema{}},
	}
	parser := NewParser(&doc, newStructsParser())

	s, err := parser.parseStruct(&ParsedType{
		Name: "Test",
		Kind: structType,
		File: getFile(t, "tests", "tests/uuid_structs.go", ""),
	})
	require.NotNil(t, err)
	require.Equal(t, "struct type with name 'Test' was not found in package 'tests' with import path ''", err.Error())

	s, err = parser.parseStruct(&ParsedType{
		Name: "UUIDUser",
		Kind: structType,
		File: getFile(t, "tests", "tests/uuid_structs.go", ""),
	})
	require.Nil(t, err)
	require.Equal(t, "", s.Type)
	require.NotEqual(t, "", s.Ref)

	user := doc.Components.Schemas["UUIDUser"]
	require.Equal(t, 4, len(user.Properties))
	require.Equal(t, "uuid", user.Properties["id"].Format)
	require.Equal(t, []string{"id", "group"}, user.Required)

	require.Equal(t, []string{"admin", "manager", "user"}, user.Properties["group"].Enum)
	require.Equal(t, "user", user.Properties["group"].Default)

	require.Equal(t, "testExample", user.Properties["description"].Example)
	require.Equal(t, "testDescription", user.Properties["description"].Description)

	// using cache
	s, err = parser.parseStruct(&ParsedType{
		Name: "[]UUIDUser",
		Kind: arrayType,
		File: File{},
		Nested: &ParsedType{
			Name: "UUIDUser",
			Kind: structType,
			File: File{},
		},
	})
	require.Nil(t, err)
	require.Equal(t, "array", s.Type)

	// nested struct
	s, err = parser.parseStruct(&ParsedType{
		Name: "Struct4",
		Kind: structType,
		File: getFile(t, "tests", "tests/structs4.go", ""),
	})
	require.Nil(t, err)
	require.Equal(t, "", s.importPath)
	require.Equal(t, "github.com/onrik/gaws/tests/nested", parser.doc.Components.Schemas["NestedStruct"].importPath)

	// even more nested struct with duplicate module name
	require.Equal(t, "github.com/onrik/gaws/tests/nested/nested", parser.doc.Components.Schemas["nested.NestedStruct"].importPath)
}

func TestTypeToProperty(t *testing.T) {
	parser := NewParser(&Doc{
		OpenAPI:    "3.0.0",
		Paths:      map[string]Path{},
		Components: Component{map[string]SecurityScheme{}, map[string]*Schema{}}}, newStructsParser())

	p, err := parser.typeToProperty(parser.mustParseType("int", File{}))
	require.Nil(t, err)
	require.Equal(t, "integer", p.Type)
	require.Equal(t, "", p.Format)

	p, err = parser.typeToProperty(parser.mustParseType("*int", File{}))
	require.Nil(t, err)
	require.Equal(t, "integer", p.Type)
	require.Equal(t, "", p.Format)

	p, err = parser.typeToProperty(parser.mustParseType("string", File{}))
	require.Nil(t, err)
	require.Equal(t, "string", p.Type)
	require.Equal(t, "", p.Format)

	p, err = parser.typeToProperty(parser.mustParseType("time.Time", File{}))
	require.Nil(t, err)
	require.Equal(t, "string", p.Type)
	require.Equal(t, "date-time", p.Format)

	p, err = parser.typeToProperty(parser.mustParseType("*time.Time", File{}))
	require.Nil(t, err)
	require.Equal(t, "string", p.Type)
	require.Equal(t, "date-time", p.Format)

	p, err = parser.typeToProperty(parser.mustParseType("[]string", File{}))
	require.Nil(t, err)
	require.Equal(t, "array", p.Type)
	require.NotNil(t, p.Items)
	require.Equal(t, "string", p.Items.Type)

	p, err = parser.typeToProperty(parser.mustParseType("User", getFile(t, "tests", "tests/structs.go", "")))
	require.Nil(t, err)
	require.Equal(t, "#/components/schemas/User", p.Ref)

	// alias
	p, err = parser.typeToProperty(parser.mustParseType("Alias", getFile(t, "tests", "tests/alias_structs.go", "")))
	require.Nil(t, err)
	require.Equal(t, "#/components/schemas/StructForAlias", p.Ref)
	require.Equal(t, map[string]Property{"name": {Type: "string"}}, parser.doc.Components.Schemas["StructForAlias"].Properties)

	// nested alias
	p, err = parser.typeToProperty(parser.mustParseType("NestedAlias", getFile(t, "tests", "tests/alias_structs.go", "")))
	require.Nil(t, err)
	require.Equal(t, "#/components/schemas/NestedStruct", p.Ref)

	// alias for simple type
	p, err = parser.typeToProperty(parser.mustParseType("SimpleAlias", getFile(t, "tests", "tests/alias_structs.go", "")))
	require.Nil(t, err)
	require.Equal(t, "string", p.Type)

	// alias for nested simple type
	p, err = parser.typeToProperty(parser.mustParseType("NestedSimpleAlias", getFile(t, "tests", "tests/alias_structs.go", "")))
	require.Nil(t, err)
	require.Equal(t, "", p.Ref)
	require.Equal(t, "string", p.Type)
}

func TestParseTags(t *testing.T) {
	tags := parseTags("@openapiTags foo,  bar")
	require.Equal(t, []string{"foo", "bar"}, tags)
}

func TestParseType(t *testing.T) {
	parser := NewParser(&Doc{
		OpenAPI:    "3.0.0",
		Paths:      map[string]Path{},
		Components: Component{map[string]SecurityScheme{}, map[string]*Schema{}}}, newStructsParser())

	_, err := parser.parseType("User", getFile(t, "nested", "tests/nested/nested.go", "github.com/onrik/gaws/tests/nested"))
	require.NotNil(t, err)
	require.Equal(t, "type with name 'User' was not found in package 'tests/nested' with import path 'github.com/onrik/gaws/tests/nested'", err.Error())
}
