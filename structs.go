package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"reflect"
	"strings"
)

type StructField struct {
	Name       string
	Type       string
	Tag        string
	IsExported bool
	IsPointer  bool
}

type Struct struct {
	Pkg    string
	Name   string
	Origin string
	Fields []StructField
}

func newStructsParser() *structsParser {
	return &structsParser{
		structs: map[string]map[string]Struct{},
	}
}

type structsParser struct {
	// importName -> local struct type name -> struct definition
	// fo example "github.com/onrik/gaws" -> "Type" -> Struct{}
	structs map[string]map[string]Struct
}

// parse parses structs from given go package
// pkg - Package to parse
func (p *structsParser) parse(pkg Package) (map[string]Struct, error) {
	if resp, ok := p.structs[pkg.ImportPath]; ok {
		return resp, nil
	}

	resp := map[string]Struct{}
	p.structs[pkg.ImportPath] = resp

	pkgs, err := parser.ParseDir(token.NewFileSet(), pkg.FSPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, parsedPkg := range pkgs {
		for name, f := range parsedPkg.Files {
			if strings.HasSuffix(name, "_test.go") {
				continue
			}
			ast.Inspect(f, p.inspectFile(pkg.ImportPath))
		}
	}

	return resp, nil
}

func (p *structsParser) inspectFile(importPath string) func(node ast.Node) bool {
	return func(node ast.Node) bool {
		t, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		// skip unexported structs from underlying packages
		if !t.Name.IsExported() && importPath != "" {
			return true
		}

		if alias := checkIsAlias(t.Type); alias != "" {
			p.structs[importPath][t.Name.Name] = Struct{
				Pkg:    importPath,
				Name:   t.Name.Name,
				Origin: alias,
			}
			return true
		}

		s, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}

		fields := []StructField{}
		for _, field := range s.Fields.List {
			f := StructField{}
			if len(field.Names) > 0 {
				f.Name = field.Names[0].Name
				f.IsExported = field.Names[0].IsExported()
			}
			if field.Tag != nil {
				f.Tag = field.Tag.Value
			}
			if !f.IsExported {
				continue
			}
			f.Type = getType(field.Type)
			f.IsPointer = strings.HasPrefix(f.Type, "*")
			if f.Type == "" {
				continue
			}

			fields = append(fields, f)
		}

		p.structs[importPath][t.Name.Name] = Struct{
			Pkg:    importPath,
			Name:   t.Name.Name,
			Fields: fields,
		}

		return true
	}
}

func getType(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + getType(t.Elt)
	case *ast.SelectorExpr:
		return getType(t.X) + "." + getType(t.Sel)
	case *ast.StarExpr:
		return "*" + getType(t.X)
	case *ast.MapType:
		return "map"
	case *ast.FuncType, *ast.ChanType, *ast.StructType, *ast.InterfaceType:
		// TODO
		return ""
	default:
		log.Println("Unsupported type:", reflect.TypeOf(t))
		return ""
	}
}

func checkIsAlias(e ast.Expr) string {
	switch e.(type) {
	case *ast.SelectorExpr, *ast.ArrayType, *ast.Ident:
		return getType(e)
	default:
		return ""
	}
}
