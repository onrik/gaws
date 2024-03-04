package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

type Package struct {
	FSPath     string
	ImportPath string
}

// File contains details about parsed file and its package
type File struct {
	ParsedFile *ast.File
	Pkg        Package
}

func NewFile(file *ast.File, fsPath, importPath string) File {
	return File{
		ParsedFile: file,
		Pkg: Package{
			FSPath:     fsPath,
			ImportPath: importPath,
		},
	}
}

// ParseImport search import in current File and returns parsed File with import's sources
func (f *File) ParseImport(pkgName, typeName string) (File, error) {
	importedPkgPath, found := f.getImportPathForPkg(pkgName, f.ParsedFile)
	if !found {
		return File{}, fmt.Errorf("not found import path for package: %s", pkgName)
	}

	importedPkgFSPath, err := f.resolvePkgFSPath(importedPkgPath, f.Pkg.FSPath)
	if err != nil {
		return File{}, err
	}

	fileWithImportedType, err := f.findSourceFileWithTypeDef(typeName, importedPkgFSPath)
	if err != nil {
		return File{}, err
	}

	return NewFile(fileWithImportedType, importedPkgFSPath, importedPkgPath), nil
}

func (f *File) getImportPathName(fileImport *ast.ImportSpec) string {
	if fileImport.Name != nil {
		return fileImport.Name.Name
	} else {
		fileImportChunks := strings.Split(strings.Trim(fileImport.Path.Value, `"`), "/")
		return fileImportChunks[len(fileImportChunks)-1]
	}
}

func (f *File) getImportPathForPkg(pkg string, file *ast.File) (string, bool) {
	for i := range file.Imports {
		fileImportName := f.getImportPathName(file.Imports[i])
		if pkg == fileImportName {
			return strings.Trim(file.Imports[i].Path.Value, `"`), true
		}
	}

	return "", false
}

func (f *File) resolvePkgFSPath(importPath string, currentPackagePath string) (string, error) {
	config := &packages.Config{
		Mode: packages.LoadImports, // nolint: staticcheck
		Dir:  currentPackagePath,
	}

	pkgList, err := packages.Load(config)
	if err != nil {
		return "", err
	}

	for i := range pkgList[0].Imports {
		pkg := pkgList[0].Imports[i]
		if pkg.PkgPath == importPath {
			fsPath, _ := filepath.Split(pkg.GoFiles[0])
			return fsPath, nil
		}
	}

	return "", fmt.Errorf("file system path for '%s' not found in package '%s'", importPath, currentPackagePath)
}

func (f *File) findSourceFileWithTypeDef(typeName, pkgImportFSPath string) (*ast.File, error) {
	pkgs, err := parser.ParseDir(token.NewFileSet(), pkgImportFSPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// TODO: filter package by name
	for _, importedPkg := range pkgs {
		for i := range importedPkg.Files {
			found := false
			ast.Inspect(importedPkg.Files[i], func(node ast.Node) bool {
				t, ok := node.(*ast.TypeSpec)
				if !ok {
					return true
				}

				// skip unexported structs from underlying packages
				if !t.Name.IsExported() {
					return true
				}

				if t.Name.Name == typeName {
					found = true
					return false
				}

				return true
			})

			if found {
				return importedPkg.Files[i], nil
			}
		}
	}

	return nil, fmt.Errorf("'%s' type definition not found in package '%s'", typeName, pkgImportFSPath)
}
