package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	log.SetOutput(os.Stderr)

	var (
		version      string
		title        string
		descriptions string
		server       string
		dir          string
		indent       int
	)

	flag.StringVar(&version, "v", "1.0.0", "Docs version")
	flag.StringVar(&title, "t", "API Docs", "Docs title")
	flag.StringVar(&descriptions, "d", "OpenAPI", "Docs description")
	flag.StringVar(&server, "s", "https://localhost:8000", "API server url")
	flag.StringVar(&dir, "path", "", "Path with go files")
	flag.IntVar(&indent, "indent", 2, "Yaml indentation")
	flag.Parse()

	path, err := filepath.Abs(dir)
	if err != nil {
		log.Println(err)
		return
	}

	paths, err := getPaths(path)
	if err != nil {
		log.Println(err)
		return
	}

	doc := Doc{
		OpenAPI: "3.0.0",
		Info: InfoProps{
			Description: descriptions,
			Title:       title,
			Version:     version,
		},
		Servers:    []Server{{URL: server}},
		Paths:      map[string]Path{},
		Components: Component{Schemas: map[string]Schema{}},
	}

	errors := []string{}
	for i := range paths {
		ps, err := parseStructs("", paths[i], true)
		if err != nil {
			log.Printf("Parse structs error: %s\n", err)
			return
		}

		pkgs, err := parser.ParseDir(token.NewFileSet(), paths[i], nil, parser.ParseComments)
		if err != nil {
			log.Printf("Parse dir error: %s\n", err)
			return
		}

		p := NewParser(&doc, ps.structs, ps.origins)
		for _, pkg := range pkgs {
			for filePath, f := range pkg.Files {
				for _, c := range f.Comments {
					err = p.parseComment(c.Text())
					if err != nil {
						errors = append(errors, fmt.Sprintf("%s - %s", err.Error(), strings.TrimPrefix(filePath, path+"/")))
					}
				}
			}
		}
	}

	if len(errors) > 0 {
		for i := range errors {
			log.Println(errors[i])
		}
		os.Exit(1)
	}

	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(indent)
	err = encoder.Encode(doc)
	if err != nil {
		log.Println(err)
	}
}

func getPaths(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	paths := []string{dir}
	for i := range files {
		if files[i].IsDir() {
			p := filepath.Join(dir, files[i].Name())
			pp, err := getPaths(p)
			if err != nil {
				return nil, err
			}

			paths = append(paths, pp...)
		}
	}

	return paths, nil
}
