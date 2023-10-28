package transform

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var VerboseLog = false

func log(v ...interface{}) {
	if !VerboseLog {
		return
	}
	msg := append([]interface{}{}, GOT_PREFIX)
	msg = append(msg, v...)
	fmt.Println(msg...)
}

func convertSlice[T any](s []T) []interface{} {
	res := []interface{}{}
	for _, v := range s {
		res = append(res, v)
	}
	return res
}

func LookupGoFiles(targetDir string) []string {
	foundFiles := []string{}

	_ = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if path == "got" || (len(path) > 1 && string(path[0]) == ".") {
			return filepath.SkipDir
		}
		if strings.Contains(filepath.Base(path), "_test") {
			return nil
		}
		if strings.Contains(filepath.Base(path), "_generated") {
			return nil
		}
		if filepath.Ext(path) == GO_FILE_EXTENSION {
			foundFiles = append(foundFiles, path)
			return nil
		}
		return nil
	})

	return foundFiles
}

func writeGoFile(path string, w io.Reader) (err error) {
	nf, err := os.Create(path)
	if err != nil {
		return err
	}

	defer nf.Close()

	_, _ = io.Copy(nf, w)

	return nil
}

func GetGoRoot() (string, error) {
	// Get GOROOT
	goRoot := os.Getenv("GOROOT")

	if goRoot == "" {
		return "", fmt.Errorf("GOROOT is not set")
	}

	return goRoot, nil
}

func executeGoImports(file string) error {
	gopath, err := GetGoPath()
	if err != nil {
		return err
	}
	goImportsBin := filepath.Join(gopath, "bin", "goimports")
	cmd := exec.Command(goImportsBin, "-w", file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to execute goimports: %s", err)
	}

	return nil
}

func GetGoPath() (string, error) {
	// Get GOPATH
	goPath := os.Getenv("GOPATH")

	if goPath == "" {
		return "", fmt.Errorf("GOPATH is not set")
	}

	return goPath, nil
}

func getFileImportsList(p *ast.File) ([]*ast.ImportSpec, error) {
	return p.Imports, nil
}

func buildAsPlugin(srcPath, dstPath string) error {
	goroot, err := GetGoRoot()
	if err != nil {
		return err
	}
	goBuildBin := filepath.Join(goroot, "bin", "go")
	cmd := exec.Command(goBuildBin, "build", "-buildmode=plugin", "-o", dstPath, srcPath)
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to build plugin: %s", err)
	}
	return nil
}

func IsLineGotPrefixed(line string) bool {
	slashes := 0
	for i := 0; i < len(line); i++ {
		if line[i] == ' ' {
			continue
		}
		if line[i] == '/' {
			slashes++
			continue
		}
		if slashes == 2 {
			if line[i] == GOT_PREFIX[0] {
				return true
			} else {
				return false
			}
		}
		return false
	}

	return false
}

func stringify(d interface{}) string {
	p, _ := json.MarshalIndent(d, "", "  ")
	return string(p)
}
