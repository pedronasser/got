//go:build !generated

// Take notice that all the following code is still valid Go code,
// we are only adding some comments to hint what we need.

package json_marshal

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"time"

	. "github.com/pedronasser/got/examples/json-marshal/json"
	"github.com/pedronasser/got/transform"
)

// Just to prevent the compiler from complaining about unused imports
var _ = JSON_START_TOKEN

// #[JSON]
type User struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Usertype  string    `json:"usertype"`
	CreatedAt time.Time `json:"created_at"`
}

// #[decorator]
func JSON(fileSrc []byte, f *ast.File, c *transform.TransformContext, args ...string) (err error) {
	// We are only interested in type declarations
	target := c.Node()

	gen, ok := target.(*ast.GenDecl)
	if !ok {
		return fmt.Errorf("JSON decorator can only be used on type declarations")
	}

	structType, ok := gen.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
	if !ok {
		return fmt.Errorf("JSON decorator can only be used on structs")
	}

	getJSONFieldName := func(field *ast.Field) string {
		if field.Tag == nil {
			return field.Names[0].Name
		}

		tags := strings.Split(field.Tag.Value[1:len(field.Tag.Value)-1], ",")
		for _, tag := range tags {
			if strings.HasPrefix(tag, "json:") {
				jsonOpts := strings.SplitN(tag[6:len(tag)-1], ",", 1)
				if len(jsonOpts) > 0 {
					return jsonOpts[0]
				}
				break
			}
		}

		return strings.ToLower(field.Names[0].Name)
	}

	writeString := func(stmts *[]ast.Stmt, stringExpr ast.Expr) {
		var ellipsis token.Pos = 1
		if stringExpr, ok := stringExpr.(*ast.BasicLit); ok {
			if stringExpr.Kind == token.STRING && len(stringExpr.Value) == 1 {
				ellipsis = 0
			}
		}

		*stmts = append(*stmts, &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("b"),
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				// append(b, ...args)
				&ast.CallExpr{
					Fun: ast.NewIdent("append"),
					Args: []ast.Expr{
						ast.NewIdent("b"),
						// []byte(stringExpr)
						&ast.CallExpr{
							Fun: ast.NewIdent("[]byte"),
							Args: []ast.Expr{
								stringExpr,
							},
						},
					},
					Ellipsis: ellipsis,
				},
			},
		})
	}

	writeFieldName := func(stmts *[]ast.Stmt, fieldName string) {
		writeString(stmts, &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`\"" + fieldName + "\":`",
		})
	}

	renderJSON := func(structType *ast.StructType) (stmts []ast.Stmt) {
		writeString(&stmts, &ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"{\"",
		})

		for i, field := range structType.Fields.List {
			if len(field.Names) < 1 {
				continue
			}

			fieldName := field.Names[0].Name
			jsonFieldName := getJSONFieldName(field)
			fmt.Println("field", fieldName, jsonFieldName)

			writeFieldName(&stmts, jsonFieldName)

			switch field.Type.(type) {
			case *ast.Ident:
				switch field.Type.(*ast.Ident).Name {
				case "string":

					writeString(&stmts, &ast.BinaryExpr{
						Op: token.ADD,
						X: &ast.BasicLit{
							Kind:  token.STRING,
							Value: "`\"`",
						},
						Y: &ast.BinaryExpr{
							Op: token.ADD,
							X: &ast.SelectorExpr{
								X:   ast.NewIdent("u"),
								Sel: ast.NewIdent(fieldName),
							},
							Y: &ast.BasicLit{
								Kind:  token.STRING,
								Value: "`\"`",
							},
						},
					})
				case "int", "int64", "int32", "int16", "int8", "uint", "uint64", "uint32", "uint16", "uint8":
					writeString(&stmts,
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("strconv"),
								Sel: ast.NewIdent("Itoa"),
							},
							Args: []ast.Expr{
								&ast.SelectorExpr{
									X:   ast.NewIdent("u"),
									Sel: ast.NewIdent(fieldName),
								},
							},
						})
				default:
				}
			case *ast.ArrayType:
				// stmts = append(stmts, &ast.AssignStmt{
				// 	Lhs: []ast.Expr{
				// 		ast.NewIdent(fieldName),
				// 		ast.NewIdent("_"),
				// 	},
				// 	Tok: token.DEFINE,
				// 	Rhs: []ast.Expr{
				// 		&ast.CallExpr{
				// 			Fun: &ast.SelectorExpr{
				// 				X:   ast.NewIdent("json"),
				// 				Sel: ast.NewIdent("Marshal"),
				// 			},
				// 			Args: []ast.Expr{
				// 				&ast.SelectorExpr{
				// 					X:   ast.NewIdent("u"),
				// 					Sel: ast.NewIdent(fieldName),
				// 				},
				// 			},
				// 		},
				// 	},
				// })

				writeString(&stmts,
					&ast.BinaryExpr{
						Op: token.ADD,
						X: &ast.BasicLit{
							Kind:  token.STRING,
							Value: "`null`",
						},
						Y: &ast.BasicLit{
							Kind:  token.STRING,
							Value: "``",
						},
					},
				)

			case *ast.SelectorExpr:
				switch field.Type.(*ast.SelectorExpr).Sel.Name {
				case "Time":
					writeString(&stmts, &ast.BinaryExpr{
						Op: token.ADD,
						X: &ast.BasicLit{
							Kind:  token.STRING,
							Value: "`\"`",
						},
						Y: &ast.BinaryExpr{
							Op: token.ADD,
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.SelectorExpr{
										X:   ast.NewIdent("u"),
										Sel: ast.NewIdent(fieldName),
									},
									Sel: ast.NewIdent("Format"),
								},
								Args: []ast.Expr{
									&ast.SelectorExpr{
										X:   ast.NewIdent("time"),
										Sel: ast.NewIdent("RFC3339"),
									},
								},
							},
							Y: &ast.BasicLit{
								Kind:  token.STRING,
								Value: "`\"`",
							},
						},
					})
				}
			default:
			}

			if i < len(structType.Fields.List)-1 {
				writeString(&stmts, ast.NewIdent("JSON_SEPARATOR_TOKEN"))
			}

		}

		writeString(&stmts, &ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"}\"",
		})

		return
	}

	var implementation = []ast.Stmt{
		&ast.AssignStmt{
			// result := getBuffer()
			Lhs: []ast.Expr{
				ast.NewIdent("result"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent("GetBuffer"),
				},
			},
		},

		// defer putBuffer(result)
		&ast.DeferStmt{
			Call: &ast.CallExpr{
				Fun: ast.NewIdent("PutBuffer"),
				Args: []ast.Expr{
					ast.NewIdent("result"),
				},
			},
		},

		// b := result.AvailableBuffer()
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("b"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("result"),
						Sel: ast.NewIdent("AvailableBuffer"),
					},
				},
			},
		},

		//
	}

	implementation = append(
		implementation,
		renderJSON(structType)...,
	)

	implementation = append(implementation,
		// _, err := result.Write(b)
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("_"),
				ast.NewIdent("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("result"),
						Sel: ast.NewIdent("Write"),
					},
					Args: []ast.Expr{
						ast.NewIdent("b"),
					},
				},
			},
		},

		// if err != nil {
		// 	return nil, err
		// }
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				Op: token.NEQ,
				X:  ast.NewIdent("err"),
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("nil"),
							ast.NewIdent("err"),
						},
					},
				},
			},
		},

		// return result.Bytes(), nil
		&ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("result"),
						Sel: ast.NewIdent("Bytes"),
					},
				},
				ast.NewIdent("nil"),
			},
		},
	)

	// implement MarshalJSON() []byte, error
	method := &ast.FuncDecl{
		Name: ast.NewIdent("MarshalJSON"),
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						ast.NewIdent("u"),
					},
					Type: &ast.StarExpr{
						X: ast.NewIdent("User"),
					},
				},
			},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("[]byte"),
					},
					{
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: implementation,
		},
	}

	c.InsertAfter(method)

	return nil
}
