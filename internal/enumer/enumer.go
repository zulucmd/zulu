// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	exact "go/constant"
	gofmt "go/format"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/zulucmd/zulu/v2/internal/template"
	"golang.org/x/tools/go/packages"
)

var (
	typeName     = flag.String("type", "", "comma-separated list of type names; must be set")
	output       = flag.String("output", "", "output file name; default srcdir/<type>_string.go")
	templateFile = flag.String("template", "", "template file to use")
	format       = flag.Bool("format", false, "format the template, only for code generation")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	_, _ = fmt.Fprintf(os.Stderr, "Enumer is a tool to generate files based on Go enums (constants with a specific type).\n")
	_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	_, _ = fmt.Fprintf(os.Stderr, "\tEnumer [flags] -type T [directory]\n")
	_, _ = fmt.Fprintf(os.Stderr, "\tEnumer [flags] -type T files... # Must be a single package\n")
	_, _ = fmt.Fprintf(os.Stderr, "For more information, see:\n")
	_, _ = fmt.Fprintf(os.Stderr, "\thttps://godoc.org/github.com/zulucmd/zulu/v2/internal/enumer\n")
	_, _ = fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("enumer: ")
	flag.Usage = Usage
	flag.Parse()
	if len(*typeName) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in the current directory.
		args = []string{"."}
	}

	// Parse the package once.
	var g Generator

	dir := getDir(args[0])
	path, err := filepath.Rel(dir, *templateFile)
	if err != nil {
		panic(err)
	}

	g.parsePackage(args)

	res, err := template.ParseFromFile(os.DirFS(dir), path, map[string]interface{}{
		"pkgName":  g.pkg.name,
		"args":     strings.Join(os.Args[1:], " "),
		"typeName": *typeName,
		"values":   g.getValues(*typeName),
	}, nil)
	if err != nil {
		panic(err)
	}

	g.Print(res)

	src := g.format(*format)
	writeSource(*typeName, dir, *output, src)
}

func writeSource(typeName, dir, outputName string, src []byte) {
	if outputName == "-" {
		_, err := os.Stdout.Write(src)
		if err != nil {
			log.Fatalf("failed to write output: %s", err)
		}
		return
	}

	if outputName == "" {
		baseName := fmt.Sprintf("%s.gen.go", typeName)
		outputName = filepath.Join(dir, strings.ToLower(baseName))
	}

	// Write to tmpfile first
	tmpFile, err := os.CreateTemp(dir, fmt.Sprintf("%s_enumer_", typeName))
	if err != nil {
		log.Fatalf("creating temporary file for output: %s", err)
	}
	_, err = tmpFile.Write(src)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		log.Fatalf("failed to write output: %s", err)
	}
	tmpFile.Close()

	// Rename tmpfile to output file
	err = os.Rename(tmpFile.Name(), outputName)
	if err != nil {
		log.Fatalf("failed to move tempfile to output file: %s", err)
	}
}

func getDir(fileOrDir string) string {
	info, err := os.Stat(fileOrDir)
	if err != nil {
		log.Fatal(err)
	}

	if info.IsDir() {
		return fileOrDir
	}

	return filepath.Dir(fileOrDir)
}

// Generator holds the state of the analysis. Primarily used to buffer
// the output for gofmt.Source.
type Generator struct {
	buf bytes.Buffer // Accumulated output.
	pkg *Package     // Package we are scanning.
}

// Printf prints the string to the output
func (g *Generator) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(&g.buf, format, args...)
}

// Print prints the string to the output
func (g *Generator) Print(str string) {
	_, _ = fmt.Fprint(&g.buf, str)
}

// File holds a single parsed file and associated data.
type File struct {
	pkg  *Package  // Package to which this file belongs.
	file *ast.File // Parsed AST.
	// These fields are reset for each type being generated.
	typeName string  // Name of the constant type.
	values   []Value // Accumulator for constant values of that type.
}

// Package holds information about a Go package
type Package struct {
	name  string
	defs  map[*ast.Ident]types.Object
	files []*File
}

// parsePackage analyzes the single package constructed from the patterns and tags.
// parsePackage exits if there is an error.
func (g *Generator) parsePackage(patterns []string) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes |
			packages.NeedSyntax | packages.NeedTypesInfo,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	g.addPackage(pkgs[0])
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &File{
			file: file,
			pkg:  g.pkg,
		}
	}
}

// getValues produces the String method for the named type.
func (g *Generator) getValues(typeName string) []Value {
	values := make([]Value, 0, 100)
	for _, file := range g.pkg.files {
		file.typeName = typeName
		file.values = nil
		if file.file != nil {
			ast.Inspect(file.file, file.genDecl)
			values = append(values, file.values...)
		}
	}

	if len(values) == 0 {
		log.Fatalf("no values defined for type %s", typeName)
	}

	return values
}

// format returns the gofmt-ed contents of the Generator's buffer.
func (g *Generator) format(format bool) []byte {
	src := g.buf.Bytes()
	if !format {
		return src
	}

	var err error
	src, err = gofmt.Source(src)
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}

// Value represents a declared constant.
type Value struct {
	Name string // The name of the constant before transformation
	// The value is stored as a bit pattern alone. The boolean tells us
	// whether to interpret it as an int64 or a uint64; the only place
	// this matters is when sorting.
	// Much of the time the Value field is all we need; it is printed
	// by Value.String.
	Value    string // The string representation given by the "go/exact" package.
	Comment  string // The comment given to this field.
	Exported bool   // Whether the field is exported.
}

func (v *Value) String() string {
	return v.Value
}

// genDecl processes one declaration clause.
func (f *File) genDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.CONST {
		// We only care about const declarations.
		return true
	}
	// The name of the type of the constants we are declaring.
	// Can change if this is a multi-element declaration.
	typ := ""
	// Loop over the elements of the declaration. Each element is a ValueSpec:
	// a list of names possibly followed by a type, possibly followed by values.
	// If the type and value are both missing, we carry down the type (and value,
	// but the "go/types" package takes care of that).
	for _, spec := range decl.Specs {
		vspec := spec.(*ast.ValueSpec) // Guaranteed to succeed as this is CONST.
		if vspec.Type == nil && len(vspec.Values) > 0 {
			// "X = 1". With no type but a value, the constant is untyped.
			// Skip this vspec and reset the remembered type.
			typ = ""
			continue
		}
		if vspec.Type != nil {
			// "X T". We have a type. Remember it.
			ident, ok := vspec.Type.(*ast.Ident)
			if !ok {
				continue
			}
			typ = ident.Name
		}
		if typ != f.typeName {
			// This is not the type we're looking for.
			continue
		}
		// We now have a list of names (from one line of source code) all being
		// declared with the desired type.
		// Grab their names and actual values and store them in f.values.
		for _, n := range vspec.Names {
			if n.Name == "_" {
				continue
			}

			// This dance lets the type checker find the values for us. It's a
			// bit tricky: look up the object declared by the n, find its
			// types.Const, and extract its value.
			obj, ok := f.pkg.defs[n]
			if !ok {
				log.Fatalf("no value for constant %s", n)
			}

			info := obj.Type().Underlying().(*types.Basic).Info()
			if info&types.IsInteger == 0 {
				log.Fatalf("can't handle non-integer constant type %s", typ)
			}

			value := obj.(*types.Const).Val() // Guaranteed to succeed as this is CONST.
			if value.Kind() != exact.Int {
				log.Fatalf("can't happen: constant is not an integer %s", n)
			}

			v := Value{
				Name:     n.Name,
				Value:    value.String(),
				Exported: n.IsExported(),
			}

			if vspec.Comment != nil || vspec.Doc != nil {
				var comment *ast.CommentGroup
				switch {
				case vspec.Comment == nil && vspec.Doc != nil:
					comment = vspec.Doc
				case vspec.Comment != nil && vspec.Doc == nil:
					comment = vspec.Comment
				default:
					log.Fatalf("cannot work with both doc comment and normal comment: %s", n.Name)
				}

				v.Comment = getComment(comment.List)
			}

			f.values = append(f.values, v)
		}
	}
	return false
}

func getComment(commentList []*ast.Comment) string {
	var comment []byte
	for _, c := range commentList {
		comment = append(comment, c.Text...)
		comment = append(comment, '\n')
	}
	return string(comment)
}
