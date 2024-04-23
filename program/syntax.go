package program

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
)

func check(fset *token.FileSet, file *ast.File) error {
	conf := types.Config{Importer: nil}
	_, err := conf.Check("", fset, []*ast.File{file}, nil)
	if err != nil {
		return err
	}
	return nil
}

func formatError(err error, n int) error {
	str := err.Error()
	idx := strings.Index(str, ":")
	s1 := str[0:idx]
	line, _ := strconv.Atoi(s1)
	str = str[idx:]
	return fmt.Errorf("%d%s", line-n, str)
}
