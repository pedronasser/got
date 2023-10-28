//go:build !generated

package main

//#[ImportPackages(pkg/)]
import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	got "github.com/pedronasser/got/transform"
)

func main() {

}

// #[decorator]
func ImportPackages(c *got.TransformContext) error {
	const MODULE_PATH = "github.com/pedronasser/got/examples/load-packages"

	args := c.Args()
	if len(args) == 0 {
		return fmt.Errorf("missing path")
	}

	node := c.Node()
	if node == nil {
		return fmt.Errorf("missing node")
	}

	decl := node.(*ast.GenDecl)
	if decl.Tok != token.IMPORT {
		return fmt.Errorf("not an import")
	}

	_ = filepath.Walk(args[0], func(path string, info os.FileInfo, err error) error {
		if path == "." || path == args[0] {
			return nil
		}

		if err != nil {
			return err
		}

		s, err := os.Stat(path)
		if err != nil {
			return err
		}

		if s.IsDir() {
			path = strings.TrimRight(path, "/")
			decl.Specs = append(decl.Specs, &ast.ImportSpec{
				Name: &ast.Ident{
					Name: "_",
				},
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("\"%s/%s\"", MODULE_PATH, path),
				},
			})
			return nil
		}
		return nil
	})

	c.Replace(decl)

	return nil
}
