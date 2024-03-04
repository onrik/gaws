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
			property, err := p.typeToProperty("", t, file)
			if err != nil {
				return content, err
			}
			content.Schema.Properties[n] = property
		}

		return content, nil
	}

	schema, err := p.parseStruct(s, file)
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

// parseStruct
func (p *Parser) parseStruct(s string, file File) (*Schema, error) {
	schema := &Schema{
		Properties: map[string]Property{},
	}

	if strings.HasPrefix(s, "[]") {
		arraySchema, err := p.parseStruct(strings.TrimPrefix(s, "[]"), file)
		if err != nil {
			return nil, err
		}
		schema.Type = "array"
		schema.Properties = nil
		schema.Items = arraySchema
		return schema, nil
	}

	pkg, name := splitName(s)
	if pkg != "" {
		fileWithImportedType, err := file.ParseImport(pkg, name)
		if err != nil {
			return nil, err
		}

		return p.parseStruct(name, fileWithImportedType)
	}

	structs, err := p.structs.parse(file.Pkg)
	if err != nil {
		return nil, err
	}

	st, ok := structs[name]
	if !ok {
		return nil, fmt.Errorf("struct type with name '%s' was not found in package '%s' with import path '%s'", name, file.Pkg.FSPath, file.Pkg.ImportPath)
	}

	s, ok = getSchemaNameForStruct(p.doc.Components.Schemas, s, st)
	if ok {
		return &Schema{
			importPath: st.Pkg,
			Ref:        fmt.Sprintf("#/components/schemas/%s", s),
		}, nil
	}

	// add struct schema to schemas before full parsing to prevent loop calls parseStruct -> typeToProperty -> parseStruct
	p.doc.Components.Schemas[s] = schema

	schema.importPath = st.Pkg
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
			return nil, err
		}

		property := Property{
			Type: params["type"],
		}
		if property.Type == "" {
			property, err = p.typeToProperty(getPkg(s), st.Fields[i].Type, file)
			if err != nil {
				return nil, err
			}
		}

		if params["format"] != "" {
			property.Format = params["format"]
		}
		if params["example"] != "" {
			property.Example = params["example"]
		}
		if params["description"] != "" {
			property.Description = params["description"]
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

	return &Schema{
		importPath: st.Pkg,
		Ref:        fmt.Sprintf("#/components/schemas/%s", s),
	}, nil
}

func (p *Parser) typeToProperty(pkg, t string, file File) (property Property, err error) {
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

	if strings.HasPrefix(t, "[]") {
		property.Type = "array"
		prop, err := p.typeToProperty(pkg, strings.TrimPrefix(t, "[]"), file)
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

	// Redeclarated type
	if pkg == "" {
		ot, ok, err := p.parseAlias(t, file)
		if err != nil {
			return Property{}, err
		}
		if ok {
			return p.typeToProperty(pkg, ot, file)
		}
	}

	// Is struct
	schema, err := p.parseStruct(addPkg(pkg, t), file)
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

func isTime(t string) bool {
	return t == "time.Time"
}

func wrapError(err error, comment string) error {
	return fmt.Errorf("%s (%s)", err.Error(), comment)
}
