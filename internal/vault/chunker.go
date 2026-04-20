package vault

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Chunker splits source code into semantically meaningful chunks for indexing.
type Chunker interface {
	Chunk(path string, content []byte) ([]CodeChunk, error)
}

// GoChunker splits Go files at function/type boundaries using go/ast.
type GoChunker struct{}

// Chunk parses a Go file and extracts one chunk per top-level declaration.
func (c *GoChunker) Chunk(path string, content []byte) ([]CodeChunk, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		// Fall back to generic chunker on parse error.
		gc := &GenericChunker{MaxChunkLines: 50, Overlap: 25}
		return gc.Chunk(path, content)
	}

	lines := strings.Split(string(content), "\n")
	var chunks []CodeChunk

	for _, decl := range file.Decls {
		start := fset.Position(decl.Pos()).Line
		end := fset.Position(decl.End()).Line
		if start < 1 {
			start = 1
		}
		if end > len(lines) {
			end = len(lines)
		}

		chunk := CodeChunk{
			File:      path,
			StartLine: start,
			EndLine:   end,
			Content:   strings.Join(lines[start-1:end], "\n"),
		}

		switch d := decl.(type) {
		case *ast.FuncDecl:
			chunk.Kind = "function"
			chunk.Name = d.Name.Name
			if d.Recv != nil && len(d.Recv.List) > 0 {
				chunk.Kind = "method"
			}
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				chunk.Kind = "type"
				if len(d.Specs) > 0 {
					if ts, ok := d.Specs[0].(*ast.TypeSpec); ok {
						chunk.Name = ts.Name.Name
					}
				}
			case token.CONST:
				chunk.Kind = "const"
			case token.VAR:
				chunk.Kind = "var"
			case token.IMPORT:
				chunk.Kind = "import"
			}
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// GenericChunker splits files using a line-based sliding window.
type GenericChunker struct {
	MaxChunkLines int // Maximum lines per chunk (default 50).
	Overlap       int // Lines of overlap between chunks (default 25).
}

// Chunk splits content into fixed-size overlapping windows.
func (c *GenericChunker) Chunk(path string, content []byte) ([]CodeChunk, error) {
	maxLines := c.MaxChunkLines
	if maxLines <= 0 {
		maxLines = 50
	}
	overlap := c.Overlap
	if overlap <= 0 {
		overlap = 25
	}
	if overlap >= maxLines {
		overlap = maxLines / 2
	}

	lines := strings.Split(string(content), "\n")
	var chunks []CodeChunk

	step := maxLines - overlap
	for i := 0; i < len(lines); i += step {
		end := i + maxLines
		if end > len(lines) {
			end = len(lines)
		}

		chunk := CodeChunk{
			File:      path,
			StartLine: i + 1,
			EndLine:   end,
			Kind:      "block",
			Content:   strings.Join(lines[i:end], "\n"),
		}
		chunks = append(chunks, chunk)

		if end >= len(lines) {
			break
		}
	}

	return chunks, nil
}
