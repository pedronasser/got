package transform

import (
	"encoding/hex"
	"go/ast"
	"os"
	"path"

	"crypto/sha256"
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
		name := v.Name.Name
		fnHashSum := hashExtracted(GOT_DECORATORS_DIR, string(c.FileSrc()[v.Pos()-1:v.End()-1]))
		if !isExtractedModified(name, fnHashSum) {
			log("skip extracting unmodified decorator:", name)
			return nil
		}

		imports, err := getFileImportsList(c.ASTFile())
		if err != nil {
			return err
		}
		fnSrc := string(c.FileSrc()[v.Pos()-1 : v.End()-1])
		err = extractAsPlugin(name, fnSrc, GOT_DECORATORS_DIR, imports, fnHashSum)
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
		name := v.Name.Name

		fnHashSum := hashExtracted(GOT_DECORATORS_DIR, string(c.FileSrc()[v.Pos()-1:v.End()-1]))
		if !isExtractedModified(name, fnHashSum) {
			log("skipping unmodified decorator:", name)
			return nil
		}

		imports, err := getFileImportsList(c.ASTFile())
		if err != nil {
			return err
		}

		fnSrc := string(c.FileSrc()[v.Pos()-1 : v.End()-1])
		err = extractAsPlugin(name, fnSrc, GOT_METHODS_DIR, imports, fnHashSum)
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

// isExtractedModified checks if the extracted plugin is modified
// by comparing the hash of the function with the hash of the
// extracted plugin.
func isExtractedModified(name, hash string) bool {
	hashFilePath := path.Join(GOT_BUILD_DIR, GOT_EXTRACT_DIR, name, "extract.hash")
	extractHash, err := os.ReadFile(hashFilePath)
	if err != nil {
		return true
	}

	if string(extractHash) == hash {
		return false
	}

	return true
}

// hashExtracted hashes the extracted plugin directory and the
// function source code to create a unique hash for the plugin.
func hashExtracted(extractedDir, src string) string {
	h := sha256.New()
	h.Write([]byte(extractedDir))
	h.Write([]byte(src))
	return hex.EncodeToString(h.Sum(nil))
}
