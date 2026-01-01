package main

import (
	"chip-8/internal/emulator"
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 64
	screenHeight = 32
	scale        = 10
)

type Game struct {
	Chip8 *emulator.Chip8
}

func (g *Game) Update() error {
	// TODO: Execute more then 1 Cycle per Frame for games to be fast enough
	g.Chip8.Cycle()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for i, pixel := range g.Chip8.GFX {
		if pixel == 1 {
			// calculate x and y from i
			x := i % emulator.ScreenWidth
			y := i / emulator.ScreenWidth

			// Draw a white 1x1 pixel, this will be scaled up automatically
			vector.FillRect(screen, float32(x), float32(y), 1, 1, color.White, false)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	fmt.Println("Starting up the CHIP-8...")
	emu := &emulator.Chip8{}
	emu.Init()
	path := "pong.rom"
	err := emu.LoadROM(path)
	if err != nil {
		log.Fatal("Error loading ROM:", err)
	}

	ebiten.SetWindowSize(screenWidth*scale, screenHeight*scale)
	ebiten.SetWindowTitle("CHIP-8 Emulator")

	game := &Game{Chip8: emu}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
