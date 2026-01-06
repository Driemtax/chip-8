# CHIP-8 Emulator in Go
A [CHIP-8](https://en.wikipedia.org/wiki/CHIP-8) emulator written in Go, using the [Ebitengine](https://ebitengine.org/) for rendering and input handling.

This project was created just for fun and to learn something new.
## Features
- **Opcode Support:** Currently implements 28/35 opcodes of the CHIP-8 standard
- **Graphics:** 64x32 monochrome display rendering.
- **Input:** Keyboard mapping for the original 16-key hex keypad
- **Timers:** Delay and Sound timers (60Hz)
- **Logging:** Optional instruction logging for debugging.
## Getting started
### Prerequisites
- Go 1.18 or higher
### Installation
1. Clone the repository:
  ```bash
    git clone https://github.com/Driemtax/chip-8.git
    cd chip-8
```
2. Install dependencies:
```bash
go mod tidy
```
### Running a ROM
Place your ROM file (e.g., Pong.ch8) in the assets folder.

Run the emulator:
```bash
go run cmd/chip-8/main.go
```

For now you have to set the path to the rom in main.go.

## Controls
The original CHIP-8 used a 16-key-hexadecimal keypad. This emulator maps them to the following keyboard keys:

|CHIP-8 Key | Keyboard Key | Function (Pong) |
|------------|--------------|-----------------|
| 1          | **W**        | Player 1 Up     |
| 4          | **S**        | Player 1 Down   |
| C          | **Arrow Up** | Player 2 Up     |
| D          | **Arrow Down**| Player 2 Down  |

*(Other keys are mapped to 2, 3, Q, E, R, A, D, F, Z, X, C, V)*

## Technical Details
For a deep dive into CHIP-8 architecture, visit [Wikipedia: CHIP-8](https://en.wikipedia.org/wiki/CHIP-8).
### Supported Opcodes
This emulator implements the standard CHIP-8 instruction set, where every instruction is 2-Bytes of length:

X,Y $=$ 4-Bit Register Index 
N $=$ 4-Bit Constant
NN $=$ 8-bit Constant
NNN $=$ 12-bit memory address 
I $=$ 16-bit Index register

| Opcode | Assembly | Description |
|------------|--------------|-----------------|
| 00E0 | CLS | Clear the display. |
| 00EE | RET | Returns from a subroutine. |
| 1NNN | JP addr | Jump to adress NNN. | 
| 2NNN | CALL addr | Call subroutine at NNN. |
| 3XNN | SE Vx, byte | Skip next instruction if Vx $==$ NN |
| 4XNN | SNE Vx, byte | Skip next instruction if Vx $\not=$ NN |
| 5XY0 | SE Vx, Vy | Skip next instruction if Vx $==$ Vy |
| 6XNN | LD Vx, byte | Set Vx $=$ NN |
| 7XNN | ADD Vx, byte | Set Vx $=$ Vx $+$ NN |
| 8XY0 | LD Vx, Vy | Set Vx $=$ Vy |
| 8XY1 | OR Vx, Vy | Set Vx $=$ Vx OR Vy |
| 8XY2 | AND Vx, Vy | Set Vx $=$ Vx AND Vy |
| 8XY3 | XOR Vx, Vy | Set Vx $=$ Vx XOR Vy |
| 8XY4 | ADD Vx, Vy | Set Vx $=$ Vx $+$ Vy, set VF $=$ carry. |
| 8XY5 | SUB Vx, Vy | Set Vx $=$ Vx $-$ Vy, set VF NOT borrow. |
| 8XY6 | SHR Vx | Set Vx $=$ Vx >> 1 |
| 8XY7 | SUBN Vx, Vy | Set Vx $=$ Vy - Vx, set VF NOT borrow. |
| 8XYE | SHL Vx | Set Vx $=$ Vx << 1 |
| 9XY0 | SNE Vx, Vy | Skips the next instruction if VX does not equal VY.  |
| ANNN | LD I, addr | Set I $=$ NNN |
| BNNN | JP V0, addr | Jump to location NNN $+$ V0 |
| CXNN | RND Vx, byte | Set Vx $=$ random byte AND NN |
| DXYN | DRW Vx, Vy, nibble | Display n-byte sprite starting at memory location I at (Vx, Vy), set VF $=$ collision |
| EX9E | SKP Vx | Skip next instruction if key with the value of Vx is pressed. |
| EXA1 | SKNP Vx | Skip next instruction if key with the value of Vx is NOT pressed. |
| FX07 | LD Vx, DT | Set Vx $=$ delay timer value. |
| FX0A | LD Vx, K | A key press is awaited, and then stored in VX (blocking operation, all instruction halted until next key event, delay and sound timers should continue processing) |
| FX15 | LD DT, Vx | Set delay timer $=$ Vx |
| FX18 | ST, Vx | Sets the sound timer to VX. |
| FX1E | ADD I, Vx | Adds VX to I. VF is not affected. |
| FX29 | LD F, Vx | Set I $=$ location of sprite for digit in Vx |
| FX33 | LD B, Vx | Store BinaryCodedDecimal representation of Vx in memory locations I, I+1, I+2 |
| FX55 | LD[I], Vx | Store registers V0 through Vx in memory starting at location I |
| FX65 | LD Vx, [I] | Read registers V0 through Vx from memory starting at location I. |

## License 
This project is open source and available under the [MIT License]().
