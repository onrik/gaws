package main

import (
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
	logger := log.New(os.Stderr, "", 0)
	logger.SetOutput(os.Stderr)

	path, err := filepath.Abs(os.Args[1])
	if err != nil {
		logger.Println(err)
		return
	}

	paths, err := getPaths(path)
	if err != nil {
		logger.Println(err)
		return
	}

	doc := Doc{
		OpenAPI: "3.0.0",
		Paths:   map[string]Path{},
	}

	errors := []string{}
	for i := range paths {
		// logger.Println(paths[i])
		structs, err := parseStructs("", paths[i], true)
		if err != nil {
			logger.Printf("Parse structs error: %s\n", err)
			return
		}

		pkgs, err := parser.ParseDir(token.NewFileSet(), paths[i], nil, parser.ParseComments)
		if err != nil {
			logger.Printf("Parse dir error: %s\n", err)
			return
		}

		p := NewParser(&doc, structs)
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
			logger.Println(errors[i])
		}
		os.Exit(1)
	}

	data, err := yaml.Marshal(doc)
	if err != nil {
		logger.Println(err)
	}

	fmt.Println(string(data))
}

/*
Рекурсия в структурах
Анонимные вложенные структуры
Обработка json.RawMessage
*/

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
