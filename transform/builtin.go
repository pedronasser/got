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

// DecoratorAttribute is a builtin attribute that extracts the function
// and saves it as a plugin in the decorators directory.
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

// MethodAttribute is a builtin attribute that extracts the function
// and saves it as a plugin in the methods directory.
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

// PlaceholderAttribute is a builtin attribute that deletes the node
// from the AST. It is used to remove the placeholder functions from
// the code before it is compiled.
func PlaceholderAttribute(fileSrc []byte, f *ast.File, c *TransformContext, args ...string) (err error) {
	c.Delete()
	return nil
}
