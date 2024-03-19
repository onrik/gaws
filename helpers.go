package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func getStr(ss []string, i int) string {
	if len(ss) <= i {
		return ""
	}

	return ss[i]
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

func upper(s string) string {
	return strings.ToUpper(s)
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func getTag(s, tag string) string {
	t := reflect.StructTag(strings.Trim(s, "`"))
	return strings.Split(t.Get(tag), ",")[0]
}

type Tags struct {
	Openapi     map[string]string
	Description string
	Enum        string
	Example     string
	Extensions  map[string]string
}

func getParamsFromTag(s string) (Tags, error) {
	t := reflect.StructTag(strings.Trim(s, "`"))
	params, err := parseParams(t.Get("openapi"))
	if err != nil {
		return Tags{}, err
	}

	description := t.Get("openapiDesc")
	enum := t.Get("openapiEnum")
	example := params["example"]
	if example == "" {
		example = t.Get("openapiExample")
	}

	extensions, err := parseParams(t.Get("openapiExt"))
	if err != nil {
		return Tags{}, err
	}

	return Tags{
		Openapi:     params,
		Description: description,
		Enum:        enum,
		Example:     example,
		Extensions:  extensions,
	}, nil
}

func parseParams(str string) (map[string]string, error) {
	temp := []rune{}
	params := map[string]string{}
	opened := 0
	for _, s := range str {
		if s == '{' {
			opened++
		}
		if s == '}' {
			opened--
		}

		if s != ',' || opened > 0 {
			temp = append(temp, rune(s))
			continue
		}

		key, value, err := parseParam(string(temp))
		if err != nil {
			return nil, err
		}
		if key != "" {
			params[key] = value
		}
		temp = []rune{}
	}

	key, value, err := parseParam(string(temp))
	if err != nil {
		return nil, err
	}
	if key != "" {
		params[key] = value
	}

	return params, nil
}

func parseParam(str string) (string, string, error) {
	ss := strings.Split(str, "=")
	key := trim(ss[0])
	value := strings.ReplaceAll(trim(getStr(ss, 1)), `'`, `"`)

	return key, value, nil

}

func strIn(s string, ss []string) bool {
	for i := range ss {
		if ss[i] == s {
			return true
		}
	}

	return false
}

func parseJSONSchema(s string) (map[string]string, error) {
	m := map[string]string{}
	s = strings.TrimSuffix(strings.TrimPrefix(s, "{"), "}")
	for _, l := range strings.Split(s, ",") {
		l = trim(l)
		if len(l) == 0 {
			continue
		}
		splits := strings.Split(l, ":")
		key := strings.Trim(splits[0], `"`)
		val := trim(getStr(splits, 1))
		if val == "" {
			return nil, fmt.Errorf("Invalid JSON schema")
		}
		m[key] = val

	}

	return m, nil
}

func getPkg(name string) string {
	if !strings.Contains(name, ".") {
		return ""
	}

	return strings.SplitN(name, ".", 3)[0]
}

func splitName(name string) (string, string) {
	first, second, found := strings.Cut(name, ".")
	if found {
		return first, second
	}

	return "", first
}

func addPkg(pkg, name string) string {
	if pkg == "" || strings.HasPrefix(name, pkg) || strings.Contains(name, ".") {
		return name
	}

	return pkg + "." + name
}

// getSchemaNameForStruct searches in schemas struct with given name
//
//	If given name already used by another struct then getSchemaForStruct tries to construct unique
//	name for given struct
//	Returns new unique name for given struct which should be used on struct insert into schemas
func getSchemaNameForStruct(schemas map[string]*Schema, name string, st Struct) (string, bool) {
	existingSchema, ok := schemas[name]
	if ok {
		if existingSchema.importPath == st.Pkg {
			return name, true
		} else if st.Pkg != "" {
			// dealing with duplicates from different packages like
			//
			// file1.go
			//    "github.com/onrik/gaws/tests/nested"
			//
			//    nested.Type{}
			//
			// file2.go
			//    "github.com/onrik/gaws/tests/nested/nested"
			//
			//     nested.Type{}
			//
			pkgChunks := strings.Split(st.Pkg, "/")
			// Trying to find unique identifier for our schema
			// For example for struct Type in package "github.com/onrik/gaws/tests/nested" we will try
			// Type -> nested.Type -> tests.nested.Type -> etc
			for i := len(pkgChunks) - 1; i >= 0; i-- {
				name = fmt.Sprintf("%s.%s", pkgChunks[i], name)
				existingSchema, ok = schemas[name]
				if ok {
					if existingSchema.importPath == st.Pkg {
						return name, true
					}
				} else {
					break
				}
			}
		}
	}

	return name, false
}
