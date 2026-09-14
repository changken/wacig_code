package code

import (
	"testing"
)

func TestMake(t *testing.T) {
	//他的意思是 1byte, 255 byte, 254 byte
	//這樣切割
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		//給opcode + 65534
		//應該要幫忙吐出 1byte, 255, 254
		{OpConstant, []int{65534}, []byte{
			byte(OpConstant), 255, 254,
		}},
	}

	for _, tt := range tests {
		//轉Opcode
		instruction := Make(tt.op, tt.operands...)

		//比對Instruction總元素
		if len(instruction) != len(tt.expected) {
			t.Errorf("instruction has wrong length. want=%d, got=%d",
				len(tt.expected), len(instruction))
		}

		//比對裡面的數值
		for i, b := range tt.expected {
			if instruction[i] != tt.expected[i] {
				t.Errorf("wrong byte at pos %d. want=%d, got=%d",
					i, b, instruction[i])
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpConstant, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected := `0000 OpConstant 1
0003 OpConstant 2
0006 OpConstant 65535
`

	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted.\nwant=%q\ngot=%q",
			expected, concatted.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		byteRead int
	}{
		{OpConstant, []int{65535}, 2},
	}

	for _, tt := range tests {
		//make用於編opcode
		instruction := Make(tt.op, tt.operands...)

		def, err := Lookup(byte(tt.op))

		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}

		//ReadOperands 用於解opcode碼
		operandsRead, n := ReadOperands(def, instruction[1:])
		if n != tt.byteRead {
			t.Fatalf("n wrong. want=%d, got=%d", tt.byteRead, n)
		}

		for i, want := range tt.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong. want=%d, got=%d",
					want, operandsRead[i])
			}
		}
	}
}
