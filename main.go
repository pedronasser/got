package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/pedronasser/got/transform"
)

var Version string

// main is the entry point of the got command.
// It checks if the command is a got command and executes it.
// If it's not a got command, it executes the go command.
func main() {
	args := getArgs()

	if len(args) > 1 && (args[1] == "build" || args[1] == "run" || args[1] == "test") {
		err := runGotCmd(args...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	if len(args) > 1 && args[1] == "version" {
		fmt.Printf("got version %s\n", Version)
	}

	err := runGoCmd(args[1:]...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Running go command:", args)
}

// getArgs returns the command line arguments.
// It also sets the VerboseLog variable if the -v flag is present.
func getArgs() []string {
	args := os.Args[:]
	for i, arg := range args {
		if arg == "-v" {
			VerboseLog = true
		}
		if strings.HasSuffix(arg, GO_FILE_EXTENSION) {
			args[i] = strings.Replace(arg, GO_FILE_EXTENSION, ".go", 1)
		}
	}
	return args
}

// runGotCmd executes a got command.
// It checks if the command is a build, run or test command.
// It checks if the -tags flag is present and adds the "generated" tag.
// It checks if the -o flag is present and sets the output file.
// Then it executes the got transformer.
// Then it executes the go command with the transformed arguments.
func runGotCmd(args ...string) error {
	commandName := args[1]

	args = args[1:]

	tagsFound := false
	outputFile := ""

	for i, arg := range args {
		if arg == "-tags" {
			tagsFound = true
			args[i+1] = args[i+1] + ",generated"
		} else if arg == "-o" {
			outputFile = args[i+1]
		}
	}

	if commandName == "run" {
		tmpFile, err := os.CreateTemp("", "gobuild")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		tmpFile.Close()

		outputFile = tmpFile.Name()
		args = append(args[:len(args)-1], "-o", outputFile, args[len(args)-1])
	}

	if !tagsFound {
		args = append([]string{args[1], "-tags", "generated"}, args[2:]...)
	}

	targetDir, err := getTargetDirectory(args[len(args)-1])
	if err != nil {
		return err
	}

	transformer := GotTransform(targetDir)
	if err := transformer.Execute(); err != nil {
		return err
	}

	switch commandName {
	case "build":
		args[len(args)-1] = targetDir
		_, err := runBuild(args)
		if err != nil {
			return err
		}
	case "run":
		args[len(args)-1] = targetDir
		isBuildSuccess, err := runBuild(args)
		if err != nil {
			return err
		}

		if !isBuildSuccess {
			fmt.Println("Build failed")
			os.Exit(1)
		}

		err = runProgram(outputFile)
		if err != nil {
			return err
		}

	case "test":
		err := runTest(args...)
		if err != nil {
			return err
		}
	}

	return nil
}

// runGoCmd executes a go command with the arguments received
func runGoCmd(args ...string) error {
	goroot, err := GetGoRoot()
	if err != nil {
		return err
	}

	goCmd := filepath.Join(goroot, "bin", "go")
	cmd := exec.Command(goCmd, args...)

	// Set the environment variables
	cmd.Env = os.Environ()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Set the working directory
	cmd.Dir = cwd

	// Set the stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to execute go command: %s", err)
	}

	return nil
}

// getTargetDirectory returns the target directory of the go command.
func getTargetDirectory(args ...string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	resolved, _ := filepath.Rel(cwd, args[len(args)-1])
	return filepath.Dir(resolved), nil
}

// runBuild executes the go build command with the arguments received
func runBuild(args []string) (bool, error) {
	args[0] = "build"
	fmt.Println("Building:", args)

	goroot, err := GetGoRoot()
	if err != nil {
		return false, err
	}

	goCmd := filepath.Join(goroot, "bin", "go")
	cmd := exec.Command(goCmd, args...)

	// Set the environment variables
	cmd.Env = os.Environ()

	cwd, err := os.Getwd()
	if err != nil {
		return false, err
	}

	// Set the working directory
	cmd.Dir = cwd

	// Set the stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err = cmd.Run(); err != nil {
		return false, err
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return false, fmt.Errorf("build failed")
	}

	return true, nil
}

// runProgram executes the program in the specified path
func runProgram(programPath string) error {
	fmt.Println("Running:", programPath)

	_ = os.Chmod(programPath, 0755)

	cmd := exec.Command(programPath)

	// Set the environment variables
	cmd.Env = os.Environ()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Set the working directory
	cmd.Dir = cwd

	// Set the stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err = cmd.Run()
	if err != nil {
		return err
	}
	_ = os.Remove(programPath)

	return nil
}

// runTest executes the go test command with the arguments received
func runTest(args ...string) error {
	args[0] = "test"

	goroot, err := GetGoRoot()
	if err != nil {
		return err
	}

	goCmd := filepath.Join(goroot, "bin", "go")
	fmt.Println("Testing:", args)
	cmd := exec.Command(goCmd, args...)

	// Set the environment variables
	cmd.Env = os.Environ()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Set the working directory
	cmd.Dir = cwd

	// Set the stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err = cmd.Run(); err != nil {
		return err
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf("test failed")
	}

	return nil
}
