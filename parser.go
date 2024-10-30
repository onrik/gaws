package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	pathPrefix     = "@openapi "
	paramPrefix    = "@openapiParam "
	tagsPrefix     = "@openapiTags "
	summaryPrefix  = "@openapiSummary "
	descPrefix     = "@openapiDesc "
	requestPrefix  = "@openapiRequest "
	responsePrefix = "@openapiResponse "
	securityPrefix = "@openapiSecurity"
)

var (
	typesMap = map[string]string{
		"int":     "integer",
		"int8":    "integer",
		"int16":   "integer",
		"int32":   "integer",
		"int64":   "integer",
		"uint":    "integer",
		"uint8":   "integer",
		"uint16":  "integer",
		"uint32":  "integer",
		"uint64":  "integer",
		"float":   "number",
		"float32": "number",
		"float64": "number",
		"bool":    "boolean",
		"string":  "string",
		"byte":    "string",
		"[]byte":  "string",
	}

	formatsMap = map[string]string{
		"float":   "float",
		"float32": "float",
		"float64": "double",
		"[]byte":  "binary",
	}
)

type Parser struct {
	structs *structsParser
	doc     *Doc
}

func NewParser(doc *Doc, structs *structsParser) *Parser {
	return &Parser{
		structs: structs,
		doc:     doc,
	}
}

func (p *Parser) parseComment(comment string, file File) (err error) {
	splits := strings.Split(comment, "\n")
	paths := map[string]map[string]bool{}
	endpoint := Endpoint{
		Responses: map[string]Response{},
		Security:  make([]map[string][]string, 0),
	}

	for _, l := range splits {
		if strings.HasPrefix(l, pathPrefix) {
			method, path, deprecated, err := p.parsePath(l)
			if err != nil {
				return wrapError(err, l)
			}
			if _, ok := paths[path]; !ok {
				paths[path] = map[string]bool{}
			}
			paths[path][method] = deprecated
		}
		if strings.HasPrefix(l, tagsPrefix) {
			endpoint.Tags = parseTags(l)
		}

		if strings.HasPrefix(l, summaryPrefix) {
			endpoint.Summary = parseSummary(l)
		}

		if strings.HasPrefix(l, descPrefix) {
			endpoint.Description = parseDesc(l)
		}

		if strings.HasPrefix(l, paramPrefix) {
			param, err := p.parseParam(l)
			if err != nil {
				return wrapError(err, l)
			}
			endpoint.Parameters = append(endpoint.Parameters, param)
		}

		if strings.HasPrefix(l, requestPrefix) {
			request, err := p.parseRequest(l, file)
			if err != nil {
				return wrapError(err, l)
			}
			endpoint.RequestBody = request
		}

		if strings.HasPrefix(l, responsePrefix) {
			status, contentType, content, err := p.parseResponse(l, file)
			if err != nil {
				return wrapError(err, l)
			}

			endpoint.Responses[status] = Response{
				Content: map[string]Content{
					contentType: content,
				},
			}
		}
		if strings.HasPrefix(l, securityPrefix) {
			sec := p.parseSecurity(l)
			if len(sec) > 0 {
				endpoint.Security = append(endpoint.Security, sec)
			}

		}
	}

	if len(paths) == 0 {
		return nil
	}

	if len(endpoint.Responses) == 0 {
		for path := range paths {
			for method := range paths[path] {
				return fmt.Errorf("no %s for: %s %s", trim(responsePrefix), upper(method), path)
			}
		}
	}

	for path := range paths {
		for method := range paths[path] {
			e := endpoint
			e.Deprecated = paths[path][method]
			if _, ex := p.doc.Paths[path]; !ex {
				p.doc.Paths[path] = Path{
					method: e,
				}
			} else {
				p.doc.Paths[path][method] = e
			}
		}
	}

	return nil
}

// parsePath @openapi GET /foo/bar
func (p *Parser) parsePath(s string) (method, path string, deprecated bool, err error) {
	s = strings.TrimPrefix(s, pathPrefix)
	splits := strings.Split(s, " ")

	method = strings.ToLower(trim(splits[0]))
	path = trim(getStr(splits, 1))
	deprecated = trim(getStr(splits, 2)) == "deprecated"
	err = validatePath(method, path)

	return
}

// parseRequest @openapiParam foo in=path, type=int, default=1, required=true, enum=1 2 3
func (p *Parser) parseParam(s string) (Parameter, error) {
	s = strings.TrimPrefix(s, paramPrefix)
	splits := strings.SplitN(s, " ", 2)

	params, err := parseParams(getStr(splits, 1))
	if err != nil {
		return Parameter{}, err
	}

	format := params["format"]
	if format == "" {
		format = formatsMap[params["type"]]
	}

	if !strIn(params["type"], paramTypes) {
		params["type"] = typesMap[params["type"]]
	}

	_, required := params["required"]
	if params["required"] == "" && params["in"] == "path" {
		required = true
	}

	var enum []string
	if enumValues, ok := params["enum"]; ok {
		enum = strings.Split(enumValues, " ")
	}

	param := Parameter{
		Name:     trim(splits[0]),
		In:       params["in"],
		Required: required,
		Schema: &Property{
			Type:        params["type"],
			Format:      format,
			Example:     params["example"],
			Default:     params["default"],
			Description: params["description"],
			Enum:        enum,
		},
	}

	return param, validateParam(param)
}

// parseRequest @openapiRequest application/json {"foo": "bar"}
func (p *Parser) parseRequest(s string, file File) (body RequestBody, err error) {
	s = strings.TrimPrefix(s, requestPrefix)
	splits := strings.SplitN(s, " ", 2)

	contentType := trim(splits[0])
	request := trim(getStr(splits, 1))

	content, err := p.parseSchema(request, file)
	if err != nil {
		return body, err
	}

	body.Content = map[string]Content{
		contentType: content,
	}
	return body, validateRequest(body)
}

// parseResponse @openapiResponse 200 application/json {"foo": "bar"}
func (p *Parser) parseResponse(s string, file File) (status string, contentType string, content Content, err error) {
	s = strings.TrimPrefix(s, responsePrefix)
	splits := strings.SplitN(s, " ", 3)

	status = trim(splits[0])
	contentType = trim(getStr(splits, 1))
	response := trim(getStr(splits, 2))

	if contentType == "application/octet-stream" {
		content.Schema = &Schema{
			Type:   "string",
			Format: "binary",
		}
	} else {
		content, err = p.parseSchema(response, file)
		if err != nil {
			return status, contentType, content, err
		}
	}

	err = validateResponse(status, contentType, content)

	return status, contentType, content, err
}

// parseSecurity @openapiSecurity api_key apiKey cookie AuthKey
// @openapiSecurity Name Type In KeyName
func (p *Parser) parseSecurity(commentLine string) map[string][]string {
	secMap := make(map[string][]string)
	splits := strings.SplitN(commentLine, " ", 5)

	securityName := trim(splits[1])
	schemeType := trim(splits[2])
	schemeIn := trim(splits[3])
	schemeName := trim(splits[4])
	p.doc.Components.SecuritySchemes = make(map[string]SecurityScheme)
	p.doc.Components.SecuritySchemes[securityName] = SecurityScheme{"", schemeType, schemeName, schemeIn}
	secMap[securityName] = []string{}
	return secMap
}

// parseSchema {"foo": "bar"}
func (p *Parser) parseSchema(s string, file File) (Content, error) {
	content := Content{}
	if strings.HasPrefix(s, "{") {
		if json.Valid([]byte(s)) {
			content.Example = s
			return content, nil
		}

		fields, err := parseJSONSchema(s)
		if err != nil {
			return content, err
		}

		content.Schema = &Schema{}
		content.Schema.Type = "object"
		content.Schema.Properties = map[string]Property{}
		for n, t := range fields {
			parsedType, err := p.parseType(t, file)
			if err != nil {
				return content, err
			}

			property, err := p.typeToProperty(parsedType)
			if err != nil {
				return content, err
			}
			content.Schema.Properties[n] = property
		}

		return content, nil
	}

	parsedType, err := p.parseType(s, file)
	if err != nil {
		return content, err
	}

	schema, err := p.parseStruct(parsedType)
	if err != nil {
		return content, err
	}
	content.Schema = schema

	return content, nil
}

// parseAlias parses given type and returns original type name and true if type is alias
func (p *Parser) parseAlias(s string, file File) (string, bool, error) {
	// alias can not have package
	if pkg := getPkg(s); pkg != "" {
		return "", false, nil
	}

	structs, err := p.structs.parse(file.Pkg)
	if err != nil {
		return "", false, err
	}

	st, ok := structs[s]
	if !ok {
		return "", false, fmt.Errorf("type with name '%s' was not found in package '%s' with import path '%s'", s, file.Pkg.FSPath, file.Pkg.ImportPath)
	}

	if st.Origin != "" {
		return st.Origin, true, nil
	}

	return "", false, nil
}

// parseStruct parses Schema from given ParsedType
func (p *Parser) parseStruct(t *ParsedType) (*Schema, error) {
	schema := &Schema{
		Properties: map[string]Property{},
	}

	switch t.Kind {
	case arrayType, structType:
	default:
		return nil, fmt.Errorf("expect array or struct parsed type kind, got: %d", t.Kind)
	}

	if t.Kind == arrayType {
		arraySchema, err := p.parseStruct(t.Nested)
		if err != nil {
			return nil, err
		}
		schema.Type = "array"
		schema.Properties = nil
		schema.Items = arraySchema
		return schema, nil
	}

	structs, err := p.structs.parse(t.File.Pkg)
	if err != nil {
		return nil, err
	}

	st, ok := structs[t.Name]
	if !ok {
		return nil, fmt.Errorf("struct type with name '%s' was not found in package '%s' with import path '%s'", t.Name, t.File.Pkg.FSPath, t.File.Pkg.ImportPath)
	}

	schemaName, ok := getSchemaNameForStruct(p.doc.Components.Schemas, t.Name, st)
	if ok {
		return &Schema{
			importPath: st.Pkg,
			Ref:        fmt.Sprintf("#/components/schemas/%s", schemaName),
		}, nil
	}

	// add struct schema to schemas before full parsing to prevent loop calls parseStruct -> typeToProperty -> parseStruct
	p.doc.Components.Schemas[schemaName] = schema

	schema.importPath = st.Pkg
	schema.Type = "object"
	for i := range st.Fields {
		tags, err := getParamsFromTag(st.Fields[i].Tag)
		if err != nil {
			return nil, err
		}

		if st.Fields[i].IsSystem {
			schema.Description = tags.Description
			continue
		}

		name := st.Fields[i].Name
		tag := getTag(st.Fields[i].Tag, "json")
		if tag == "-" {
			continue
		}
		if tag != "" {
			name = tag
		}

		property := Property{
			Type: tags.Openapi["type"],
		}
		if property.Type == "" {
			parsedType, err := p.parseType(st.Fields[i].Type, t.File)
			if err != nil {
				return nil, err
			}

			property, err = p.typeToProperty(parsedType)
			if err != nil {
				return nil, err
			}
		}

		if tags.Openapi["format"] != "" {
			property.Format = tags.Openapi["format"]
		}
		if tags.Example != "" {
			property.Example = tags.Example
		}
		if tags.Description != "" {
			property.Description = tags.Description
		}
		if tags.Openapi["default"] != "" {
			property.Default = tags.Openapi["default"]
		}
		if tags.Enum != "" {
			property.Enum = strings.Split(tags.Enum, ",")
		}
		if _, ok := tags.Openapi["required"]; ok {
			schema.Required = append(schema.Required, name)
		}
		if tags.Extensions != nil {
			property.Extensions = tags.Extensions
		}

		schema.Properties[name] = property
	}

	return &Schema{
		importPath: st.Pkg,
		Ref:        fmt.Sprintf("#/components/schemas/%s", schemaName),
	}, nil
}

// parseType parses given string and return ParsedType from it.
//
//	This function automagickally resolves pointers, types with packages and aliases.
func (p *Parser) parseType(t string, file File) (*ParsedType, error) {
	t = strings.TrimPrefix(t, "*")

	if isTime(t) {
		return &ParsedType{
			Kind: timeType,
			Name: t,
			File: file,
		}, nil
	}

	if isBaseType(t) {
		return &ParsedType{
			Kind: baseType,
			Name: t,
			File: file,
		}, nil
	}

	if t == "map" {
		return &ParsedType{
			Kind: mapType,
			Name: t,
			File: file,
		}, nil
	}

	if strings.HasPrefix(t, "[]") {
		nested, err := p.parseType(strings.TrimPrefix(t, "[]"), file)
		if err != nil {
			return nil, err
		}

		return &ParsedType{
			Kind:   arrayType,
			Name:   t,
			File:   file,
			Nested: nested,
		}, nil
	}

	pkg, t := splitName(t)
	if pkg != "" {
		fileWithImportedType, err := file.ParseImport(pkg, t)
		if err != nil {
			return nil, err
		}

		return p.parseType(t, fileWithImportedType)
	}

	ot, ok, err := p.parseAlias(t, file)
	if err != nil {
		return nil, err
	}
	if ok {
		return p.parseType(ot, file)
	}

	return &ParsedType{
		Kind: structType,
		Name: t,
		File: file,
	}, nil
}

func (p *Parser) mustParseType(t string, file File) *ParsedType {
	resp, err := p.parseType(t, file)
	if err != nil {
		panic(err)
	}

	return resp
}

// typeToProperty constructs Property from given ParsedType
func (p *Parser) typeToProperty(t *ParsedType) (property Property, err error) {
	switch t.Kind {
	case timeType:
		property.Type = "string"
		property.Format = "date-time"
		return

	case baseType:
		property.Type = typesMap[t.Name]
		property.Format = formatsMap[t.Name]
		return

	case mapType:
		property.Type = "object"
		property.AdditionalProperties = &map[string]string{}
		return

	case arrayType:
		property.Type = "array"
		prop, err := p.typeToProperty(t.Nested)
		if err != nil {
			return property, err
		}

		property.Items = &Schema{
			Type: prop.Type,
		}

		if prop.Type == "object" {
			property.Items.Properties = prop.Properties
		}

		if prop.Type == "array" {
			property.Items.Items = prop.Items
		}

		property.Items.Ref = prop.Ref
		if property.Items.Ref != "" {
			property.Items.Type = ""
		}

		return property, nil

	case structType:
		schema, err := p.parseStruct(t)
		if err != nil {
			return property, err
		}
		property.Ref = schema.Ref
		if property.Ref == "" {
			property.Type = "object"
		}

		property.Properties = schema.Properties
		return property, nil

	default:
		return property, fmt.Errorf("unknown parsed type kind: %d", t.Kind)
	}
}

// parseTags @openapiTags foo, bar
func parseTags(s string) []string {
	s = strings.TrimPrefix(s, tagsPrefix)
	tags := []string{}
	for _, t := range strings.Split(s, ",") {
		t = trim(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	return tags
}

// parseSummary @openapiSummary Some text
func parseSummary(s string) string {
	return strings.TrimPrefix(s, summaryPrefix)
}

// parseDesc @openapiDesc Some text
func parseDesc(s string) string {
	return strings.TrimPrefix(s, descPrefix)
}

func isBaseType(t string) bool {
	return typesMap[t] != ""
}

func isTime(t string) bool {
	return t == "time.Time"
}

func wrapError(err error, comment string) error {
	return fmt.Errorf("%s (%s)", err.Error(), comment)
}
