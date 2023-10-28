package transform

import (
	"go/ast"
)

type BuiltinAttributeFn = func(c *TransformContext) error

var BuiltinAttributes = map[string]BuiltinAttributeFn{
	"method":    MethodAttribute,
	"decorator": DecoratorAttribute,
}

var exportedMethods = []string{}
var exportedDecorators = []string{}

func DecoratorAttribute(c *TransformContext) error {
	target := c.Node()

	if v, ok := target.(*ast.FuncDecl); ok {
		imports, err := getFileImportsList(c.ASTFile())
		if err != nil {
			return err
		}
		name := v.Name.Name
		fnSrc := string(c.FileSrc()[v.Pos()-1 : v.End()-1])
		err = extractAsPlugin(name, fnSrc, GOT_DECORATORS_DIR, imports)
		if err != nil {
			log(err)
			return err
		}

		exportedDecorators = append(exportedDecorators, name)
	}

	return nil
}

func MethodAttribute(c *TransformContext) error {
	target := c.Node()

	if v, ok := target.(*ast.FuncDecl); ok {
		imports, err := getFileImportsList(c.ASTFile())
		if err != nil {
			return err
		}
		name := v.Name.Name
		fnSrc := string(c.FileSrc()[v.Pos()-1 : v.End()-1])
		err = extractAsPlugin(name, fnSrc, GOT_METHODS_DIR, imports)
		if err != nil {
			return err
		}

		exportedMethods = append(exportedMethods, name)
	}

	return nil
}

func PlaceholderAttribute(fileSrc []byte, f *ast.File, c *TransformContext, args ...string) (err error) {
	c.Delete()
	return nil
}
