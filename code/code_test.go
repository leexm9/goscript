package code

import (
	"testing"
)

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpPop, []int{}, []byte{byte(OpPop)}},
		{OpADD, []int{}, []byte{byte(OpADD)}},
		{OpSUB, []int{}, []byte{byte(OpSUB)}},
		{OpTrue, []int{}, []byte{byte(OpTrue)}},
		{OpEQL, []int{}, []byte{byte(OpEQL)}},
		{OpJumpNotTruthy, []int{65534}, []byte{byte(OpJumpNotTruthy), 255, 254}},
		{OpJump, []int{65534}, []byte{byte(OpJump), 255, 254}},
		{OpSetGlobal, []int{65534}, []byte{byte(OpSetGlobal), 255, 254}},
		{OpGetGlobal, []int{65534}, []byte{byte(OpGetGlobal), 255, 254}},
		{OpSetFree, []int{255}, []byte{byte(OpSetFree), 255}},
		{OpGetFree, []int{255}, []byte{byte(OpGetFree), 255}},
		{OpCall, []int{255}, []byte{byte(OpCall), 255}},
		{OpClosure, []int{65534, 255}, []byte{byte(OpClosure), 255, 254, 255}},
	}

	for _, tt := range tests {
		ins := Make(tt.op, tt.operands...)

		if len(ins) != len(tt.expected) {
			t.Errorf("instruction has wrong length. want=%d, got=%d", len(tt.expected), len(ins))
		}

		for i, b := range tt.expected {
			if ins[i] != tt.expected[i] {
				t.Errorf("wrong byte at pod %d. want=%d, got=%d", i, b, ins[i])
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	ins := []Instructions{
		Make(OpConstant, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
		Make(OpADD),
		Make(OpTrue),
		Make(OpEQL),
		Make(OpJumpNotTruthy, 1),
		Make(OpJump, 1),
		Make(OpSetGlobal, 1),
		Make(OpGetGlobal, 1),
		Make(OpSetLocal, 2),
		Make(OpGetLocal, 2),
		Make(OpSetFree, 3),
		Make(OpGetFree, 3),
		Make(OpCall, 2),
		Make(OpClosure, 65534, 255),
		Make(OpPop),
	}

	expected := `0000 OpConstant 1
0003 OpConstant 2
0006 OpConstant 65535
0009 OpADD
0010 OpTrue
0011 OpEQL
0012 OpJumpNotTruthy 1
0015 OpJump 1
0018 OpSetGlobal 1
0021 OpGetGlobal 1
0024 OpSetLocal 2
0027 OpGetLocal 2
0030 OpSetFree 3
0032 OpGetFree 3
0034 OpCall 2
0036 OpClosure 65534 255
0040 OpPop
`
	var concatted Instructions
	for _, in := range ins {
		concatted = append(concatted, in...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted. \nwant=%q\ngot=%q", expected, concatted.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
		{OpSetLocal, []int{65535}, 2},
		{OpClosure, []int{65535, 255}, 3},
	}

	for _, tt := range tests {
		ins := Make(tt.op, tt.operands...)

		def, err := Lookup(byte(tt.op))
		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}

		operandsRead, n := ReadOperands(def, ins[1:])
		if n != tt.bytesRead {
			t.Fatalf("n wrong. want=%d. got=%d", tt.bytesRead, n)
		}

		for i, want := range tt.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
			}
		}
	}
}
