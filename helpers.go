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

func getParamsFromTag(s string) (map[string]string, error) {
	t := reflect.StructTag(strings.Trim(s, "`"))
	params, err := parseParams(t.Get("openapi"))
	if err != nil {
		return nil, err
	}

	params["description"] = t.Get("openapiDesc")
	params["enum"] = t.Get("openapiEnum")
	if params["example"] == "" {
		params["example"] = t.Get("openapiExample")
	}

	return params, nil
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

func addPkg(pkg, name string) string {
	if pkg == "" || strings.HasPrefix(name, pkg) || strings.Contains(name, ".") {
		return name
	}

	return pkg + "." + name
}

func structByName(structs []Struct, name string) *Struct {
	for i := range structs {
		if structs[i].Name == name {
			return &structs[i]
		}
	}

	return nil
}
