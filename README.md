# GOT - Go Language Transformer (experimental)

**GOT** is a simple tool to transform Go code. It uses the [Go AST](https://golang.org/pkg/go/ast/) to parse and transform the code.

It works by parsing attributes (similar to Rust's [attributes](https://doc.rust-lang.org/reference/attributes.html)) and transforming the code based on them. The transformed code is written to a new file with an added suffix `_generated` to its name.

**DISCLAIMERS:**
- This is an experimental project.
- It's definitely not the best way to do this. This is just a proof of concept.
- The API is not stable and will change

### Goals

- Provide safe and ergonomic ways to transform Go code inside a project
- Allow devs to create their own transformations specific to their projects

### Non-Goals

- This tool is not intended to be used to publish/use third-party transformations (macros, decorators, etc).

### Features
- ✅ [Attributes](#attributes)
- ✅ [Decorators](#decorator)

### Missing

- [ ] Cache transformations (based on build tags)
- [ ] Test coverage

## Installation

### Requirements

- Go 1.21+

### Go Install

```bash
go install github.com/pedronasser/got
```

### Other methods

Not available yet.

## Usage

For ergonomic reasons, you can use GOT just like you would use `go` command.
Got performs the transformation on the target package and then runs the `go` command with the same arguments.

```
got <run/build/test> [-v] <target>
```

`-v` - Verbose mode

For example:

```bash
got run main.go
```

or 

```bash
got build .
```

## Transformations

Transformations are performed by parsing comments with the following format:

```go
// #[attr(arg1, ...), ...]
```

As you can see we can specify multiple attributes in the same comment.

Each attribute can receive arguments separated by commas and will be executed one after the another.

If any attribute execution fails, that transformation will be aborted.

### Attributes

**Attributes** are structured comments used to specify what transformations should be performed on the following expression or declaration.

#### Builtin attributes

Builtin attributes are executed before any user-defined attributes. Only builtin attributes have lowercase names.

- `#[decorator]` - Creates a [decorator](#Decorator) function. 

- `#[method]` - Creates a [method](#Method).

- `#[tag]` - Specify which build tag must be present for the following expression or declaration to be transformed. 
It expects `go:build` constraints as the argument

### Decorator

**Decorators** are functions the transform an expression or declaration.

They are created by adding the attribute `#[decorator]` to a function having the following signarure:

```go
func (c *got.TransformContext) (err error)
```

####

<details>
    <summary>Examples</summary>

### Creating a decorator

```go
package main

import (
    got "github.com/pedronasser/got"
)

func main() {
    Hello("World")
}

// This function will be transformed by the Log decorator
//#[Log]
func Hello(input string) {
	fmt.Println("Hello", input)
}

// Here we are creating a decorator called Log
// It will log the function name and the arguments received
// #[decorator]
func Log(c *got.TransformContext) error {
	node := c.Node()
	fn := node.(*ast.FuncDecl)

	printFormat := "\"Called func %s("
	for i, arg := range fn.Type.Params.List {
		if i > 0 {
			printFormat += ", "
		}
		printFormat = printFormat + arg.Names[0].Name + ": %v"
	}
	printFormat += ")\\n\""

	printArgs := []ast.Expr{
		&ast.BasicLit{Kind: token.STRING, Value: printFormat},
		&ast.BasicLit{Kind: token.STRING, Value: "\"" + fn.Name.Name + "\""},
	}
	for _, arg := range fn.Type.Params.List {
		printArgs = append(printArgs, arg.Names[0])
	}

	fn.Body.List = append([]ast.Stmt{
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent("fmt"),
					Sel: ast.NewIdent("Printf"),
				},
				Args: printArgs,
			},
		},
	}, fn.Body.List...)

	c.Replace(node)

	return nil
}
```

Check the [examples](examples/) folder for more examples.

</details>







