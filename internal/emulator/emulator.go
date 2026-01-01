package emulator

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	c.Ram = [4096]uint8{}
	c.PC = 512
	c.GFX = [ScreenWidth * ScreenHeight]uint8{0}
}

// Loads a programm into ram. If the programm is too long, there will be an error.
func (c *Chip8) LoadROM(name string) error {
	if name == "" {
		return errors.New("No name specified!")
	}

	path := filepath.Join("assets", name)
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("There was an error reading the file:", err)
	}

	if len(data) > 4096-512 {
		return errors.New("Programm is too long to be read into RAM. Maximal allowed size is 3584 Bytes!")
	}

	startOffset := 512 // We start at 512 since 0-511 is reserved for the fonts
	for i, b := range data {
		c.Ram[startOffset+i] = b
	}

	return nil
}

// Fetches, Decodes and Executes an instruction from RAM
func (c *Chip8) Cycle() {
	// --- FETCH ---
	op1 := c.Ram[c.PC]
	op2 := c.Ram[c.PC+1]
	c.PC += 2 // always add 2 to the PC since we read in 2 Bytes per instruction
	opcode := uint16(op1)
	opcode = opcode << 8
	opcode = opcode | uint16(op2)
	if opcode != 0x1228 {
		fmt.Println("Fetched:", ToHex(opcode))
	}

	// --- DECODE & EXECUTE ---
	c.Execute(opcode)
}

func (c *Chip8) Execute(opcode uint16) {
	// The first nibble holds the instruction
	switch opcode & 0xF000 {

	case 0x0000:
		// 0x00E0 CLS = Clear the Screen
		if opcode == ClearScreen {
			c.OpClearScreen()
		} else if opcode == ReturnFromSubroutine {
			fmt.Println("RET - Return from Subroutine")
			// TODO: reduce SP, set PC
		}

	// Jumps to NNN
	case 0x1000:
		c.OpJump(opcode)

	// Sets VX = NN
	case 0x6000:
		c.OpSetRegister(opcode)

	// VX += NN
	case 0x7000:
		c.OpAddToRegister(opcode)

	// Sets I = NNN (Index Register)
	case 0xA000:
		c.OpSetIndexRegister(opcode)

	// DXYN
	// Draws a sprite at VX, VY with width = 8px and height = Npx
	case 0xD000:
		c.OpDraw(opcode)
	default:
		fmt.Println("Unknown Opcode:", ToHex(opcode))
	}
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
