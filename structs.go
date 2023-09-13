package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
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

func parseStructs(prefix, path string, deep bool) (structsParser, error) {
	p := structsParser{
		prefix:    prefix,
		recursive: deep,
		origins:   make(map[string]string),
	}

	err := p.parse(path)

	return p, err
}

type structsParser struct {
	prefix    string
	recursive bool
	structs   []Struct
	origins   map[string]string
}

func (p *structsParser) parse(path string) error {
	pkgs, err := parser.ParseDir(token.NewFileSet(), path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	renamedImports := map[string]map[string]bool{}
	for _, pkg := range pkgs {
		for name, f := range pkg.Files {
			if strings.HasSuffix(name, "_test.go") {
				continue
			}
			for _, i := range f.Imports {
				if i.Name != nil {
					name := strings.Trim(i.Path.Value, `"`)
					if renamedImports[name] == nil {
						renamedImports[name] = map[string]bool{}
					}
					renamedImports[name][i.Name.Name] = true
					_, pkg := filepath.Split(name)
					renamedImports[name][pkg] = true
				}
			}
			ast.Inspect(f, p.inspectFile(pkg.Name))
		}
	}

	if p.prefix != "" {
		// Add prefix for package structs, for example Time -> time.Time
		for i := range p.structs {
			p.structs[i].Name = p.prefix + "." + p.structs[i].Name
		}
	}

	if p.recursive {
		config := &packages.Config{
			Mode: packages.LoadAllSyntax, // nolint: staticcheck
			Dir:  path,
		}

		pkgList, err := packages.Load(config)
		if err != nil {
			return err
		}

		for name, pkg := range pkgList[0].Imports {
			if len(pkg.GoFiles) == 0 {
				continue
			}

			names := []string{}
			pkgPath, _ := filepath.Split(pkg.GoFiles[0])
			if renamed := renamedImports[name]; renamed != nil {
				for n := range renamed {
					names = append(names, n)
				}
			} else {
				_, name = filepath.Split(name)
				names = append(names, name)
			}

			for i := range names {
				pkg, err := parseStructs(names[i], pkgPath, false)
				if err != nil {
					return err
				}
				p.structs = append(p.structs, pkg.structs...)

			}

			// For redeclarated types
			for i := range p.structs {
				if p.structs[i].Origin != "" {
					p.origins[p.structs[i].Name] = p.structs[i].Origin
				}
			}
		}
	}

	return nil
}

func (p *structsParser) inspectFile(pkgName string) func(node ast.Node) bool {
	return func(node ast.Node) bool {
		t, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		if !t.Name.IsExported() && p.prefix != "" {
			return true
		}

		if alias := checkIsAlias(t.Type); alias != "" {
			p.structs = append(p.structs, Struct{
				Pkg:    pkgName,
				Name:   t.Name.Name,
				Origin: alias,
			})
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

		p.structs = append(p.structs, Struct{
			Pkg:    pkgName,
			Name:   t.Name.Name,
			Fields: fields,
		})

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
