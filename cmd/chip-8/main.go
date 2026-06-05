package main

import (
	"chip-8/internal/emulator"
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth      = 64
	screenHeight     = 32
	scale            = 10
	IbmLogoPath      = "IBM Logo.ch8"
	Chip8PicturePath = "Chip8 Picture.ch8"
)

type Game struct {
	Chip8         *emulator.Chip8
	State         GameState
	FrameCounter  int
	AvailableROMs []string
	SelectedROM   int
}

type GameState int

const (
	StateStartup GameState = iota
	StateMenu
	StatePlaying
)

func (g *Game) Update() error {
	switch g.State {
	case StateStartup:
		g.DoStartup()
	case StateMenu:
		g.MenuCycle()
	case StatePlaying:
		g.PlayCycle()

	}

	return nil
}

func (g *Game) DoStartup() {
	// Show logo (12 Cycles per frame)
	for i := 0; i < 12; i++ {
		g.Chip8.Cycle()
	}

	g.reduceTimers()
	g.FrameCounter++

	if g.FrameCounter == 80 || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.Chip8.Init()
		g.Chip8.LoadROM(Chip8PicturePath)
		g.FrameCounter = 80
	}

	if g.FrameCounter > 240 || (g.FrameCounter >= 80 && inpututil.IsKeyJustPressed(ebiten.KeyEnter)) {
		g.State = StateMenu
	}
}

func (g *Game) MenuCycle() {
	// Navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		g.SelectedROM = (g.SelectedROM + 1) % len(g.AvailableROMs)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		g.SelectedROM--
		if g.SelectedROM < 0 {
			g.SelectedROM = len(g.AvailableROMs) - 1
		}
	}

	// Load Rom and start game
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.Chip8.Init()
		path := g.AvailableROMs[g.SelectedROM]
		g.Chip8.LoadROM(path)
		g.State = StatePlaying
	}
}

func (g *Game) PlayCycle() {
	// 1. Check for inputs
	g.MapInput()

	// 2. We cycle faster to speed up the cpu
	for i := 0; i < 12; i++ {
		g.Chip8.Cycle()
	}

	g.reduceTimers()
}

func (g *Game) reduceTimers() {
	// Reduce timers
	if g.Chip8.DelayTimer > 0 {
		g.Chip8.DelayTimer--
	}
	if g.Chip8.SoundTimer > 0 {
		g.Chip8.SoundTimer--
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.State {
	case StateStartup:
		g.DrawEmulator(screen)
	case StateMenu:
		g.DrawMenu(screen)
	case StatePlaying:
		g.DrawEmulator(screen)
	}
}

func (g *Game) DrawMenu(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Waehle eine ROM: \n")

	for i, rom := range g.AvailableROMs {
		msg := rom
		if i == g.SelectedROM {
			msg = "> " + rom + " <"
		}
		ebitenutil.DebugPrintAt(screen, msg, 20, 20+i*15)
	}
}

func (g *Game) DrawEmulator(screen *ebiten.Image) {
	for i, pixel := range g.Chip8.GFX {
		if pixel == 1 {
			// calculate x and y from i
			x := (i % emulator.ScreenWidth) * scale
			y := (i / emulator.ScreenWidth) * scale

			// Draw a white 1x1 pixel, this will be scaled up automatically
			vector.FillRect(screen, float32(x), float32(y), scale, scale, color.White, false)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth * scale, screenHeight * scale
}

func main() {
	fmt.Println("Starting up the CHIP-8...")
	emu := &emulator.Chip8{}
	emu.Init()

	ebiten.SetWindowSize(screenWidth*scale, screenHeight*scale)
	ebiten.SetWindowTitle("CHIP-8 Emulator")

	romFIles := []string{"tetris.rom", "Clock Program.ch8", "Pong.ch8"}
	game := &Game{
		Chip8:         emu,
		State:         StateStartup,
		AvailableROMs: romFIles,
	}

	game.Chip8.LoadROM(IbmLogoPath)

	// defer game.SaveLogs()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) MapInput() {
	// Reset keys
	for i := 0; i < 16; i++ {
		g.Chip8.Key[i] = 0
	}

	// check for every possible key
	// Key 1
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.Chip8.Key[0x1] = 1
	}
	// Key 2
	if ebiten.IsKeyPressed(ebiten.Key2) {
		g.Chip8.Key[0x2] = 1
	}
	// Key 3
	if ebiten.IsKeyPressed(ebiten.Key3) {
		g.Chip8.Key[0x3] = 1
	}
	// Key 4
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.Chip8.Key[0xC] = 1
	}

	// Key Q
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.Chip8.Key[0x4] = 1
	}
	// Key W
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.Chip8.Key[0x5] = 1
	}
	// Key E
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.Chip8.Key[0x6] = 1
	}
	// KeyR
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.Chip8.Key[0xD] = 1
	}

	// KeyA
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.Chip8.Key[0x7] = 1
	}
	// Key S
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.Chip8.Key[0x8] = 1
	}
	// Key D
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.Chip8.Key[0x9] = 1
	}
	// Key F
	if ebiten.IsKeyPressed(ebiten.KeyF) {
		g.Chip8.Key[0xE] = 1
	}

	// Key Y
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		g.Chip8.Key[0xA] = 1
	}
	// Key X
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		g.Chip8.Key[0x0] = 1
	}
	// Key C
	if ebiten.IsKeyPressed(ebiten.KeyC) {
		g.Chip8.Key[0xB] = 1
	}
	// Key V
	if ebiten.IsKeyPressed(ebiten.KeyV) {
		g.Chip8.Key[0xF] = 1
	}
}

func (g *Game) SaveLogs() {
	fileName := fmt.Sprintf("log_%s.txt", time.Now().Format("2006-01-02_15-04-05"))
	path := filepath.Join("logs", fileName)

	content := strings.Join(g.Chip8.Logger, "\n")

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		fmt.Println("Error saving the logs:", err)
	} else {
		fmt.Println("Saved logs to:", path)
	}
}
