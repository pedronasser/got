package transform

import (
	"io"
)

const (
	// AttributeInstructionType is the type of an attribute instruction.
	AttributeInstructionType uint = 1 << iota
)

// AllInstructions is a mask that includes all instruction types.
const AllInstructions = AttributeInstructionType

// Instruction is an interface that represents an instruction.
type Instruction interface {
	Type() uint
}

// AttributeInstruction is an instruction that represents an attribute.
type AttributeInstruction struct {
	Name      string
	Arguments []string
	IsBuiltin bool
}

// Type returns the type of the instruction.
func (a AttributeInstruction) Type() uint {
	return AttributeInstructionType
}

// InstructionParser is a parser that parses instructions.
type InstructionParser struct {
	src        io.Reader
	parseTypes uint
	result     []Instruction
}

// NewInstructionParser creates a new instruction parser.
func NewInstructionParser(src io.Reader, parseTypes uint) *InstructionParser {
	return &InstructionParser{
		src:        src,
		parseTypes: parseTypes,
		result:     []Instruction{},
	}
}

// Parse trys to parse the instructions from the source.
func (p *InstructionParser) Parse() ([]Instruction, error) {
	buf := make([]byte, 1)

	for {
		_, err := p.src.Read(buf)
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		switch string(buf)[0] {
		case AttributeListStart:
			if p.parseTypes&AttributeInstructionType == 0 {
				continue
			}
			p.parseAttributesList()
		default:
		}

	}

	return p.result, nil
}

// The following constants are used to parse instructions.
const (
	Space                = '\u0020' // space
	InstructionsStart    = '\u0023' // #
	AttributeListStart   = '\u005B' // ]
	AttributeListEnd     = '\u005D' // ]
	AttributeSeparator   = '\u002C' // ,
	AttributeParamsStart = '\u0028' // (
	AttributeParamsEnd   = '\u0029' // )
)

// parseAttributesList trys to parse the attributes list from the source.
func (p *InstructionParser) parseAttributesList() {
	buf := make([]byte, 1)
	var name string
	var args []string

	for {
		_, err := p.src.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}

		switch buf[0] {
		case AttributeParamsEnd, Space:
			continue

		case AttributeParamsStart:
			args = p.parseAttributesParams()
			continue

		case AttributeSeparator, AttributeListEnd:
			var builtin bool = false
			if _, ok := BuiltinAttributes[name]; ok {
				builtin = true
			}

			p.result = append(p.result, AttributeInstruction{
				Name:      name,
				Arguments: args,
				IsBuiltin: builtin,
			})

			args = []string{}
			name = ""
			continue
		}
		name += string(buf[0])
	}
}

// parseAttributesParams trys to parse the attributes params from the source.
func (p *InstructionParser) parseAttributesParams() []string {
	buf := make([]byte, 1)
	var args []string
	var arg string

	for {
		_, err := p.src.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil
		}

		switch buf[0] {

		case AttributeParamsEnd:
			args = append(args, arg)
			return args

		case AttributeSeparator:
			args = append(args, arg)
			arg = ""
			continue

		}
		arg += string(buf[0])
	}

	return args
}
