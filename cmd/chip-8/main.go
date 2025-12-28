package main

import (
	"chip-8/internal/emulator"
	"fmt"
)

func main() {
	op1 := uint8(1)   //c.Ram[c.PC]
	op2 := uint8(255) //c.Ram[c.PC+1]
	//c.PC += 2
	opcode := uint16(op1)
	opcode = opcode << 8
	opcode = opcode | uint16(op2)
	fmt.Println("Fetched:", emulator.ToHex(opcode))
}
