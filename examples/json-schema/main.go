//go:build !generated

// Take notice that all the following code is still valid Go code,
// we are only adding some comments to hint what we need.

package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/pedronasser/got/transform"
)

// #[FromSchema(user.schema.json)]
type User struct {
}

func main() {
	f, err := os.OpenFile("user.json", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	var user User

	err = json.NewDecoder(f).Decode(&user)
	if err != nil {
		panic(err)
	}

	pretty, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(pretty))
}

// FromSchema is a decorator that will generate a struct from a JSON schema
// The first argument is the path (relative to the current directory) of the JSON schema file
// #[decorator]
func FromSchema(c *transform.TransformContext) (err error) {
	args := c.Args()
	if len(args) < 1 {
		return fmt.Errorf("FromSchema decorator requires the path to the JSON schema file")
	}

	target := c.Node()

	// We are only interested in type declarations
	if _, ok := target.(*ast.GenDecl); !ok {
		// By returning a nil []byte we are not replacing anything
		return nil
	}

	// We are only interested in type declarations
	if target.(*ast.GenDecl).Tok != token.TYPE || len(target.(*ast.GenDecl).Specs) < 1 {
		return nil
	}

	// We are only interested in type declarations
	spec := target.(*ast.GenDecl).Specs[0]
	if _, ok := spec.(*ast.TypeSpec); !ok {
		return nil
	}

	// We are only interested in struct types
	typeSpec := spec.(*ast.TypeSpec)
	if _, ok := typeSpec.Type.(*ast.StructType); !ok {
		return nil
	}

	// Based on https://json-schema.org/draft/2020-12/schema
	type JSONSchema struct {
		// We are only interested in the properties of the JSON schema
		Properties map[string]interface{} `json:"properties,omitempty"`
	}

	// Open the JSON schema file
	f, err := os.OpenFile(args[0], os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	var schema JSONSchema

	// Decode the JSON schema file
	err = json.NewDecoder(f).Decode(&schema)
	if err != nil {
		return err
	}

	var createStruct func(map[string]interface{}) (*ast.StructType, error)
	var createField func(string, map[string]interface{}) (*ast.Field, error)

	createStruct = func(fields map[string]interface{}) (*ast.StructType, error) {
		s := &ast.StructType{}

		list := []*ast.Field{}
		for fieldName, attrs := range fields {
			parts := strings.Split(fieldName, "_")
			for i, part := range parts {
				parts[i] = strings.ToUpper(string(part[0])) + part[1:]
			}
			formattedName := strings.Join(parts, "")

			field, err := createField(formattedName, attrs.(map[string]interface{}))
			if err != nil {
				return nil, err
			}

			if field == nil {
				continue
			}

			field.Tag = &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("`json:\"%s\"`", fieldName),
			}

			list = append(list, field)
		}

		s.Fields = &ast.FieldList{
			List: list,
		}

		return s, nil
	}

	createField = func(name string, fieldProps map[string]interface{}) (*ast.Field, error) {
		field := &ast.Field{
			Names: []*ast.Ident{
				ast.NewIdent(name),
			},
		}

		fieldType, ok := fieldProps["type"].(string)
		if !ok {
			return nil, fmt.Errorf("missing field type %s", fieldType)
		}

		switch fieldType {
		case "string":
			if v, ok := fieldProps["format"]; ok {
				switch v {
				case "date-time":
					field.Type = &ast.SelectorExpr{
						X:   ast.NewIdent("time"),
						Sel: ast.NewIdent("Time"),
					}
				default:
					return nil, fmt.Errorf("unknown field format %s", v)
				}
			} else {
				field.Type = ast.NewIdent("string")
			}
		case "integer":
			field.Type = ast.NewIdent("int")
		case "number":
			field.Type = ast.NewIdent("float64")
		case "boolean":
			field.Type = ast.NewIdent("bool")
		case "array":
			subfield, err := createField("item", fieldProps["items"].(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			field.Type = &ast.ArrayType{
				Elt: subfield.Type,
			}
		case "object":
			// field.Type = ast.NewIdent("interface{}")
			obj, err := createStruct(fieldProps["properties"].(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			field.Type = obj
		default:
			return nil, fmt.Errorf("unknown field type %s", fieldType)
		}

		return field, nil
	}

	s, err := createStruct(schema.Properties)
	if err != nil {
		return nil
	}

	c.Replace(&ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: typeSpec.Name,
				Type: s,
			},
		},
	})

	return nil
}
