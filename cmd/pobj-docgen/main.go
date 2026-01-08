// pobj-docgen extracts godoc comments and generates SetDoc calls for pobj registrations.
//
// Usage:
//
//	//go:generate go run github.com/KarpelesLab/pobj/cmd/pobj-docgen
//
// The tool scans the current package for pobj.Register, pobj.RegisterActions, and
// pobj.RegisterMethod calls, finds the associated godoc comments for the registered
// types and functions, and generates a pobj_doc.go file with init() that sets
// the documentation.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	var (
		outputFile = flag.String("o", "pobj_doc.go", "output file name")
		pkgDir     = flag.String("dir", ".", "package directory to process")
	)
	flag.Parse()

	if err := run(*pkgDir, *outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "pobj-docgen: %v\n", err)
		os.Exit(1)
	}
}

func run(pkgDir, outputFile string) error {
	fset := token.NewFileSet()

	// Parse all Go files in the directory
	pkgs, err := parser.ParseDir(fset, pkgDir, func(fi os.FileInfo) bool {
		// Skip test files and generated doc file
		name := fi.Name()
		return !strings.HasSuffix(name, "_test.go") && name != outputFile
	}, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing package: %w", err)
	}

	if len(pkgs) == 0 {
		return fmt.Errorf("no packages found in %s", pkgDir)
	}

	// Process each package (usually just one)
	for pkgName, pkg := range pkgs {
		docs, err := extractDocs(pkg)
		if err != nil {
			return err
		}

		if len(docs.types) == 0 && len(docs.methods) == 0 {
			fmt.Printf("pobj-docgen: no pobj registrations found in package %s\n", pkgName)
			continue
		}

		// Generate output
		output, err := generateOutput(pkgName, docs)
		if err != nil {
			return fmt.Errorf("generating output: %w", err)
		}

		outPath := filepath.Join(pkgDir, outputFile)
		if err := os.WriteFile(outPath, output, 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}

		// Count total field docs
		fieldCount := 0
		for _, td := range docs.types {
			fieldCount += len(td.fields)
		}

		fmt.Printf("pobj-docgen: generated %s with %d type docs, %d field docs, and %d method docs\n",
			outPath, len(docs.types), fieldCount, len(docs.methods))
	}

	return nil
}

type docInfo struct {
	types   map[string]typeDoc   // registration path -> doc info
	methods map[string]methodDoc // "Object:method" -> doc info
}

type typeDoc struct {
	path   string            // registration path
	doc    string            // documentation
	fields map[string]string // field name -> field documentation
}

type methodDoc struct {
	path string // full "Object:method" path
	doc  string // documentation
}

// typeInfo holds documentation for a type and its fields
type typeInfo struct {
	doc    string
	fields map[string]string // field name -> documentation
}

func extractDocs(pkg *ast.Package) (*docInfo, error) {
	info := &docInfo{
		types:   make(map[string]typeDoc),
		methods: make(map[string]methodDoc),
	}

	// First pass: build maps of type and function documentation
	typeInfos := make(map[string]*typeInfo) // type name -> info
	funcDocs := make(map[string]string)     // func name -> doc
	varFuncs := make(map[string]string)     // var name -> func name (for var x = funcName patterns)

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						ti := &typeInfo{
							fields: make(map[string]string),
						}

						// Get type-level documentation
						if s.Doc != nil {
							ti.doc = strings.TrimSpace(s.Doc.Text())
						} else if d.Doc != nil {
							ti.doc = strings.TrimSpace(d.Doc.Text())
						}

						// Extract struct field documentation
						if structType, ok := s.Type.(*ast.StructType); ok {
							for _, field := range structType.Fields.List {
								fieldDoc := ""
								if field.Doc != nil {
									fieldDoc = strings.TrimSpace(field.Doc.Text())
								} else if field.Comment != nil {
									// Inline comment like `field int // comment`
									fieldDoc = strings.TrimSpace(field.Comment.Text())
								}

								if fieldDoc != "" {
									// A field can have multiple names (e.g., `a, b int`)
									for _, name := range field.Names {
										ti.fields[name.Name] = fieldDoc
									}
								}
							}
						}

						if ti.doc != "" || len(ti.fields) > 0 {
							typeInfos[s.Name.Name] = ti
						}
					case *ast.ValueSpec:
						// Track variable assignments to functions
						if len(s.Names) == 1 && len(s.Values) == 1 {
							if ident, ok := s.Values[0].(*ast.Ident); ok {
								varFuncs[s.Names[0].Name] = ident.Name
							}
						}
					}
				}
			case *ast.FuncDecl:
				if d.Doc != nil {
					funcDocs[d.Name.Name] = strings.TrimSpace(d.Doc.Text())
				}
			}
		}
	}

	// Second pass: find pobj registration calls
	for _, file := range pkg.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check for pobj.Register, pobj.RegisterActions, pobj.RegisterMethod
			var funcName string
			var pkgIdent string

			switch fn := call.Fun.(type) {
			case *ast.IndexExpr:
				// Generic call like pobj.Register[Type](...)
				if sel, ok := fn.X.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						pkgIdent = ident.Name
						funcName = sel.Sel.Name
					}
				}
			case *ast.SelectorExpr:
				// Non-generic call like pobj.RegisterMethod(...)
				if ident, ok := fn.X.(*ast.Ident); ok {
					pkgIdent = ident.Name
					funcName = fn.Sel.Name
				}
			}

			if pkgIdent != "pobj" {
				return true
			}

			switch funcName {
			case "Register", "RegisterActions":
				info.processRegister(call, typeInfos)
			case "RegisterMethod", "RegisterStatic":
				info.processRegisterMethod(call, funcDocs, varFuncs)
			}

			return true
		})
	}

	return info, nil
}

// processRegister handles pobj.Register[Type]("path") and pobj.RegisterActions[Type]("path", ...)
func (info *docInfo) processRegister(call *ast.CallExpr, typeInfos map[string]*typeInfo) {
	if len(call.Args) < 1 {
		return
	}

	// Get the registration path from first argument
	path := extractStringLit(call.Args[0])
	if path == "" {
		return
	}

	// Get the type parameter
	indexExpr, ok := call.Fun.(*ast.IndexExpr)
	if !ok {
		return
	}

	typeName := extractTypeName(indexExpr.Index)
	if typeName == "" {
		return
	}

	if ti, ok := typeInfos[typeName]; ok {
		info.types[path] = typeDoc{
			path:   path,
			doc:    ti.doc,
			fields: ti.fields,
		}
	}
}

// processRegisterMethod handles pobj.RegisterMethod("Object:method", funcName)
func (info *docInfo) processRegisterMethod(call *ast.CallExpr, funcDocs map[string]string, varFuncs map[string]string) {
	if len(call.Args) < 2 {
		return
	}

	// Get the method path from first argument
	path := extractStringLit(call.Args[0])
	if path == "" || !strings.Contains(path, ":") {
		return
	}

	// Get the function name from second argument
	funcName := ""
	switch arg := call.Args[1].(type) {
	case *ast.Ident:
		funcName = arg.Name
	case *ast.SelectorExpr:
		// pkg.FuncName
		funcName = arg.Sel.Name
	}

	if funcName == "" {
		return
	}

	// Check if it's a variable pointing to a function
	if actualFunc, ok := varFuncs[funcName]; ok {
		funcName = actualFunc
	}

	if doc, ok := funcDocs[funcName]; ok {
		info.methods[path] = methodDoc{
			path: path,
			doc:  doc,
		}
	}
}

func extractStringLit(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	// Remove quotes
	s := lit.Value
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
		if s[0] == '`' && s[len(s)-1] == '`' {
			return s[1 : len(s)-1]
		}
	}
	return ""
}

func extractTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		// *Type -> Type
		return extractTypeName(e.X)
	case *ast.SelectorExpr:
		// pkg.Type -> Type
		return e.Sel.Name
	}
	return ""
}

func generateOutput(pkgName string, docs *docInfo) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("// Code generated by pobj-docgen. DO NOT EDIT.\n\n")
	buf.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	buf.WriteString("import \"github.com/KarpelesLab/pobj\"\n\n")
	buf.WriteString("func init() {\n")

	// Generate SetDoc calls for types and their fields
	for _, td := range docs.types {
		if td.doc != "" {
			buf.WriteString(fmt.Sprintf("\tpobj.Get(%q).SetDoc(%s)\n", td.path, formatDoc(td.doc)))
		}
		// Generate SetFieldDoc calls for each field
		for fieldName, fieldDoc := range td.fields {
			buf.WriteString(fmt.Sprintf("\tpobj.Get(%q).SetFieldDoc(%q, %s)\n", td.path, fieldName, formatDoc(fieldDoc)))
		}
	}

	// Generate SetDoc calls for methods
	for _, md := range docs.methods {
		parts := strings.SplitN(md.path, ":", 2)
		if len(parts) != 2 {
			continue
		}
		objPath, methodName := parts[0], parts[1]
		buf.WriteString(fmt.Sprintf("\tpobj.Get(%q).Method(%q).SetDoc(%s)\n", objPath, methodName, formatDoc(md.doc)))
	}

	buf.WriteString("}\n")

	// Format the output
	return format.Source(buf.Bytes())
}

var needsRawString = regexp.MustCompile("[`]")

func formatDoc(doc string) string {
	// Use raw string literal if no backticks, otherwise use quoted string
	if !needsRawString.MatchString(doc) {
		return "`" + doc + "`"
	}
	// Fall back to quoted string with escaping
	return fmt.Sprintf("%q", doc)
}
