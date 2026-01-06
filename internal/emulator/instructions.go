package emulator

import (
	"fmt"
	"math/rand"
	"slices"
)

const (
	ClearScreen           uint16 = 0x00E0
	ReturnFromSubroutine  uint16 = 0x00EE
	Jump                  uint16 = 0x1000
	CallSubroutine        uint16 = 0x2000
	SkipIfEquals          uint16 = 0x3000
	SkipIfNotEquals       uint16 = 0x4000
	SkipIfVXEQVY          uint16 = 0x5000
	SetVX                 uint16 = 0x6000
	AddVX                 uint16 = 0x7000
	SetVXtoVY             uint16 = 0x8000
	BitwiseOR             uint16 = 0x8001
	BitwiseAND            uint16 = 0x8002
	BitwiseXOR            uint16 = 0x8003
	Add                   uint16 = 0x8004
	Sub                   uint16 = 0x8005
	RightShift            uint16 = 0x8006
	SubReverse            uint16 = 0x8007
	LeftShift             uint16 = 0x800E
	SkipRegistersNotEqual uint16 = 0x9000
	SetIndex              uint16 = 0xA000
	JumpPlusV0            uint16 = 0xB000
	Random                uint16 = 0xC000
	DrawSprite            uint16 = 0xD000
	SkipIfKeyNotPressed   uint16 = 0xE0A1
	SkipIfKeyIsPressed    uint16 = 0xE09E
	GetDelayTimer         uint16 = 0x07
	WaitKeypress          uint16 = 0x0A
	SetDelayTimer         uint16 = 0x15
	SetSoundToVx          uint16 = 0x18
	AddVXToIndex          uint16 = 0x1E
	SetIndexToSprite      uint16 = 0x29
	BinaryCodedDecimal    uint16 = 0x33
	StoreRegisters        uint16 = 0x55
	LoadRegisters         uint16 = 0x65
)

// 00E0 --- Clears the screen
func (c *Chip8) OpClearScreen() {
	c.GFX = [ScreenWidth * ScreenHeight]uint8{0}
	message := fmt.Sprintln("CLS - Clear Screen")
	c.Log(message)
}

// OOEE --- RET Returns from the current subroutine
func (c *Chip8) OpReturnFromSubroutine() {
	// 1. Reduce Stack Pointer to get to the address before subroutine was called
	c.SP--
	// 2. Reset the Programm Counter to that address
	c.PC = c.Stack[c.SP]

	message := fmt.Sprintln("RET to", ToHex(c.PC))
	c.Log(message)
}

// 1NNN --- Jumps to address NNN.
func (c *Chip8) OpJump(opcode uint16) {
	nnn := extractNNN(opcode)
	c.PC = nnn

	if nnn != 0x0228 {
		message := fmt.Sprintf("JMP %s\n", ToHex(nnn))
		c.Log(message)
	}
}

// 2NNN --- Calls a subroutine at adress NNN
func (c *Chip8) OpCallSubroutine(opcode uint16) {
	nnn := extractNNN(opcode)
	// 1. Save current Programm Counter on the Stack, this is the return address
	c.Stack[c.SP] = c.PC
	// 2. increase Stack Pointer to not overwrite current address on next call
	c.SP++
	// 3. Set Programm Counter to Address NNN, this is the actual CALL
	c.PC = nnn

	message := fmt.Sprintln("CALL", ToHex(nnn))
	c.Log(message)
}

// 3XNN --- Skips the next instruction if VX == NN
func (c *Chip8) OpSkipIfEqulas(opcode uint16) {
	x := extractX(opcode)
	nn := extractNN(opcode)

	if c.V[x] == nn {
		// Remember that one instruction is 2 Bytes long!
		c.PC += 2
		message := fmt.Sprintf("IFEQ %X, %X\n", c.V[x], nn)
		c.Log(message)
	} else {
		message := fmt.Sprintf("PASS since %X != %X\n", c.V[x], nn)
		c.Log(message)
	}
}

// 4XNN --- Skips the next instruction if VX != NN
func (c *Chip8) OpSkipIfNotEqual(opcode uint16) {
	x := extractX(opcode)
	nn := extractNN(opcode)

	if c.V[x] != nn {
		// Remember that one instruction is 2 Bytes long!
		c.PC += 2
		message := fmt.Sprintf("IFNEQ %X, %X\n", c.V[x], nn)
		c.Log(message)
	} else {
		message := fmt.Sprintf("PASS since %X == %X\n", c.V[x], nn)
		c.Log(message)
	}
}

// 5XY0 --- Skips the next instruction if VX == VY
func (c *Chip8) OpSkipVXEQVY(opcode uint16) {
	if opcode&0x000F != 0 {
		message := fmt.Sprintln("Invalid opcode:", ToHex(opcode))
		c.Log(message)
		return
	}

	x := extractX(opcode)
	y := extractY(opcode)

	if c.V[x] == c.V[y] {
		// Remember that one instruction is 2 Bytes long!
		c.PC += 2
		message := fmt.Sprintf("IFXEQY %X, %Y\n", c.V[x], c.V[y])
		c.Log(message)
	} else {
		message := fmt.Sprintf("PASS since %X != %X\n", c.V[x], c.V[y])
		c.Log(message)
	}
}

// 6XNN --- Sets register VX = NN
func (c *Chip8) OpSetRegister(opcode uint16) {
	x := extractX(opcode)
	nn := extractNN(opcode)
	c.V[x] = nn
	message := fmt.Sprintf("LD V%X, %X\n", x, nn)
	c.Log(message)
}

// 7XNN --- Adds NN to VX (carry flag is not changed).
func (c *Chip8) OpAddToRegister(opcode uint16) {
	x := extractX(opcode)
	nn := extractNN(opcode)
	c.V[x] += nn
	message := fmt.Sprintf("ADD V%X, %X\n", x, nn)
	c.Log(message)
}

// 8XY0 --- Sets VX to the value of VY
func (c *Chip8) OpSetVXtoVY(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VY := c.V[y]
	c.V[x] = VY

	message := fmt.Sprintf("LD V%X, V%X (%X)\n", x, y, VY)
	c.Log(message)
}

// 8XY1 --- Bitwise OR of VX | VY. The result is stored in VX
func (c *Chip8) OpBitwiseOR(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]
	c.V[x] = VX | VY

	message := fmt.Sprintf("OR %X, %X\n", VX, VY)
	c.Log(message)
}

// 8XY2 --- Bitwise AND of VX & VY. The result is stored in VX
func (c *Chip8) OpBitwiseAND(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]
	c.V[x] = VX & VY

	message := fmt.Sprintf("AND %X, %X\n", VX, VY)
	c.Log(message)
}

// 8XY3 --- Bitwise XOR of VX ^ VY. The result is stored in VX
func (c *Chip8) OpBitwiseXOR(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]
	c.V[x] = VX ^ VY

	message := fmt.Sprintf("XOR %X, %X\n", VX, VY)
	c.Log(message)
}

// 8XY4 --- Adds VY to VX. VF is set to 1 when there's an overflow, and to 0 when there is not
func (c *Chip8) OpAdd(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]

	result := uint16(VX) + uint16(VY)
	// if there is an overflow, VF is set to 1, otherwise to 0
	var borrow uint8 = 0
	if result > 255 {
		borrow = 1
	}
	c.V[x] = uint8(result)
	c.V[0xF] = borrow

	message := fmt.Sprintf("ADD %X, %X\n", VX, VY)
	c.Log(message)
}

// 8XY5 --- VY is subtracted from VX. VF is set to 0 when there's an underflow, and 1 when there is not. (i.e. VF set to 1 if VX >= VY and 0 if not).
func (c *Chip8) OpSub(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]

	result := VX - VY

	// if there is an underflow, VF is set to 0, otherwise to 1
	var notBorrow uint8 = 1
	if VY >= VX {
		notBorrow = 0
	}
	c.V[x] = result
	c.V[0xF] = notBorrow

	message := fmt.Sprintf("SUB %X, %X\n", VX, VY)
	c.Log(message)
}

// 8XY6 --- Shifts VX to the right by 1, then stores the least significant bit of VX prior to the shift into VF.
func (c *Chip8) OpRightShift(opcode uint16) {
	x := extractX(opcode)
	VX := c.V[x]
	lsb := VX & 0x01
	c.V[15] = lsb
	c.V[x] = VX >> 1

	message := fmt.Sprintf("SHR V%X (%X)\n", x, VX)
	c.Log(message)
}

// 8XY7 --- Sets VX to VY minus VX.
// VF is set to 0 when there's an underflow, and 1 when there is not. (i.e. VF set to 1 if VY >= VX).
func (c *Chip8) OpSubReverse(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]

	result := VY - VX

	var notBorrow uint8 = 1
	// check for underflow
	if VX >= VY {
		notBorrow = 0
	}

	c.V[x] = result
	c.V[0xF] = notBorrow

	message := fmt.Sprintf("SUBN %X, %X\n", VY, VX)
	c.Log(message)
}

// 8XYE --- Shifts VX to the left by 1, then sets VF to 1 if the most significant bit of VX prior to that shift was set, or to 0 if it was unset.
func (c *Chip8) OpLeftShift(opcode uint16) {
	x := extractX(opcode)
	VX := c.V[x]
	msb := (VX & 0x80) >> 7 // 0x80 = 0b1000000
	c.V[15] = msb
	c.V[x] = VX << 1

	message := fmt.Sprintf("SHL V%X (%X)\n", x, VX)
	c.Log(message)
}

// 9XY0 --- Skips the next instruction if VX does not equal VY. (Usually the next instruction is a jump to skip a code block).
func (c *Chip8) OpSkipRegistersNotEqual(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]

	if VX != VY {
		c.PC += 2
		message := fmt.Sprintf("SNE V%X, V%Y\n", x, y)
		c.Log(message)
	}
}

// ANNN --- Sets I to the address NNN.
func (c *Chip8) OpSetIndexRegister(opcode uint16) {
	nnn := extractNNN(opcode)
	c.I = nnn
	message := fmt.Sprintf("MOV I, %s\n", ToHex(nnn))
	c.Log(message)
}

// BNNN --- Jumps to the address NNN plus V0.
func (c *Chip8) OpJumpPlusV0(opcode uint16) {
	nnn := extractNNN(opcode)
	c.PC = nnn + uint16(c.V[0])

	message := fmt.Sprintf("JMP V0 (%X), %s\n", c.V[0], ToHex(nnn))
	c.Log(message)
}

// CXNN --- Set VX = random byte AND NN
func (c *Chip8) OpRandom(opcode uint16) {
	x := extractX(opcode)
	nn := extractNN(opcode)

	randomByte := uint8(rand.Intn(256))

	c.V[x] = randomByte & nn

	message := fmt.Sprintf("RND V%X, %X (Result: %X)\n", x, nn, c.V[x])
	c.Log(message)
}

// DXYN --- Draw Sprite from VX to VY with Width = 8 and Height = N
func (c *Chip8) OpDraw(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	height := extractN(opcode)

	vx := c.V[x]
	vy := c.V[y]

	// Need to reset VF because?
	c.V[0xF] = 0

	// iterate over every row of the sprite (Y-axis)
	for row := uint16(0); row < height; row++ {
		// get the row byte
		spriteByte := c.Ram[c.I+row]

		for col := uint16(0); col < 8; col++ {
			// check if the bot at pos col is 1
			// 0x80 = 0b1000000.
			if (spriteByte & (0x80 >> col)) != 0 {
				// calculate position on screen
				// using modulo for wrapping the lines
				screenX := (uint16(vx) + col) //% ScreenWidth
				screenY := (uint16(vy) + row) //% ScreenHeight

				// Clipping
				if screenX >= ScreenWidth || screenY >= ScreenHeight {
					continue
				}

				// calculate index in GFX (1D-Araay)
				// i = y*w+x
				index := screenY*ScreenWidth + screenX

				// if the pixel is already set there is a collision
				if c.GFX[index] == 1 {
					c.V[0xF] = 1
				}

				// What does this line even do? Why do we need it?
				c.GFX[index] ^= 1
			}
		}
	}

	message := fmt.Sprintf("DRAW V%X(%d), V%X(%d), H:%d\n", x, vx, y, vy, height)
	c.Log(message)
}

// EX9E --- Skips the next instruction if the key stored in VX(only consider the lowest nibble)
// is pressed (usually the next instruction is a jump to skip a code block).
func (c *Chip8) OpSkipIfKeyIsPressed(opcode uint16) {
	x := extractX(opcode)
	keyIndex := c.V[x] & 0x0F

	if c.Key[keyIndex] == 1 {
		c.PC += 2
		message := fmt.Sprintf("SKIP KEY PRESS V%X (%X)\n", x, keyIndex)
		c.Log(message)
	} else {
		message := fmt.Sprintf("NO SKIP KEY PRESS V%X (%X)\n", x, keyIndex)
		c.Log(message)
	}
}

// EXA1 --- Skips the next instruction if the key stored in VX(only consider the lowest nibble)
// is not pressed (usually the next instruction is a jump to skip a code block).
func (c *Chip8) OpSkipIfKeyNotPressed(opcode uint16) {
	x := extractX(opcode)
	keyIndex := c.V[x] & 0x0F

	if c.Key[keyIndex] == 0 {
		c.PC += 2
		message := fmt.Sprintf("SKIP KEY NOT PRESSED V%X (%X)\n", x, keyIndex)
		c.Log(message)
	} else {
		message := fmt.Sprintf("NO SKIP KEY NOT PRESSED V%X (%X)\n", x, keyIndex)
		c.Log(message)
	}
}

// FX07 --- Gets the Delay Timer and stores the result in VX
func (c *Chip8) OpGetDelayTimer(opcode uint16) {
	x := extractX(opcode)
	c.V[x] = c.DelayTimer

	message := fmt.Sprintf("LD V%X, %X\n", x, c.DelayTimer)
	c.Log(message)
}

// FX0A --- A key press is awaited, and then stored in VX
// (blocking operation, all instruction halted until next key event, delay and sound timers should continue processing)
func (c *Chip8) OpWaitKeyPress(opcode uint16) {
	// If a key is pressed, just continue normally
	key := c.isKeyPressed()

	// Case no key is pressed currently
	if key == -1 {
		// Was there a key pressed before?
		if c.KeyPressedBuffer != 255 {
			x := extractX(opcode)
			c.V[x] = c.KeyPressedBuffer

			message := fmt.Sprintf("KEY RELEASED V%X, %X\n", x, c.KeyPressedBuffer)
			c.Log(message)

			c.KeyPressedBuffer = 255 // Reset Buffer, IMPORTANT!!
			return                   // Get on with the next instruction
		}

		// No key was remembered in the buffer, means we still wait for the first keypress
		c.PC -= 2 // This line will effect, that the instruction we fetch is this one again until one key was pressed.
		// message := fmt.Sprintf("WAIT KEY...\n")
		// c.Log(message)
		return
	}

	// Case there is a key pressed currently
	c.KeyPressedBuffer = uint8(key)
	// Reset PC because we wait until the key is released
	c.PC -= 2

}

// FX15 --- Sets the Delay Timer to the Value in VX
func (c *Chip8) OpSetDelayTimer(opcode uint16) {
	x := extractX(opcode)
	c.DelayTimer = c.V[x]

	message := fmt.Sprintf("LD DT, V%X (%X)\n", x, c.DelayTimer)
	c.Log(message)
}

// FX18 --- Sets the sound timer to VX.
func (c *Chip8) OpSetSoundToVX(opcode uint16) {
	x := extractX(opcode)
	c.SoundTimer = c.V[x]

	message := fmt.Sprintf("LD SD, V%X (%X)\n", x, c.V[x])
	c.Log(message)
}

// FX1E --- Adds VX to I. VF is not affected.
func (c *Chip8) OpAddVXToIndex(opcode uint16) {
	x := extractX(opcode)
	c.I += uint16(c.V[x])

	message := fmt.Sprintf("ADD I, V%X (%X)\n", x, c.V[x])
	c.Log(message)
}

// FX29 --- Sets the Index Register I to the location of the sprite for the character in VX
// Makes it possible to print a character of the fontSet
func (c *Chip8) OpSetIndexToSprite(opcode uint16) {
	x := extractX(opcode)
	c.I = uint16(c.V[x]) * 5 // One Character in the font is 5 Bytes long, so we need ot offset

	message := fmt.Sprintf("LD I, V%X (Sprite für %X)\n", x, c.I)
	c.Log(message)
}

// FX33 --- Store BCD representation of VX in memory locations I, I+1, and I+2
func (c *Chip8) OpBCD(opcode uint16) {
	x := extractX(opcode)
	value := c.V[x]

	// 100 position
	c.Ram[c.I] = value / 100
	// 10 position
	c.Ram[c.I+1] = (value / 10) % 10
	// 1 position
	c.Ram[c.I+2] = value % 10

	message := fmt.Sprintf("BCD V%X (%d) -> Ram[%X...%X]\n", x, value, c.I, c.I+2)
	c.Log(message)
}

// FX55 --- Store registers V0 through VX in memory starting at location I
func (c *Chip8) OpStoreRegisters(opcode uint16) {
	x := extractX(opcode)
	for i := uint16(0); i <= x; i++ {
		c.Ram[c.I+i] = c.V[i]
	}

	message := fmt.Sprintf("DUMP V0...V%X to Ram[%X]\n", x, c.I)
	c.Log(message)
}

// FX65 --- Read Registers V0 through VX from memory starting at location I
func (c *Chip8) OpLoadRegisters(opcode uint16) {
	x := extractX(opcode)
	for i := uint16(0); i <= x; i++ {
		c.V[i] = c.Ram[c.I+i]
	}

	message := fmt.Sprintf("LOAD V0...V%X from Ram[%X]\n", x, c.I)
	c.Log(message)
}

func extractX(opcode uint16) uint16 {
	return (opcode & 0x0F00) >> 8
}

func extractY(opcode uint16) uint16 {
	return (opcode & 0x00F0) >> 4
}

func extractN(opcode uint16) uint16 {
	return opcode & 0x000F
}

// This function returns uint8 since NN is always 8 Bytes and often is used as uint8.
// We just save us a lot of type casting later on.
func extractNN(opcode uint16) uint8 {
	return uint8(opcode & 0x00FF)
}

func extractNNN(opcode uint16) uint16 {
	return opcode & 0x0FFF
}

// Checks wether any key is pressed or not. Returns the key if pressend, -1 if no key was pressed
func (c *Chip8) isKeyPressed() int {
	index := slices.Index(c.Key[:], uint8(1))

	return index
}
