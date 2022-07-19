package main

type Property struct {
	Type        string              `yaml:"type,omitempty"`
	Description string              `yaml:"description,omitempty"`
	Format      string              `yaml:"format,omitempty"`
	Minimum     int                 `yaml:"minimum,omitempty"`
	Maximum     int                 `yaml:"maximum,omitempty"`
	Enum        []string            `yaml:"enum,omitempty"`
	Default     string              `yaml:"default,omitempty"`
	Example     string              `yaml:"example,omitempty"`
	Style       string              `yaml:"style,omitempty"`
	Explode     bool                `yaml:"explode,omitempty"`
	Properties  map[string]Property `yaml:"properties,omitempty"`
	Items       *Schema             `yaml:"items,omitempty"`
}

type Schema struct {
	Type       string              `yaml:"type,omitempty"`
	Format     string              `yaml:"format,omitempty"`
	Ref        string              `yaml:"$ref,omitempty"`
	Properties map[string]Property `yaml:"properties,omitempty"`
	Items      *Schema             `yaml:"items,omitempty"`
}

type Content struct {
	Schema  Schema `yaml:"schema,omitempty"`
	Example string `yaml:"example,omitempty"`
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
	Tags        []string            `yaml:"tags,omitempty"`
	Summary     string              `yaml:"summary,omitempty"`
	Description string              `yaml:"description,omitempty"`
	Parameters  []Parameter         `yaml:"parameters,omitempty"`
	RequestBody RequestBody         `yaml:"requestBody,omitempty"`
	Deprecated  bool                `yaml:"deprecated,omitempty"`
	Responses   map[string]Response `yaml:"responses,omitempty"`
}

type Path map[string]Endpoint

type Component struct {
	Schemas map[string]Schema `yaml:"schemas,omitempty"`
}

type Doc struct {
	OpenAPI    string               `yaml:"openapi,omitempty"`
	Paths      map[string]Path      `yaml:"paths,omitempty"`
	Components map[string]Component `yaml:"components,omitempty"`
}
