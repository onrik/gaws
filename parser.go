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
	structs []Struct
	doc     *Doc
	origins map[string]string
}

func NewParser(doc *Doc, structs []Struct, origins map[string]string) *Parser {
	return &Parser{
		structs: structs,
		doc:     doc,
		origins: origins,
	}
}

func (p *Parser) parseComment(comment string) (err error) {
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
			request, err := p.parseRequest(l)
			if err != nil {
				return wrapError(err, l)
			}
			endpoint.RequestBody = request
		}

		if strings.HasPrefix(l, responsePrefix) {
			status, contentType, content, err := p.parseResponse(l)
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

// parseRequest @openapiParam foo in=path, type=int, default=1, required=true
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
		},
	}

	return param, validateParam(param)
}

// parseRequest @openapiRequest application/json {"foo": "bar"}
func (p *Parser) parseRequest(s string) (body RequestBody, err error) {
	s = strings.TrimPrefix(s, requestPrefix)
	splits := strings.SplitN(s, " ", 2)

	contentType := trim(splits[0])
	request := trim(getStr(splits, 1))

	content, err := p.parseSchema(request)
	if err != nil {
		return body, err
	}

	body.Content = map[string]Content{
		contentType: content,
	}
	return body, validateRequest(body)
}

// parseResponse @openapiResponse 200 application/json {"foo": "bar"}
func (p *Parser) parseResponse(s string) (status string, contentType string, content Content, err error) {
	s = strings.TrimPrefix(s, responsePrefix)
	splits := strings.SplitN(s, " ", 3)

	status = trim(splits[0])
	contentType = trim(getStr(splits, 1))
	response := trim(getStr(splits, 2))

	if contentType == "application/octet-stream" {
		content.Schema = Schema{
			Type:   "string",
			Format: "binary",
		}
	} else {
		content, err = p.parseSchema(response)
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
func (p *Parser) parseSchema(s string) (Content, error) {
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

		content.Schema.Type = "object"
		content.Schema.Properties = map[string]Property{}
		for n, t := range fields {
			property, err := p.typeToProperty("", t, []string{})
			if err != nil {
				return content, err
			}
			content.Schema.Properties[n] = property
		}

		return content, nil
	}

	schema, err := p.parseStruct(s, []string{})
	if err != nil {
		return content, err
	}
	content.Schema = schema

	return content, nil
}

// parseStruct
func (p *Parser) parseStruct(s string, stack []string) (Schema, error) {
	var schema Schema
	if _, ok := p.doc.Components.Schemas[s]; ok {
		return Schema{Ref: fmt.Sprintf("#/components/schemas/%s", s)}, nil
	}

	for i := range stack {
		if stack[i] == s {
			return Schema{Ref: fmt.Sprintf("#/components/schemas/%s", s)}, nil
		}
	}

	schema = Schema{
		Properties: map[string]Property{},
	}

	if strings.HasPrefix(s, "[]") {
		arraySchema, err := p.parseStruct(strings.TrimPrefix(s, "[]"), stack)
		if err != nil {
			return schema, err
		}
		schema.Type = "array"
		schema.Properties = nil
		schema.Items = &arraySchema
		return schema, nil
	}

	stack = append(stack, s)
	st := p.structByName(s)
	if st == nil {
		return schema, fmt.Errorf("unknown type: %s", s)
	}

	schema.Type = "object"
	for i := range st.Fields {
		name := st.Fields[i].Name
		tag := getTag(st.Fields[i].Tag, "json")
		if tag == "-" {
			continue
		}
		if tag != "" {
			name = tag
		}

		params, err := getParamsFromTag(st.Fields[i].Tag)
		if err != nil {
			return schema, err
		}

		property := Property{
			Type: params["type"],
		}
		if property.Type == "" {
			property, err = p.typeToProperty(getPkg(s), st.Fields[i].Type, stack)
			if err != nil {
				return schema, err
			}
		}

		if params["format"] != "" {
			property.Format = params["format"]
		}
		if params["example"] != "" {
			property.Example = params["example"]
		}
		if params["description"] != "" {
			property.Example = params["description"]
		}
		if params["default"] != "" {
			property.Default = params["default"]
		}
		if params["enum"] != "" {
			property.Enum = strings.Split(params["enum"], ",")
		}
		if _, ok := params["required"]; ok {
			schema.Required = append(schema.Required, name)
		}

		schema.Properties[name] = property
	}

	p.doc.Components.Schemas[s] = schema

	return Schema{
		Ref: fmt.Sprintf("#/components/schemas/%s", s),
	}, nil
}

func (p *Parser) structByName(name string) *Struct {
	for i := range p.structs {
		if p.structs[i].Name == name {
			return &p.structs[i]
		}
	}

	return nil
}

func (p *Parser) typeToProperty(pkg, t string, stack []string) (property Property, err error) {
	t = strings.TrimPrefix(t, "*")

	if isBaseType(t) {
		property.Type = typesMap[t]
		property.Format = formatsMap[t]
		return
	}
	if isTime(t) {
		property.Type = "string"
		property.Format = "date-time"
		return
	}

	if t == "map" {
		property.Type = "object"
		property.AdditionalProperties = &map[string]string{}
		return
	}

	// Redeclarated type
	ot, hasOrigin := p.origin(t)
	if hasOrigin {
		return p.typeToProperty(pkg, ot, stack)
	}

	if strings.HasPrefix(t, "[]") {
		property.Type = "array"
		prop, err := p.typeToProperty(pkg, strings.TrimPrefix(t, "[]"), stack)
		if err != nil {
			return property, err
		}

		property.Items = &Schema{
			Type: property.Type,
		}

		property.Items.Items = prop.Items
		property.Items.Type = prop.Type
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
	}

	// Is struct
	schema, err := p.parseStruct(addPkg(pkg, t), stack)
	if err != nil {
		return property, err
	}
	property.Ref = schema.Ref
	if property.Ref == "" {
		property.Type = "object"
	}

	property.Properties = schema.Properties

	return property, nil
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

// origin return origin of redeclarated type and true flag
func (p *Parser) origin(t string) (string, bool) {
	ot := p.origins[t]
	if ot != "" && t != ot {
		return ot, true
	}
	return t, false
}

func isTime(t string) bool {
	return t == "time.Time"
}

func wrapError(err error, comment string) error {
	return fmt.Errorf("%s (%s)", err.Error(), comment)
}
