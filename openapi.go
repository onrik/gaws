package main

type Property struct {
	Type                 string              `yaml:"type,omitempty"`
	Description          string              `yaml:"description,omitempty"`
	Format               string              `yaml:"format,omitempty"`
	Minimum              int                 `yaml:"minimum,omitempty"`
	Maximum              int                 `yaml:"maximum,omitempty"`
	Enum                 []string            `yaml:"enum,omitempty"`
	Default              string              `yaml:"default,omitempty"`
	Example              string              `yaml:"example,omitempty"`
	Style                string              `yaml:"style,omitempty"`
	Explode              bool                `yaml:"explode,omitempty"`
	Properties           map[string]Property `yaml:"properties,omitempty"`
	AdditionalProperties *map[string]string  `yaml:"additionalProperties,omitempty"`
	Items                *Schema             `yaml:"items,omitempty"`
	Ref                  string              `yaml:"$ref,omitempty"`
}

type Schema struct {
	// service field for deduplication
	importPath string              `yaml:"-"`
	Type       string              `yaml:"type,omitempty"`
	Format     string              `yaml:"format,omitempty"`
	Ref        string              `yaml:"$ref,omitempty"`
	Properties map[string]Property `yaml:"properties,omitempty"`
	Required   []string            `yaml:"required,omitempty"`
	Items      *Schema             `yaml:"items,omitempty"`
}

type Content struct {
	Schema  *Schema `yaml:"schema,omitempty"`
	Example string  `yaml:"example,omitempty"`
}

type Parameter struct {
	Name        string    `yaml:"name,omitempty"`
	In          string    `yaml:"in,omitempty"`
	Required    bool      `yaml:"required"`
	Description string    `yaml:"description,omitempty"`
	Schema      *Property `yaml:"schema,omitempty"`
}

type Response struct {
	Description string             `yaml:"description"`
	Content     map[string]Content `yaml:"content,omitempty"`
	Headers     map[string]Content `yaml:"headers,omitempty"`
	Ref         string             `yaml:"$ref,omitempty"`
}

type RequestBody struct {
	Description string             `yaml:"description,omitempty"`
	Content     map[string]Content `yaml:"content,omitempty"`
	Ref         string             `yaml:"$ref,omitempty"`
}

type Endpoint struct {
	Tags        []string              `yaml:"tags,omitempty"`
	Summary     string                `yaml:"summary,omitempty"`
	Description string                `yaml:"description,omitempty"`
	Parameters  []Parameter           `yaml:"parameters,omitempty"`
	RequestBody RequestBody           `yaml:"requestBody,omitempty"`
	Deprecated  bool                  `yaml:"deprecated,omitempty"`
	Responses   map[string]Response   `yaml:"responses,omitempty"`
	Security    []map[string][]string `yaml:"security,omitempty"` // TODO
}

type Path map[string]Endpoint

type Component struct {
	SecuritySchemes map[string]SecurityScheme `yaml:"securitySchemes,omitempty"`
	Schemas         map[string]*Schema        `yaml:"schemas,omitempty"`
}

type Doc struct {
	OpenAPI    string          `yaml:"openapi,omitempty"`
	Info       InfoProps       `yaml:"info,omitempty"`
	Servers    []Server        `yaml:"servers,omitempty"`
	BasePath   string          `yaml:"basePath,omitempty"`
	Paths      map[string]Path `yaml:"paths,omitempty"`
	Components Component       `yaml:"components,omitempty"`
}

type Server struct {
	URL string `yaml:"url,omitempty"`
}

type SecurityScheme struct {
	Scheme string `yaml:"scheme,omitempty"`
	Type   string `yaml:"type"`
	Name   string `yaml:"name"`
	In     string `yaml:"in"`
}

type InfoProps struct {
	Description string `yaml:"description,omitempty"`
	Title       string `yaml:"title,omitempty"`
	Version     string `yaml:"version,omitempty"`
}
