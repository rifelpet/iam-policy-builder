package parser

import (
	"go/ast"
)

func parseAST(f *ast.File) error {
	ast.Inspect(f, func(n ast.Node) bool {
		/*var s string
		switch x := n.(type) {
		case *ast.BasicLit:
			s = x.Value
		case *ast.Ident:
			s = x.Name
		}*/
		/*if s != "" {
			fmt.Printf("%s:\t%s\n", fset.Position(n.Pos()), s)
		}*/
		return true
	})
	return nil
}
