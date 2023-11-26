package transform

import (
	"bytes"
	"fmt"
	"testing"
)

func testParseInstruction(t *testing.T, input string, expected ...Instruction) {
	fmt.Println("CASE", input)

	parser := NewInstructionParser(
		bytes.NewBufferString(input), AllInstructions)

	result, err := parser.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != len(expected) {
		t.Fatal("invalid result")
	}

	for i, instruction := range result {
		if instruction.Type() != expected[i].Type() {
			t.Fatalf("invalid instruction type, expected %d, got %d", expected[i].Type(), instruction.Type())
		}

		fmt.Println("expected:", stringify(expected[i]))
		fmt.Println("got:", stringify(instruction))

		switch instruction.Type() {
		case AttributeInstructionType:
			assertAttribute(t, instruction.(AttributeInstruction), expected[i].(AttributeInstruction))
		}
	}
}

func assertAttribute(t *testing.T, instruction, expected AttributeInstruction) {
	if instruction.Name != expected.Name {
		t.Fatalf("invalid instruction name, expected %s, got %s", expected.Name, instruction.Name)
	}

	if len(instruction.Arguments) != len(expected.Arguments) {
		t.Fatalf("invalid instruction arguments, expected %d, got %d", len(expected.Arguments), len(instruction.Arguments))
	}

	for i, argument := range instruction.Arguments {
		if argument != expected.Arguments[i] {
			t.Fatalf("invalid instruction argument, expected %s, got %s", expected.Arguments[i], argument)
		}
	}

	if instruction.IsBuiltin != expected.IsBuiltin {
		t.Fatalf("invalid instruction builtin, expected %t, got %t", expected.IsBuiltin, instruction.IsBuiltin)
	}
}

func TestParse(t *testing.T) {
	testParseInstruction(t,
		"#[decorator]",
		AttributeInstruction{
			Name:      "decorator",
			Arguments: []string{},
			IsBuiltin: true,
		},
	)

	testParseInstruction(t,
		"#[test(1,2,3)]",
		AttributeInstruction{
			Name:      "test",
			Arguments: []string{"1", "2", "3"},
			IsBuiltin: false,
		},
	)

	testParseInstruction(t,
		"#[test, decorator]",
		AttributeInstruction{
			Name:      "test",
			Arguments: []string{},
			IsBuiltin: false,
		},
		AttributeInstruction{
			Name:      "decorator",
			Arguments: []string{},
			IsBuiltin: true,
		},
	)

	testParseInstruction(t,
		"#[test, decorator, test(1,2,3)]",
		AttributeInstruction{
			Name:      "test",
			Arguments: []string{},
			IsBuiltin: false,
		},
		AttributeInstruction{
			Name:      "decorator",
			Arguments: []string{},
			IsBuiltin: true,
		},
		AttributeInstruction{
			Name:      "test",
			Arguments: []string{"1", "2", "3"},
			IsBuiltin: false,
		},
	)

}
