package rules

import (
	"errors"
	"go/ast"
	"go/token"
)

// RuleUndocumented names an exported declaration with no doc comment, or a
// package with no package comment (NFR-20).
const RuleUndocumented = "undocumented"

// checkDocs applies the documentation rule to one non-test file. Unlike the
// goroutine, clock and output rules it spares no package: the test kit and
// these rules are documented like any other.
func (c *checker) checkDocs(file *ast.File, rel string) error {
	if file == nil {
		return errors.New("rules: checkDocs needs a parsed file")
	}
	if rel == "" {
		return errors.New("rules: checkDocs needs the file's name")
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Doc == nil && d.Name.IsExported() && receiverExported(d) {
				c.found = append(c.found, Finding{rel, c.fset.Position(d.Pos()).Line, RuleUndocumented, "exported " + d.Name.Name + " has no doc comment"})
			}
		case *ast.GenDecl:
			if d.Tok == token.IMPORT || d.Doc != nil {
				continue // one comment may serve a whole group
			}
			for _, name := range undocumentedSpecs(d) {
				c.found = append(c.found, Finding{rel, c.fset.Position(name.Pos()).Line, RuleUndocumented, "exported " + name.Name + " has no doc comment"})
			}
		}
	}
	return nil
}

// receiverExported reports whether a function is part of the package's
// exported surface: a plain function, or a method on an exported type.
func receiverExported(d *ast.FuncDecl) bool {
	if d.Recv == nil || len(d.Recv.List) == 0 {
		return true
	}
	t := d.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if idx, ok := t.(*ast.IndexExpr); ok {
		t = idx.X
	}
	id, ok := t.(*ast.Ident)
	return ok && id.IsExported()
}

// undocumentedSpecs returns the exported names of a type, constant or
// variable declaration that have neither a doc comment nor a trailing one.
func undocumentedSpecs(d *ast.GenDecl) []*ast.Ident {
	var bare []*ast.Ident
	for _, spec := range d.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Doc == nil && s.Comment == nil && s.Name.IsExported() {
				bare = append(bare, s.Name)
			}
		case *ast.ValueSpec:
			if s.Doc != nil || s.Comment != nil {
				continue
			}
			for _, name := range s.Names {
				if name.IsExported() {
					bare = append(bare, name)
				}
			}
		}
	}
	return bare
}
