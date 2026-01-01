package emulator

import "fmt"

const (
	ClearScreen          uint16 = 0x00E0
	ReturnFromSubroutine uint16 = 0x00EE
	Jump                 uint16 = 0x1000
	CallSubroutine       uint16 = 0x2000
	SkipIfEquals         uint16 = 0x3000
	SkipIfNotEquals      uint16 = 0x4000
	SkipIfVXEQVY         uint16 = 0x5000
	SetVX                uint16 = 0x6000
	AddVX                uint16 = 0x7000
	BitwiseOR            uint16 = 0x8001
	BitwiseAND           uint16 = 0x8002
	BitwiseXOR           uint16 = 0x8003
	Add                  uint16 = 0x8004
	Sub                  uint16 = 0x8005
	RightShift           uint16 = 0x8006
	SubReverse           uint16 = 0x8007
	LeftShift            uint16 = 0x800E
	DrawSprite           uint16 = 0xD000
	SkipIfKeyNotPressed  uint16 = 0xE0A1
	SkipIfKeyIsPressed   uint16 = 0xE09E
	GetDelayTimer        uint16 = 0x07
	SetDelayTimer        uint16 = 0x15
	SetIndexToSprite     uint16 = 0x29
)

func (c *Chip8) OpClearScreen() {
	c.GFX = [ScreenWidth * ScreenHeight]uint8{0}
	fmt.Println("CLS - Clear Screen")
}

// OOEE --- RET Returns from the current subroutine
func (c *Chip8) OpReturnFromSubroutine() {
	// 1. Reduce Stack Pointer to get to the address before subroutine was called
	c.SP--
	// 2. Reset the Programm Counter to that address
	c.PC = c.Stack[c.SP]
	fmt.Println("RET to", ToHex(c.PC))
}

func (c *Chip8) OpJump(opcode uint16) {
	nnn := extractNNN(opcode)
	c.PC = nnn

	if nnn != 0x0228 {
		fmt.Printf("JMP %s\n", ToHex(nnn))
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
	fmt.Println("CALL", ToHex(nnn))
}

// 3XNN --- Skips the next instruction if VX == NN
func (c *Chip8) OpSkipIfEqulas(opcode uint16) {
	x := extractX(opcode)
	nn := extractNN(opcode)

	if c.V[x] == nn {
		// Remember that one instruction is 2 Bytes long!
		c.PC += 2
		fmt.Printf("IFEQ %X, %X\n", c.V[x], nn)
	} else {
		fmt.Printf("PASS since %X != %X\n", c.V[x], nn)
	}
}

// 4XNN --- Skips the next instruction if VX != NN
func (c *Chip8) OpSkipIfNotEqual(opcode uint16) {
	x := extractX(opcode)
	nn := extractNN(opcode)

	if c.V[x] != nn {
		// Remember that one instruction is 2 Bytes long!
		c.PC += 2
		fmt.Printf("IFNEQ %X, %X\n", c.V[x], nn)
	} else {
		fmt.Printf("PASS since %X == %X\n", c.V[x], nn)
	}
}

// 5XY0 --- Skips the next instruction if VX == VY
func (c *Chip8) OpSkipVXEQVY(opcode uint16) {
	if opcode&0x000F != 0 {
		fmt.Println("Invalid opcode:", ToHex(opcode))
		return
	}

	x := extractX(opcode)
	y := extractY(opcode)

	if c.V[x] == c.V[y] {
		// Remember that one instruction is 2 Bytes long!
		c.PC += 2
		fmt.Printf("IFXEQY %X, %Y\n", c.V[x], c.V[y])
	} else {
		fmt.Printf("PASS since %X != %X\n", c.V[x], c.V[y])
	}
}

func (c *Chip8) OpSetIndexRegister(opcode uint16) {
	nnn := extractNNN(opcode)
	c.I = nnn
	fmt.Printf("MOV I, %s\n", ToHex(nnn))
}

// 6XNN --- Sets register VX = NN
func (c *Chip8) OpSetRegister(opcode uint16) {
	x := extractX(opcode)
	nn := extractNN(opcode)
	c.V[x] = nn
	fmt.Printf("LD V%X, %X\n", x, nn)
}

func (c *Chip8) OpAddToRegister(opcode uint16) {
	x := extractX(opcode)
	nn := extractNN(opcode)
	c.V[x] += nn
	fmt.Printf("ADD V%X, %X\n", x, nn)
}

// 8XY1 --- Bitwise OR of VX | VY. The result is stored in VX
func (c *Chip8) OpBitwiseOR(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]
	c.V[x] = VX | VY

	fmt.Printf("OR %X, %X\n", VX, VY)
}

// 8XY2 --- Bitwise AND of VX & VY. The result is stored in VX
func (c *Chip8) OpBitwiseAND(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]
	c.V[x] = VX & VY

	fmt.Printf("AND %X, %X\n", VX, VY)
}

// 8XY3 --- Bitwise XOR of VX ^ VY. The result is stored in VX
func (c *Chip8) OpBitwiseXOR(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]
	c.V[x] = VX ^ VY

	fmt.Printf("XOR %X, %X\n", VX, VY)
}

// 8XY4 --- Adds VY to VX. VF is set to 1 when there's an overflow, and to 0 when there is not
func (c *Chip8) OpAdd(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]

	result := uint16(VX) + uint16(VY)
	// if there is an overflow, VF is set to 1, otherwise to 0
	if result > 255 {
		c.V[15] = 1
	} else {
		c.V[15] = 0
	}
	c.V[x] = VX + VY

	fmt.Printf("ADD %X, %X\n", VX, VY)
}

// 8XY5 --- VY is subtracted from VX. VF is set to 0 when there's an underflow, and 1 when there is not. (i.e. VF set to 1 if VX >= VY and 0 if not).
func (c *Chip8) OpSub(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]

	// if there is an underflow, VF is set to 0, otherwise to 1
	if VY > VX {
		c.V[15] = 0
	} else {
		c.V[15] = 1
	}
	c.V[x] = VX - VY

	fmt.Printf("SUB %X, %X\n", VX, VY)
}

// 8XY6 --- Shifts VX to the right by 1, then stores the least significant bit of VX prior to the shift into VF.
func (c *Chip8) OpRightShift(opcode uint16) {
	x := extractX(opcode)
	VX := c.V[x]
	lsb := VX & 0x01
	c.V[15] = lsb
	c.V[x] = VX >> 1

	fmt.Printf("SHR V%X (%X)\n", x, VX)
}

// 8XY7 --- Sets VX to VY minus VX.
// VF is set to 0 when there's an underflow, and 1 when there is not. (i.e. VF set to 1 if VY >= VX).
func (c *Chip8) OpSubReverse(opcode uint16) {
	x := extractX(opcode)
	y := extractY(opcode)
	VX := c.V[x]
	VY := c.V[y]

	// check for underflow
	if VX > VY {
		c.V[15] = 0
	} else {
		c.V[15] = 1
	}

	c.V[x] = VY - VX

	fmt.Printf("SUBN %X, %X\n", VY, VX)
}

// 8XYE --- Shifts VX to the left by 1, then sets VF to 1 if the most significant bit of VX prior to that shift was set, or to 0 if it was unset.
func (c *Chip8) OpLeftShift(opcode uint16) {
	x := extractX(opcode)
	VX := c.V[x]
	msb := (VX & 0x80) >> 7 // 0x80 = 0b1000000
	c.V[15] = msb
	c.V[x] = VX << 1

	fmt.Printf("SHL V%X (%X)\n", x, VX)
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
				screenX := (uint16(vx) + col) % ScreenWidth
				screenY := (uint16(vy) + row) % ScreenHeight

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

	fmt.Printf("DRAW V%X(%d), V%X(%d), H:%d\n", x, vx, y, vy, height)
}

// EX9E --- Skips the next instruction if the key stored in VX(only consider the lowest nibble)
// is pressed (usually the next instruction is a jump to skip a code block).
func (c *Chip8) OpSkipIfKeyIsPressed(opcode uint16) {

}

// EXA1 --- Skips the next instruction if the key stored in VX(only consider the lowest nibble)
// is not pressed (usually the next instruction is a jump to skip a code block).
func (c *Chip8) OpSkipIfKeyNotPressed(opcode uint16) {

}

// FX07 --- Gets the Delay Timer and stores the result in VX
func (c *Chip8) OpGetDelayTimer(opcode uint16) {
	x := extractX(opcode)
	c.V[x] = c.DelayTimer
	fmt.Printf("LD V%X, %X\n", x, c.DelayTimer)
}

// FX15 --- Sets the Delay Timer to the Value in VX
func (c *Chip8) OpSetDelayTimer(opcode uint16) {
	x := extractX(opcode)
	c.DelayTimer = c.V[x]
	fmt.Printf("LD DT, V%X (%X)\n", x, c.DelayTimer)
}

// FX29 --- Sets the Index Register I to the location of the sprite for the character in VX
// Makes it possible to print a character of the fontSet
func (c *Chip8) OpSetIndexToSprite(opcode uint16) {
	x := extractX(opcode)
	c.I = uint16(c.V[x]) * 5 // One Character in the font is 5 Bytes long, so we need ot offset
	fmt.Printf("LD I, V%X (Sprite für %X)\n", x, c.I)
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
