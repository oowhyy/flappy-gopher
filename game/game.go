package game

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	resources "github.com/hajimehoshi/ebiten/v2/examples/resources/images/flappy"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	titleFontSize = fontSize * 1.5
	fontSize      = 24
	smallFontSize = fontSize / 2
)

var (
	gopherImage     *ebiten.Image
	tilesImage      *ebiten.Image
	titleArcadeFont font.Face
	arcadeFont      font.Face
	smallArcadeFont font.Face
)

func init() {
	img, _, err := image.Decode(bytes.NewReader(resources.Gopher_png))
	if err != nil {
		log.Fatal(err)
	}
	gopherImage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(resources.Tiles_png))
	if err != nil {
		log.Fatal(err)
	}
	tilesImage = ebiten.NewImageFromImage(img)

	tt, err := opentype.Parse(fonts.PressStart2P_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	titleArcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    titleFontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	arcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	smallArcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    smallFontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
)

type Game struct {
	mode Mode

	// window
	windowW int
	windowH int

	gophers map[int]*Gopher

	Speed int

	// Pipes
	Pipes      []*Pipe
	SpawnDelay int
	GapY       int
	Timer      int

	// Base
	Base *Base

	gameoverCount int
	score         int

	// api
	inpChan chan map[int]bool
	done    chan bool
}

func NewGame(windowW, windowHeight int) *Game {
	g := &Game{
		windowW: windowW,
		windowH: windowHeight,
	}
	g.Restart()
	return g
}

func (g *Game) Restart() {
	g.mode = ModeTitle
	g.gophers = make(map[int]*Gopher, 0)
	g.gophers[1] = NewGopher(1, 120, 100)
	// g.gophers[2] = NewGopher(2, 200, 100)

	g.Timer = 0
	g.inpChan = make(chan map[int]bool)
	g.done = make(chan bool)
	g.SpawnDelay = 110
	g.GapY = 180
	g.Pipes = make([]*Pipe, 0)
	g.Speed = 3
	g.Base = NewBase(g.windowH, g.Speed)
}

func (g *Game) GameOver() {
	g.mode = ModeGameOver
	g.gameoverCount = 30
	// close(g.inpChan)
	close(g.done)
}

func (g *Game) isKeyJustPressed() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return true
	}
	return false
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.windowW, g.windowH
}

func (g *Game) Step() {
	switch g.mode {
	case ModeTitle:
		if g.isKeyJustPressed() {
			g.mode = ModeGame
		}
	case ModeGame:

		// if g.qInput != nil {
		// 	fmt.Println("jumping...")
		// 	for id, jump := range g.qInput {
		// 		if jump {
		// 			g.gophers[id].Jump()
		// 		}
		// 	}
		// 	g.qInput = nil
		// }
		select {
		case inp := <-g.inpChan:
			fmt.Println("jumping...")
			for id, jump := range inp {
				if jump {
					g.gophers[id].Jump()
				}
			}
		default:
			// fmt.Println("no input this tick...")
		}

		for _, gopher := range g.gophers {
			gopher.Move()
		}
		g.Base.Move()

		// Update pipes
		// 1. remove old
		if len(g.Pipes) > 0 && g.Pipes[0].PosX()+g.Pipes[0].Width() < 0 {
			g.Pipes = g.Pipes[1:]
		}
		// 2. move
		for _, pipe := range g.Pipes {
			pipe.Move()
		}
		// 3. spawn new
		if g.Timer <= 0 {
			g.Timer = g.SpawnDelay
			newPipe := NewPipe(g.windowW, g.windowH, g.GapY, g.Speed)
			g.Pipes = append(g.Pipes, newPipe)
		}
		g.Timer--

		// check hit
		for _, pipe := range g.Pipes {
			for _, gopher := range g.gophers {
				if pipe.Collide(gopher) || gopher.OffScreenY(g.windowH-tileSize) {
					delete(g.gophers, gopher.ID)
				}
				g.score += pipe.Score(gopher)
			}
		}
		if len(g.gophers) == 0 {
			g.GameOver()
		}

	case ModeGameOver:
		if g.gameoverCount > 0 {
			g.gameoverCount--
		}
		if g.gameoverCount == 0 && g.isKeyJustPressed() {
			g.Restart()
		}
	}
}

func (g *Game) Update() error {

	g.Step()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})
	for _, pipe := range g.Pipes {
		pipe.Draw(screen)
	}
	if g.mode != ModeTitle {
		for _, gopher := range g.gophers {
			gopher.Draw(screen)
		}
		g.Base.Draw(screen)
	}
	var titleTexts []string
	var texts []string
	switch g.mode {
	case ModeTitle:
		titleTexts = []string{"FLAPPY GOPHER"}
		texts = []string{"", "", "", "", "", "", "", "PRESS SPACE KEY", "", "OR A/B BUTTON", "", "OR TOUCH SCREEN"}
	case ModeGameOver:
		texts = []string{"", "GAME OVER!"}
	}

	// texts
	for i, l := range titleTexts {
		x := (g.windowW - len(l)*titleFontSize) / 2
		text.Draw(screen, l, titleArcadeFont, x, (i+4)*titleFontSize, color.White)
	}
	for i, l := range texts {
		x := (g.windowW - len(l)*fontSize) / 2
		text.Draw(screen, l, arcadeFont, x, (i+4)*fontSize, color.White)
	}

	if g.mode == ModeTitle {
		msg := []string{
			"Go Gopher by Renee French is",
			"licenced under CC BY 3.0.",
		}
		for i, l := range msg {
			x := (g.windowW - len(l)*smallFontSize) / 2
			text.Draw(screen, l, smallArcadeFont, x, g.windowH-4+(i-1)*smallFontSize, color.White)
		}
	}

	scoreStr := fmt.Sprintf("%04d", g.Score())
	text.Draw(screen, scoreStr, arcadeFont, g.windowW-len(scoreStr)*fontSize, fontSize, color.White)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func (g *Game) Score() int {
	return g.score
}
