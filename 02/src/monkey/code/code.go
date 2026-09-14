package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// 指令集為 byte 陣列
type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "Error: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (int *Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0: //OpAdd
		return def.Name
	case 1: //OpConstant
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

// Opcode byte
type Opcode byte

// 指令集的編號是iota遞增
const (
	OpConstant Opcode = iota

	OpAdd
	OpPop
	OpSub
	OpMul
	OpDiv

	OpTrue
	OpFalse

	OpEqual
	OpNotEqual
	OpGreaterThan
)

// definition則是有 name + operandWith int陣列
type Definition struct {
	Name          string
	OperandWidths []int
}

// 這個則是opcode進去 取得definition object
var definitions = map[Opcode]*Definition{
	OpConstant:    {"OpConstant", []int{2}},
	OpAdd:         {"OpAdd", []int{}},         // +
	OpPop:         {"OpPop", []int{}},         // poo stack
	OpSub:         {"OpSub", []int{}},         // -
	OpMul:         {"OpMul", []int{}},         // *
	OpDiv:         {"OpDiv", []int{}},         // /
	OpTrue:        {"OpTrue", []int{}},        // true
	OpFalse:       {"OpFalse", []int{}},       // false
	OpEqual:       {"OpEqual", []int{}},       // ==
	OpNotEqual:    {"OpNotEqual", []int{}},    // !=
	OpGreaterThan: {"OpGreaterThan", []int{}}, // > (以及被重寫過的 <)
}

// 查definition map 看有沒有在裡面
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]

	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

// 產生opcode
func Make(op Opcode, operands ...int) []byte {
	//查表
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	//如果裡面的operandwidth有int array推進
	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	//製作byte array for instruction
	instruction := make([]byte, instructionLen)
	// 0 位置放opcode -> header 1 byte
	instruction[0] = byte(op)

	// 1位置開始抓operands
	offset := 1
	for i, o := range operands {
		//寬度
		width := def.OperandWidths[i]
		switch width {
		case 2:
			//高位的byte排前面
			// 255, 254 類似這種
			//把o轉為uint16放到instruction[offset:]
			binary.BigEndian.PutUint16(
				instruction[offset:], uint16(o))
		}
		//offset加寬度
		offset += width
	}

	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
