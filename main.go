package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

var (
	debug = false
)

func main() {
	log.SetOutput(os.Stderr)

	var (
		version      string
		title        string
		descriptions string
		server       string
		dir          string
		skipDirs     string
		indent       int
	)

	flag.StringVar(&version, "v", "1.0.0", "Docs version")
	flag.StringVar(&title, "t", "API Docs", "Docs title")
	flag.StringVar(&descriptions, "d", "OpenAPI", "Docs description")
	flag.StringVar(&server, "s", "https://localhost:8000", "API server url")
	flag.StringVar(&dir, "path", "", "Path with go files")
	flag.StringVar(&skipDirs, "skip", "", "paths to skipping")
	flag.IntVar(&indent, "indent", 2, "Yaml indentation")
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.Parse()

	path, err := filepath.Abs(dir)
	if err != nil {
		log.Println(err)
		return
	}

	paths, err := getPaths(path, skipDirs)
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
		Components: Component{Schemas: map[string]*Schema{}},
	}

	errors := []string{}
	for i := range paths {
		pkgs, err := parser.ParseDir(token.NewFileSet(), paths[i], nil, parser.ParseComments)
		if err != nil {
			log.Printf("Parse dir error: %s\n", err)
			return
		}

		p := NewParser(&doc, newStructsParser())
		for _, pkg := range pkgs {
			for filePath, f := range pkg.Files {
				for _, c := range f.Comments {
					err = p.parseComment(c.Text(), NewFile(f, filepath.Dir(filePath), ""))
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

	encoder := yaml.NewEncoder(os.Stdout, yaml.Indent(indent))
	err = encoder.Encode(doc)
	if err != nil {
		log.Println(err)
	}
}

func getPaths(dir, skip string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	paths := []string{dir}
	for i := range files {
		if !files[i].IsDir() {
			continue
		}

		p := filepath.Join(dir, files[i].Name())
		if skip != "" && strings.Contains(skip, p) {
			continue
		}
		pp, err := getPaths(p, skip)
		if err != nil {
			return nil, err
		}

		paths = append(paths, pp...)
	}

	return paths, nil
}
