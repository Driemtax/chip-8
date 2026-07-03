# CHIP-8 Emulator in Go

A basic [CHIP-8](https://en.wikipedia.org/wiki/CHIP-8) emulator written in Go. Uses [Ebitengine](https://ebitengine.org/) for rendering and input handling.

This project was created purely for fun and to learn more about emulators and low-level programming.

https://github.com/user-attachments/assets/1726c0d5-185c-4fbf-a31d-e10c3ff9f170

## Features
- **Opcode Support:** Full implementation of the standard CHIP-8 instruction set.
- **Graphics & Splash:** Accurate 64x32 monochrome display rendering with custom boot-up animations.
- **ROM Launcher:** Simple built-in menu to select and load `.ch8` or `.rom` files from your `assets/` folder.
- **Timers:** Proper 60Hz Delay and Sound timers.
- **Logging:** Optional instruction logging for debugging purposes.
## Getting started
### Pre-built Binarries
You don't need to compile the project yourself. I provide pre-built binaries for **Windows** and **Linux (Ubuntu)** in the [Releases section](../../releases) on GitHub. Just download the executable, make sure there is an `assets` folder with ROMs next to it, and run it!
### Compile from source
**Requirements**:
- Go 1.18 or higher

**Installation**
1. Clone the repository:
  ```bash
    git clone https://github.com/Driemtax/chip-8.git
    cd chip-8
```
2. Install dependencies:
```bash
go mod tidy
```

Place your ROM files (e.g., Pong.ch8) in the assets folder.

Run the emulator:
```bash
go run cmd/chip-8/main.go
```

## Controls
The original CHIP-8 used a 16-key-hexadecimal keypad. This emulator maps them to the following keyboard keys:

1 2 3 4    maps to    1 2 3 C
Q W E R    maps to    4 5 6 D
A S D F    maps to    7 8 9 E
Y X C V    maps to    A 0 B F 

| CHIP-8 Key | Keyboard Key | 
|------------|--------------|
| 1          | **1**        |
| 2          | **2**        |
| 3          | **3**        | 
| C          | **4**        |
| 4          | **Q**        |
| 5          | **W**        |
| 6          | **E**        |
| D          | **R**        |
| 7          | **A**        |
| 8          | **S**        |
| 9          | **D**        |
| E          | **F**        | 
| A          | **Y**        |
| 0          | **X**        |
| B          | **C**        |
| F          | **V**        |

Every ROM has its controlls hardcoded and therefore cannot be configured by the emulator. For example Pong has the following controlls:

| Instruction | CHIP-8 Key | Keyboard Key |
|-------------|------------|--------------|
| P1 UP       | **1**      | 1            |
| P1 DOWN     | **4**      | Q            |
| P2 UP       | **C**      | 4            |
| P2 DOWN     | **D**      | R            |

For the other games and roms you have to test and discover yourself, like in the old times.

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
This project is open source and available under the [MIT License](https://opensource.org/license/mit).
