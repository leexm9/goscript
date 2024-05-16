package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Opcode byte

const (
	OpConstant Opcode = iota
	OpPop

	OpADD
	OpSUB
	OpMUL
	OpQUO
	OpREM

	OpAND
	OpOR
	OpXOR
	OpSHL
	OpSHR
	OpAND_NOT

	OpADD_ASSIGN
	OpSUB_ASSIGN
	OpMUL_ASSIGN
	OpQUO_ASSIGN
	OpREM_ASSIGN

	OpAND_ASSIGN
	OpOR_ASSIGN
	OpXOR_ASSIGN
	OpSHL_ASSIGN
	OpSHR_ASSIGN
	OpAND_NOT_ASSIGN

	OpLAND
	OpLOR
	OpARROW
	OpINC
	OpDEC

	OpEQL
	OpLSS
	OpGTR
	OpASSIGN
	OpNOT

	OpNEQ
	OpLEQ
	OpGEQ
	OpDEFINE
	OpELLIPSIS

	// ---------------------
	OpPrefixSub
	OpPrefixAdd

	OpTrue
	OpFalse
	OpNull

	OpJumpNotTruthy
	OpJump

	OpSetGlobal
	OpGetGlobal
	OpSetLocal
	OpGetLocal
	OpSetFree
	OpGetFree
	OpGetBuiltin
	OpSetNil

	OpArray
	OpHash
	OpIndex

	OpCall
	OpReturnValue
	OpReturn

	OpClosure
	OpCurrentClosure

	OpForLoop
	OpRangeLoop
	OpContinue
	OpBreak

	OpSetGlobalIndex
	OpSetLocalIndex
)

var codeLitMap = map[Opcode]string{
	OpConstant: "const",
	OpPop:      "pop",

	OpADD: "+",
	OpSUB: "-",
	OpMUL: "*",
	OpQUO: "/",
	OpREM: "%",

	OpAND:     "&",
	OpOR:      "|",
	OpXOR:     "^",
	OpSHL:     "<<",
	OpSHR:     ">>",
	OpAND_NOT: "&^",

	OpADD_ASSIGN: "+=",
	OpSUB_ASSIGN: "-=",
	OpMUL_ASSIGN: "*=",
	OpQUO_ASSIGN: "/=",
	OpREM_ASSIGN: "%=",

	OpAND_ASSIGN:     "$=",
	OpOR_ASSIGN:      "!=",
	OpXOR_ASSIGN:     "^=",
	OpSHL_ASSIGN:     "<<=",
	OpSHR_ASSIGN:     ">>=",
	OpAND_NOT_ASSIGN: "&^=",

	OpLAND:  "&&",
	OpLOR:   "||",
	OpARROW: "<-",
	OpINC:   "++",
	OpDEC:   "--",

	OpEQL:    "==",
	OpLSS:    "<",
	OpGTR:    ">",
	OpASSIGN: "=",
	OpNOT:    "!",

	OpNEQ:      "!=",
	OpLEQ:      "<=",
	OpGEQ:      ">=",
	OpDEFINE:   ":=",
	OpELLIPSIS: "...",

	// ---------------------
	OpPrefixSub: "-",
	OpPrefixAdd: "+",

	OpTrue:  "true",
	OpFalse: "false",
	OpNull:  "nil",

	OpJumpNotTruthy: "ifFalse",
	OpJump:          "jump",

	OpSetGlobal:  "setG",
	OpGetGlobal:  "getG",
	OpSetLocal:   "setL",
	OpGetLocal:   "getL",
	OpSetFree:    "setF",
	OpGetFree:    "getF",
	OpGetBuiltin: "getBuiltin",
	OpSetNil:     "setNil",

	OpArray: "array",
	OpHash:  "hash",
	OpIndex: "index",

	OpCall:        "call",
	OpReturnValue: "return",
	OpReturn:      "return",

	OpClosure:        "closure",
	OpCurrentClosure: "curClosure",

	OpForLoop:   "forLoop",
	OpRangeLoop: "rangeLoop",
	OpContinue:  "continue",
	OpBreak:     "break",

	OpSetGlobalIndex: "setGIndex",
	OpSetLocalIndex:  "setLIndex",
}

func (o Opcode) String() string {
	return codeLitMap[o]
}

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitons = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpPop:      {"OpPop", []int{}},

	OpADD: {"OpADD", []int{}}, // +
	OpSUB: {"OpSUB", []int{}}, // -
	OpMUL: {"OpMUL", []int{}}, // *
	OpQUO: {"OpQUO", []int{}}, // /
	OpREM: {"OpREM", []int{}}, // %

	OpAND:     {"OpAND", []int{}},     // &
	OpOR:      {"OpOR", []int{}},      // |
	OpXOR:     {"OpXOR", []int{}},     // ^
	OpSHL:     {"OpSHL", []int{}},     // <<
	OpSHR:     {"OpSHR", []int{}},     // >>
	OpAND_NOT: {"OpAND_NOT", []int{}}, // &^

	OpADD_ASSIGN: {"OpADD_ASSIGN", []int{}}, // +=
	OpSUB_ASSIGN: {"OpSUB_ASSIGN", []int{}}, // -=
	OpMUL_ASSIGN: {"OpMUL_ASSIGN", []int{}}, // *=
	OpQUO_ASSIGN: {"OpQUO_ASSIGN", []int{}}, // /=
	OpREM_ASSIGN: {"OpREM_ASSIGN", []int{}}, // %=

	OpAND_ASSIGN:     {"OpAND_ASSIGN", []int{}},     // &=
	OpOR_ASSIGN:      {"OpOR_ASSIGN", []int{}},      // |=
	OpXOR_ASSIGN:     {"OpXOR_ASSIGN", []int{}},     // ^=
	OpSHL_ASSIGN:     {"OpSHL_ASSIGN", []int{}},     // <<=
	OpSHR_ASSIGN:     {"OpSHR_ASSIGN", []int{}},     // >>=
	OpAND_NOT_ASSIGN: {"OpAND_NOT_ASSIGN", []int{}}, // &^=

	OpLAND:  {"OpLAND", []int{}},  // &&
	OpLOR:   {"OpLOR", []int{}},   // ||
	OpARROW: {"OpARROW", []int{}}, // <-
	OpINC:   {"OpINC", []int{}},   // ++
	OpDEC:   {"OpDEC", []int{}},   // --

	OpEQL:    {"OpEQL", []int{}},    // ==
	OpLSS:    {"OpLSS", []int{}},    // <
	OpGTR:    {"OpGTR", []int{}},    // >
	OpASSIGN: {"OpASSIGN", []int{}}, // =
	OpNOT:    {"OpNOT", []int{}},    // !

	OpNEQ:      {"OpNEQ", []int{}},      // !=
	OpLEQ:      {"OpLEQ", []int{}},      // <=
	OpGEQ:      {"OpGEQ", []int{}},      // >=
	OpDEFINE:   {"OpDEFINE", []int{}},   // :=
	OpELLIPSIS: {"OpELLIPSIS", []int{}}, // ...

	// ---------------------
	OpPrefixSub: {"OpPrefixSub", []int{}},
	OpPrefixAdd: {"OpPrefixAdd", []int{}},

	OpTrue:  {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},
	OpNull:  {"OpNull", []int{}},

	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpJump:          {"OpJump", []int{2}},

	OpSetGlobal:  {"OpSetGlobal", []int{2}},
	OpGetGlobal:  {"OpGetGlobal", []int{2}},
	OpSetLocal:   {"OpSetLocal", []int{2}},
	OpGetLocal:   {"OpGetLocal", []int{2}},
	OpSetFree:    {"OpSetFree", []int{1}},
	OpGetFree:    {"OpGetFree", []int{1}},
	OpGetBuiltin: {"OpGetBuiltin", []int{1}},
	OpSetNil:     {"OpSetNil", []int{}},

	OpArray: {"OpArray", []int{2}},
	OpHash:  {"OpHash", []int{2}},
	OpIndex: {"OpIndex", []int{}},

	OpCall:        {"OpCall", []int{1}},
	OpReturnValue: {"OpReturnValue", []int{1}},
	OpReturn:      {"OpReturn", []int{}},

	OpClosure:        {"OpClosure", []int{2, 1}},
	OpCurrentClosure: {"OpCurrentClosure", []int{}},

	OpForLoop:   {"OpForLoop", []int{}},
	OpRangeLoop: {"OpRangeLoop", []int{}},
	OpContinue:  {"OpContinue", []int{}},
	OpBreak:     {"OpBreak", []int{}},

	OpSetGlobalIndex: {"OpSetGlobalIndex", []int{2}},
	OpSetLocalIndex:  {"OpSetLocalIndex", []int{2}},
}

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	offset := 0
	for offset < len(ins) {
		def, err := Lookup(ins[offset])
		if err != nil {
			_, _ = fmt.Fscanf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[offset+1:])
		_, _ = fmt.Fprintf(&out, "%04d %s\n", offset, ins.fmtInstruction(def, operands))
		offset = offset + 1 + read
	}
	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	default:
		return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
	}
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitons[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitons[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 1:
			instruction[offset] = byte(o)
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += width
	}
	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))

	offset := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUint8(ins Instructions) uint8 {
	return ins[0]
}

func ReadUint16(instructions Instructions) uint16 {
	return binary.BigEndian.Uint16(instructions)
}
