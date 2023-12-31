package transform

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

// extractAsPlugin extracts the function and saves it as a plugin in the
// specified directory.
// First it creates a new directory for the extracted function.
// Then it creates a new file in the directory with the extracted function.
// Then it executes goimports on the file.
// Then it builds the file as a plugin.
func extractAsPlugin(name, src, extractDir string, imports []*ast.ImportSpec, hashSum string) error {
	extractedSrc := "package main\n\n"
	if len(imports) > 0 {
		extractedSrc += "import (\n"
		for _, imp := range imports {
			if imp.Name == nil {
				extractedSrc += fmt.Sprintf("\t%s\n", imp.Path.Value)
			} else {
				extractedSrc += fmt.Sprintf("\t%s %s\n", imp.Name, imp.Path.Value)
			}
		}
		extractedSrc += ")\n"
	}
	extractedSrc += string(src)
	extractedSrcDir := filepath.Join(GOT_BUILD_DIR, GOT_EXTRACT_DIR, name)
	extractedSrcPath := filepath.Join(extractedSrcDir, "extract.go")

	methodBinPath := filepath.Join(GOT_BUILD_DIR, extractDir, fmt.Sprintf("%s.so", name))

	if err := os.MkdirAll(extractedSrcDir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(
		extractedSrcPath,
		[]byte(extractedSrc),
		0644,
	); err != nil {
		return err
	}

	err := executeGoImports(extractedSrcPath)
	if err != nil {
		return fmt.Errorf("Failed to execute goimports: %s", err)
	}

	err = buildAsPlugin(extractedSrcPath, methodBinPath)
	if err != nil {
		return fmt.Errorf("Failed to build plugin: %s", err)
	}

	hashFile := filepath.Join(extractedSrcDir, "extract.hash")
	_ = os.Remove(hashFile)

	err = os.WriteFile(hashFile, []byte(hashSum), 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadExtractedFunction[T any](path string) (T, error) {
	method, err := plugin.Open(path)
	if err != nil {
		log(err)
		os.Exit(1)
	}

	name := filepath.Base(path)
	name = strings.Replace(name, ".so", "", 1)
	fn, err := method.Lookup(name)
	if err != nil {
		return *new(T), err
	}

	return fn.(T), nil
}
