package main

import (
	"fmt"
	"net/url"
)

var (
	httpMethods = []string{"get", "head", "post", "put", "delete", "connect", "options", "trace", "patch"}

	paramIn    = []string{"path", "query", "header"}
	paramTypes = []string{"string", "integer", "number", "boolean", "object", "array"}
	// paramFormats = []string{"float", "double", "date", "date-time", "byte", "binary", "email", "uuid", "uri", "hostname", "ipv4", "ipv6"}

	requestContentTypes  = []string{"application/json", "multipart/form-data"}
	responseContentTypes = []string{"text/plain", "application/json", "application/octet-stream"}
)

func validatePath(method, path string) error {
	if !strIn(method, httpMethods) {
		return fmt.Errorf("Unknown HTTP method")
	}
	if _, err := url.ParseRequestURI(path); err != nil {
		return fmt.Errorf("Invalid HTTP path")
	}
	return nil
}

func validateParam(p Parameter) error {
	if p.Name == "" {
		return fmt.Errorf("Invalid param name")
	}
	if !strIn(p.In, paramIn) {
		return fmt.Errorf("Invalid param 'in'")
	}
	if !strIn(p.Schema.Type, paramTypes) {
		return fmt.Errorf("Invalid param 'type'")
	}

	return nil
}

func validateRequest(body RequestBody) error {
	for c := range body.Content {
		if !strIn(c, requestContentTypes) {
			return fmt.Errorf("Unsupported Content-Type")
		}
	}

	return nil
}

func validateResponse(status, contentType string, content Content) error {
	s := atoi(status)
	if s < 100 || s > 526 {
		return fmt.Errorf("Invalid HTTP status code")
	}

	if !strIn(contentType, responseContentTypes) {
		return fmt.Errorf("Unsupported Content-Type")
	}

	return nil
}
