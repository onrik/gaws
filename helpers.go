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

func getFormatFromTag(s string) (map[string]string, error) {
	t := reflect.StructTag(strings.Trim(s, "`"))
	o := t.Get("openapi")
	if o == "" {
		return nil, nil
	}
	return parseParams(o)
}

func parseParams(s string) (map[string]string, error) {
	ss := strings.Split(s, ",")
	params := map[string]string{}
	for _, s := range ss {
		splits := strings.Split(s, "=")
		params[trim(splits[0])] = trim(getStr(splits, 1))
	}

	return params, nil
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
	if pkg == "" || strings.HasPrefix(name, pkg) {
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
