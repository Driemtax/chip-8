package emulator

import (
	"fmt"
)

const (
	ScreenWidth  = 64
	ScreenHeight = 32
)

type Chip8 struct {
	Opcode uint16
	// --- REGISTERS ---
	V  [16]uint8 // 16 8-bit all purpose registers V0-VF
	I  uint16    // Index Register
	PC uint16    // Programm Counter
	// --- STACK ---
	Stack [16]uint16
	SP    uint16 // Stack Pointer
	// --- MEMORY ---
	Ram [4096]uint8
	// --- DISPLAY ---
	// The Pixels are stored bytewise for easier handling, you could pack it
	// since one pixels only stores 1 or 0
	GFX [ScreenWidth * ScreenHeight]uint8 // 1 = white, 0 = black
}

func (c *Chip8) Init() {

}

func (c *Chip8) Cycle() {
	// --- FETCH ---
	op1 := uint8(1)   //c.Ram[c.PC]
	op2 := uint8(255) //c.Ram[c.PC+1]
	//c.PC += 2
	opcode := uint16(op1)
	opcode = opcode << 8
	opcode = opcode | uint16(op2)
	fmt.Println("Fetched:", ToHex(opcode))
	// --- DECODE ---
	// --- EXECUTE ---
}

// Converts an uint16 opcode to a hex string for pretty printing
func ToHex(op uint16) string {
	n4 := op & 0x000F
	op = op >> 4
	n3 := op & 0x000F
	op = op >> 4
	n2 := op & 0x000F
	op = op >> 4
	n1 := op & 0x000F

	return fmt.Sprintf("%x%x%x%x", n1, n2, n3, n4)
}
