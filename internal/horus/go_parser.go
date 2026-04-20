package horus

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// GoParser implements Parser using Go stdlib go/ast.
// Zero external dependencies. Handles .go files only.
type GoParser struct {
	fset *token.FileSet
}

// NewGoParser creates a new Go parser.
func NewGoParser() *GoParser {
	return &GoParser{fset: token.NewFileSet()}
}

func (p *GoParser) Language() string { return "go" }

// ParseFile extracts symbols from a single Go file.
func (p *GoParser) ParseFile(path string, src []byte) ([]Symbol, error) {
	file, err := parser.ParseFile(p.fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	var symbols []Symbol

	// Package declaration.
	if file.Name != nil {
		symbols = append(symbols, Symbol{
			Name:      file.Name.Name,
			Kind:      KindPackage,
			File:      path,
			Line:      p.fset.Position(file.Name.Pos()).Line,
			Signature: "package " + file.Name.Name,
			Exported:  true,
		})
	}

	// Top-level declarations.
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			symbols = append(symbols, p.extractFunc(path, d)...)
		case *ast.GenDecl:
			symbols = append(symbols, p.extractGenDecl(path, d)...)
		}
	}

	return symbols, nil
}

// ParseDir walks a directory tree and parses all .go files.
func (p *GoParser) ParseDir(root string) (*SymbolGraph, error) {
	graph := NewGraph(root)
	totalLines := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		totalLines += bytes.Count(src, []byte{'\n'}) + 1

		relPath, _ := filepath.Rel(root, path)
		if relPath == "" {
			relPath = path
		}

		syms, err := p.ParseFile(relPath, src)
		if err != nil {
			return nil // skip unparseable files
		}

		graph.Symbols = append(graph.Symbols, syms...)
		graph.Stats.Files++
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}

	graph.Stats.TotalLines = totalLines
	graph.computeStats()
	return graph, nil
}

func (p *GoParser) extractFunc(file string, fn *ast.FuncDecl) []Symbol {
	kind := KindFunc
	parent := ""
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		kind = KindMethod
		parent = typeExprString(fn.Recv.List[0].Type)
	}

	sig := p.funcSignature(fn)
	doc := strings.TrimSpace(fn.Doc.Text())

	return []Symbol{{
		Name:      fn.Name.Name,
		Kind:      kind,
		File:      file,
		Line:      p.fset.Position(fn.Pos()).Line,
		EndLine:   p.fset.Position(fn.End()).Line,
		Signature: sig,
		Doc:       doc,
		Exported:  isExported(fn.Name.Name),
		Parent:    parent,
	}}
}

func (p *GoParser) extractGenDecl(file string, gd *ast.GenDecl) []Symbol {
	var symbols []Symbol
	doc := strings.TrimSpace(gd.Doc.Text())

	for _, spec := range gd.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			sym := Symbol{
				Name:     s.Name.Name,
				File:     file,
				Line:     p.fset.Position(s.Pos()).Line,
				EndLine:  p.fset.Position(s.End()).Line,
				Doc:      doc,
				Exported: isExported(s.Name.Name),
			}

			switch t := s.Type.(type) {
			case *ast.InterfaceType:
				sym.Kind = KindInterface
				sym.Signature = p.interfaceSignature(s.Name.Name, t)
			case *ast.StructType:
				sym.Kind = KindStruct
				sym.Signature = p.structSignature(s.Name.Name, t)
			default:
				sym.Kind = KindType
				sym.Signature = fmt.Sprintf("type %s %s", s.Name.Name, typeExprString(s.Type))
			}
			symbols = append(symbols, sym)

		case *ast.ValueSpec:
			kind := KindVar
			if gd.Tok == token.CONST {
				kind = KindConst
			}
			for _, name := range s.Names {
				symbols = append(symbols, Symbol{
					Name:     name.Name,
					Kind:     kind,
					File:     file,
					Line:     p.fset.Position(name.Pos()).Line,
					EndLine:  p.fset.Position(s.End()).Line,
					Exported: isExported(name.Name),
				})
			}
		}
	}

	return symbols
}

func (p *GoParser) funcSignature(fn *ast.FuncDecl) string {
	// Build a minimal function declaration without the body or doc comment.
	clone := *fn
	clone.Body = nil
	clone.Doc = nil
	var buf bytes.Buffer
	printer.Fprint(&buf, p.fset, &clone)
	return strings.TrimSpace(buf.String())
}

func (p *GoParser) interfaceSignature(name string, iface *ast.InterfaceType) string {
	var buf strings.Builder
	buf.WriteString("type ")
	buf.WriteString(name)
	buf.WriteString(" interface {")
	if iface.Methods != nil {
		for _, m := range iface.Methods.List {
			if len(m.Names) > 0 {
				buf.WriteString(" ")
				buf.WriteString(m.Names[0].Name)
				buf.WriteString("()")
				buf.WriteString(";")
			}
		}
	}
	buf.WriteString(" }")
	return buf.String()
}

func (p *GoParser) structSignature(name string, st *ast.StructType) string {
	var buf strings.Builder
	buf.WriteString("type ")
	buf.WriteString(name)
	buf.WriteString(" struct {")
	if st.Fields != nil {
		for _, f := range st.Fields.List {
			for _, n := range f.Names {
				buf.WriteString(" ")
				buf.WriteString(n.Name)
				buf.WriteString(" ")
				buf.WriteString(typeExprString(f.Type))
				buf.WriteString(";")
			}
		}
	}
	buf.WriteString(" }")
	return buf.String()
}

// typeExprString returns a short string representation of a type expression.
func typeExprString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeExprString(t.X)
	case *ast.SelectorExpr:
		return typeExprString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + typeExprString(t.Elt)
	case *ast.MapType:
		return "map[" + typeExprString(t.Key) + "]" + typeExprString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func(...)"
	case *ast.ChanType:
		return "chan " + typeExprString(t.Value)
	default:
		return "..."
	}
}

func isExported(name string) bool {
	if name == "" {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}
