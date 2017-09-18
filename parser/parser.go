package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

const commentPrefix = "easyjson:"

type Parser struct {
	PkgPath    string
	PkgName    string
	Types      []TypeInfo
	AllStructs bool
}

type TypeInfo struct {
	Name string
	Tags []string
}

type visitor struct {
	*Parser

	name string
	tags []string
}

func (p *Parser) easyTags(comments string) []string {
	for _, v := range strings.Split(comments, "\n") {
		if strings.HasPrefix(v, commentPrefix) {
			v = strings.TrimPrefix(v, commentPrefix)
			return strings.Split(v, ",")
		}
	}
	return nil
}

func (v *visitor) Visit(n ast.Node) (w ast.Visitor) {
	switch n := n.(type) {
	case *ast.Package:
		return v
	case *ast.File:
		v.PkgName = n.Name.String()
		return v

	case *ast.GenDecl:
		v.tags = v.easyTags(n.Doc.Text())

		if v.tags == nil && !v.AllStructs {
			return nil
		}
		return v
	case *ast.TypeSpec:
		v.name = n.Name.String()

		// Allow to specify non-structs explicitly independent of '-all' flag.
		if v.tags != nil {
			v.Types = append(v.Types, TypeInfo{Name: v.name, Tags: v.tags})
			return nil
		}
		return v
	case *ast.StructType:
		v.Types = append(v.Types, TypeInfo{Name: v.name, Tags: v.tags})
		return nil
	}
	return nil
}

func (p *Parser) Parse(fname string, isDir bool) error {
	var err error
	if p.PkgPath, err = getPkgPath(fname, isDir); err != nil {
		return err
	}

	fset := token.NewFileSet()
	if isDir {
		packages, err := parser.ParseDir(fset, fname, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, pckg := range packages {
			ast.Walk(&visitor{Parser: p}, pckg)
		}
	} else {
		f, err := parser.ParseFile(fset, fname, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		ast.Walk(&visitor{Parser: p}, f)
	}
	return nil
}
