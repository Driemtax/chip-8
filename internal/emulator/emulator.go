package emulator

import (
	"chip-8/pkg"
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

// Those are the hex values for the pixels of the Digits aka the Fonts
var fontSet = [80]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

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

	// --- TIMER ---
	DelayTimer uint8
	SoundTimer uint8

	// --- INPUT ---
	Key [16]uint8 // current state of the hex-keypad
	// Helper for FX0A to wait for key release
	// 255 = no key stored, otherwise stores the key index (0-15)
	KeyPressedBuffer uint8

	// --- LOGGING ---
	Logger pkg.Logger
}

func (c *Chip8) Init() {
	c.Ram = [4096]uint8{0}
	c.Stack = [16]uint16{0}
	c.V = [16]uint8{0}
	c.PC = 512
	c.GFX = [ScreenWidth * ScreenHeight]uint8{0}
	c.SP = 0
	c.I = 512
	c.KeyPressedBuffer = 255 // init buffer with "empty" value

	// Load fonts into Ram
	for i, v := range fontSet {
		c.Ram[i] = v
	}

	// New empty Logger instance for logging the instructions (debugging)
	c.Logger = pkg.Logger{}
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
			c.OpReturnFromSubroutine()
		}

	// Jumps to NNN
	case Jump:
		c.OpJump(opcode)

	// Calls Subroutine at address NNN
	case CallSubroutine:
		c.OpCallSubroutine(opcode)

	// Skips next instruction if VX == NN
	case SkipIfEquals:
		c.OpSkipIfEqulas(opcode)

	// Skips the next instruction if VX != NN
	case SkipIfNotEquals:
		c.OpSkipIfNotEqual(opcode)

	// Sets VX = NN
	case SetVX:
		c.OpSetRegister(opcode)

	// VX += NN
	case AddVX:
		c.OpAddToRegister(opcode)

	// Math and Bitwise operators
	case 0x8000:
		switch opcode & 0xF00F {
		case SetVXtoVY:
			c.OpSetVXtoVY(opcode)
		case BitwiseOR:
			c.OpBitwiseOR(opcode)
		case BitwiseAND:
			c.OpBitwiseAND(opcode)
		case BitwiseXOR:
			c.OpBitwiseXOR(opcode)
		case Add:
			c.OpAdd(opcode)
		case Sub:
			c.OpSub(opcode)
		case RightShift:
			c.OpRightShift(opcode)
		case SubReverse:
			c.OpSubReverse(opcode)
		case LeftShift:
			c.OpLeftShift(opcode)
		}

	case SkipRegistersNotEqual:
		c.OpSkipRegistersNotEqual(opcode)

	// Sets I = NNN (Index Register)
	case SetIndex:
		c.OpSetIndexRegister(opcode)

	case JumpPlusV0:
		c.OpJumpPlusV0(opcode)

	case Random:
		c.OpRandom(opcode)
	// DXYN
	// Draws a sprite at VX, VY with width = 8px and height = Npx
	case DrawSprite:
		c.OpDraw(opcode)

	case 0xE000:
		switch opcode & 0xF0FF {
		case SkipIfKeyNotPressed:
			c.OpSkipIfKeyNotPressed(opcode)
		case SkipIfKeyIsPressed:
			c.OpSkipIfKeyIsPressed(opcode)
		}

	// Sets VX = DelayTimer
	case 0xF000:
		switch opcode & 0x00FF {
		case GetDelayTimer:
			c.OpGetDelayTimer(opcode)
		case WaitKeypress:
			c.OpWaitKeyPress(opcode)
		case SetDelayTimer:
			c.OpSetDelayTimer(opcode)
		case SetSoundToVx:
			c.OpSetSoundToVX(opcode)
		case AddVXToIndex:
			c.OpAddVXToIndex(opcode)
		case SetIndexToSprite:
			c.OpSetIndexToSprite(opcode)
		case BinaryCodedDecimal:
			c.OpBCD(opcode)
		case StoreRegisters:
			c.OpStoreRegisters(opcode)
		case LoadRegisters:
			c.OpLoadRegisters(opcode)
		}

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

func (c *Chip8) Log(message string) {
	fmt.Println(message)
	c.Logger = append(c.Logger, message)
}
