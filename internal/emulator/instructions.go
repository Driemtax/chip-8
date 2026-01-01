package emulator

import "fmt"

const (
	ClearScreen          uint16 = 0x00E0
	ReturnFromSubroutine uint16 = 0x00EE
	Jump                 uint16 = 0x1000
)

func (c *Chip8) OpClearScreen() {
	c.GFX = [ScreenWidth * ScreenHeight]uint8{0}
	fmt.Println("CLS - Clear Screen")
}

func (c *Chip8) OpJump(opcode uint16) {
	nnn := opcode & 0x0FFF
	c.PC = nnn

	if nnn != 0x0228 {
		fmt.Printf("JMP %s\n", ToHex(nnn))
	}
}

func (c *Chip8) OpSetIndexRegister(opcode uint16) {
	nnn := opcode & 0x0FFF
	c.I = nnn
	fmt.Printf("MOV I, %s\n", ToHex(nnn))
}

// 6XNN --- Sets register VX = NN
func (c *Chip8) OpSetRegister(opcode uint16) {
	x := (opcode & 0x0F00) >> 8
	nn := uint8(opcode & 0x00FF)
	c.V[x] = nn
	fmt.Printf("LD V%X, %X\n", x, nn)
}

func (c *Chip8) OpAddToRegister(opcode uint16) {
	x := (opcode & 0x0F00) >> 8
	nn := uint8(opcode & 0x00FF)
	c.V[x] += nn
	fmt.Printf("ADD V%X, %X\n", x, nn)
}

// DXYN --- Draw Sprite from VX to VY with Width = 8 and Height = N
func (c *Chip8) OpDraw(opcode uint16) {
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	height := opcode & 0x000F

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
