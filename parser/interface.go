package parser

import (
	"bytes"
	"errors"
	"go/ast"
	"go/token"
	"io"
	"os"
)

func readSource(filename string, src any) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			return io.ReadAll(s)
		}
		return nil, errors.New("invalid source")
	}
	return os.ReadFile(filename)
}

// A Mode value is a set of flags (or 0).
// They control the amount of source code parsed and other optional
// parser functionality.
type Mode uint

const (
	PackageClauseOnly    Mode             = 1 << iota // stop parsing after package clause
	ImportsOnly                                       // stop parsing after import declarations
	ParseComments                                     // parse comments and add them to AST
	Trace                                             // print a trace of parsed productions
	DeclarationErrors                                 // report declaration errors
	SpuriousErrors                                    // same as AllErrors, for backward-compatibility
	SkipObjectResolution                              // don't resolve identifiers to objects - see ParseFile
	AllErrors            = SpuriousErrors             // report all errors (not just the first 10 on different lines)
)

func ParseFile(fset *token.FileSet, filename string, src any, mode Mode) (astFile *ast.File, tokenFile *token.File, err error) {
	text, err := readSource(filename, src)
	if err != nil {
		return nil, nil, err
	}

	var p parser
	defer func() {
		if e := recover(); e != nil {
			// resume same panic if it's not a bailout
			bail, ok := e.(bailout)
			if !ok {
				panic(e)
			} else if bail.msg != "" {
				p.errors.Add(p.file.Position(bail.pos), bail.msg)
			}
		}

		// set result values
		if astFile == nil {
			// source is not a valid Go source file - satisfy
			// ParseFile API and return a valid (but) empty
			// *ast.File
			astFile = &ast.File{
				Name:  new(ast.Ident),
				Scope: ast.NewScope(nil),
			}
		}

		p.errors.Sort()
		err = p.errors.Err()
	}()

	// parse source
	p.init(fset, filename, text, mode)
	astFile = p.parseFile()
	tokenFile = p.file

	return
}
