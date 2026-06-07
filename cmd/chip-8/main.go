package main

import (
	"bytes"
	"chip-8/internal/emulator"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth             = 64
	screenHeight            = 32
	scale                   = 10
	IbmLogoPath             = "IBM UltraFade.ch8"
	Chip8PicturePath        = "Chip8 Picture.ch8"
	Chip8ExitLogoPath       = "Chip8 emulator Logo.ch8"
	startupStage1FrameCount = 100
	startupStage2FrameCount = startupStage1FrameCount + 120
)

type Game struct {
	Chip8         *emulator.Chip8
	State         GameState
	FrameCounter  int
	AvailableROMs []string
	SelectedROM   int

	// Audio
	audioContext *audio.Context
	beepPlayer   *audio.Player
}

type GameState int

const (
	StateStartup GameState = iota
	StateMenu
	StatePlaying
	StateShutdown
	StateExit
)

func (g *Game) Update() error {
	switch g.State {
	case StateStartup:
		g.DoStartup()
	case StateMenu:
		g.MenuCycle()
	case StatePlaying:
		g.PlayCycle()
	case StateShutdown:
		g.Shutdown()
	case StateExit:
		os.Exit(0)
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

	if g.FrameCounter == startupStage1FrameCount || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.Chip8.Init()
		g.Chip8.LoadROM(Chip8PicturePath)
		g.FrameCounter = startupStage1FrameCount
	}

	if g.FrameCounter > startupStage2FrameCount || (g.FrameCounter >= startupStage1FrameCount && inpututil.IsKeyJustPressed(ebiten.KeyEnter)) {
		g.State = StateMenu
	}

	// if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
	// 	g.SaveLogs()
	// 	g.State = StateExit
	// }
}

func (g *Game) MenuCycle() {
	// Shutdown
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.Chip8.Init()
		g.Chip8.LoadROM(Chip8ExitLogoPath)
		g.State = StateShutdown
		g.FrameCounter = 0
	}

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

func (g *Game) Shutdown() {
	g.FrameCounter++

	if g.FrameCounter == 60 {
		g.State = StateExit
	}

	g.PlayCycle()
}

func (g *Game) PlayCycle() {
	// 1. Check for inputs
	g.MapInput()

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.State = StateMenu
	}

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

		// if sound timer is set play a beep
		if !g.beepPlayer.IsPlaying() {
			g.beepPlayer.Rewind()
			g.beepPlayer.Play()
		}
	} else {
		if g.beepPlayer.IsPlaying() {
			g.beepPlayer.Pause()
		}
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
	case StateShutdown:
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

	// initialize audio
	audioContext := audio.NewContext(44100)
	beepPlayer := createBeepPlayer(audioContext)

	ebiten.SetWindowSize(screenWidth*scale, screenHeight*scale)
	ebiten.SetWindowTitle("CHIP-8 Emulator")

	// TOOD: Read in all files from assets dir
	romFiles := listRomFiles("assets")
	game := &Game{
		Chip8:         emu,
		State:         StateStartup,
		AvailableROMs: romFiles,
		audioContext:  audioContext,
		beepPlayer:    beepPlayer,
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
	if ebiten.IsKeyPressed(ebiten.Key1) {
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
	if ebiten.IsKeyPressed(ebiten.Key4) {
		g.Chip8.Key[0xC] = 1
	}

	// Key Q
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
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
	if ebiten.IsKeyPressed(ebiten.KeyR) {
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
	if ebiten.IsKeyPressed(ebiten.KeyY) {
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

func listRomFiles(dir string) []string {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal("Could not read assets directory:", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if strings.HasSuffix(name, ".ch8") || strings.HasSuffix(name, ".rom") {
			if name == IbmLogoPath || name == Chip8PicturePath || name == Chip8ExitLogoPath || name == "IBM Logo.ch8" {
				continue
			}

			files = append(files, name)
		}
	}

	return files
}

func createBeepPlayer(context *audio.Context) *audio.Player {
	const sampleRate = 44100
	const freq = 440.0 // A4 = 440Hz
	const duration = 1 // 1 second buffer

	length := sampleRate * duration
	buffer := make([]byte, length*4) // 4 bytes per sample (16-bit, 2 channels)

	// generates a simple square wave signal
	for i := 0; i < length; i++ {
		sample := 0.0
		if math.Sin(2*math.Pi*freq*float64(i)/sampleRate) > 0 {
			sample = 0.3 // Amplitude 50%
		} else {
			sample = -0.3
		}

		// convert to int16
		v := int16(sample * math.MaxInt16)

		// left channel
		buffer[4*i] = byte(v)
		buffer[4*i+1] = byte(v >> 8)
		// right channel
		buffer[4*i+2] = byte(v)
		buffer[4*i+3] = byte(v >> 8)
	}

	player, err := context.NewPlayer(bytes.NewReader(buffer))
	if err != nil {
		log.Fatal(err)
	}

	return player
}
