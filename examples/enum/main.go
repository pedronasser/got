//go:build !generated

// Take notice that all the following code is still valid Go code,
// we are only adding some comments to hint what we need.

package main

import (
	"fmt"
	"go/ast"
	"go/token"

	got "github.com/pedronasser/got/transform"
)

// #[Enum]
type PokemonType string

// #[Options(PokemonType)]
const (
	NormalPokemon   PokemonType = "Normal"
	FirePokemon     PokemonType = "Fire"
	WaterPokemon    PokemonType = "Water"
	GrassPokemon    PokemonType = "Grass"
	ElectricPokemon PokemonType = "Electric"
	IcePokemon      PokemonType = "Ice"
	FightingPokemon PokemonType = "Fighting"
	PoisonPokemon   PokemonType = "Poison"
	GroundPokemon   PokemonType = "Ground"
	FlyingPokemon   PokemonType = "Flying"
	PsychicPokemon  PokemonType = "Psychic"
	BugPokemon      PokemonType = "Bug"
	RockPokemon     PokemonType = "Rock"
	GhostPokemon    PokemonType = "Ghost"
	DragonPokemon   PokemonType = "Dragon"
	DarkPokemon     PokemonType = "Dark"
	SteelPokemon    PokemonType = "Steel"
	FairyPokemon    PokemonType = "Fairy"
)

// Note that this value is not inside the previous const block which contains
// all the valid enum options
const InvalidPokemon = "InvalidPokemon"

func main() {
	// The following expression will panic during build time
	var ptype PokemonType = NormalPokemon

	// The following expression by default doesn't throw error during build time
	// But now it will throw an error because we have added the #[Enum] and #[Options] attributes
	// if you execute using the got command
	ptype = InvalidPokemon

	fmt.Println(ptype)
}

// #[decorator]
func Options(c *got.TransformContext) error {
	file := c.File
	args := c.Args()
	if len(args) < 1 {
		fmt.Println("Enum attribute requires the name of the type")
		return nil
	}

	enumName := args[0]
	fmt.Printf("Creating options for `%s`", enumName)
	for _, decl := range file.Decls {
		if v, ok := decl.(*ast.GenDecl); ok {
			if v.Tok != token.TYPE {
				continue
			}
			if len(v.Specs) < 1 {
				continue
			}
			if typeSpec, ok := v.Specs[0].(*ast.TypeSpec); ok {
				if typeSpec.Name.Name == enumName {
					if _, ok := typeSpec.Type.(*ast.Ident); ok {
						fmt.Printf("Found enum type `%s`", enumName)
					}
				}
			}
		}
	}

	enumValues := map[string]string{}
	if v, ok := c.Node().(*ast.GenDecl); ok {
		if v.Tok != token.CONST {
			return nil
		}
		for _, spec := range v.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				fmt.Printf("Found enum `%s` value `%s`\n", args[0], valueSpec.Names[0].Name)
				enumValues[valueSpec.Names[0].Name] = valueSpec.Values[0].(*ast.BasicLit).Value
			}
		}
	}

	// Create interface
	c.InsertBefore(&ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(enumName),
				Type: &ast.InterfaceType{
					Methods: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									ast.NewIdent(fmt.Sprintf("__%s", enumName)),
								},
								Type: &ast.FuncType{},
							},
						},
					},
				},
			},
		},
	})

	for k := range enumValues {
		enumType := ast.NewIdent(fmt.Sprintf("__%s", k))

		c.InsertBefore(&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: enumType,
					Type: &ast.Ident{
						Name: "string",
					},
				},
			},
		})

		enumTypeMethod := ast.NewIdent(fmt.Sprintf("__%s", enumName))
		c.InsertBefore(&ast.FuncDecl{
			Name: enumTypeMethod,
			Type: &ast.FuncType{},
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.Ident{
							Name: enumType.Name,
						},
					},
				},
			},
			Body: &ast.BlockStmt{},
		})

		c.InsertBefore(&ast.GenDecl{
			Tok: token.CONST,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Type: enumType,
					Names: []*ast.Ident{
						ast.NewIdent(k),
					},
					Values: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: enumValues[k],
						},
					},
				},
			},
		})
	}

	c.Delete()

	return nil
}

// #[decorator]
func Enum(c *got.TransformContext) (err error) {
	c.Delete()

	return nil
}
