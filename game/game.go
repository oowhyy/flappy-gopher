package game

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	resources "github.com/hajimehoshi/ebiten/v2/examples/resources/images/flappy"
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
}

type GameMode string

const (
	ModePlay     GameMode = "play"
	ModeGameOver GameMode = "gameover"
)

type Game struct {
	// meta
	mode           GameMode
	stepID         int
	mu             sync.Mutex
	muDraw         sync.Mutex
	score          int
	dynamicSPU     bool
	stepsPerUpdate int
	resetsNum      int

	// window
	windowW int
	windowH int

	// gopher
	gophers  map[int]*Gopher
	gophersX int

	// general scrollX speed
	speed int

	// pipes
	pipes      []*Pipe
	pipesAhead []*Pipe
	spawnDelay int
	gapY       int
	spawnTimer int

	// base
	base *Base

	// api
	inpChan chan map[int]bool

	statesChan     chan *State
	activeRequests atomic.Int32
	done           chan bool
}

func NewGame(windowW, windowHeight int, gopherN int) *Game {
	g := &Game{
		windowW: windowW,
		windowH: windowHeight,
	}
	g.restart(gopherN)
	return g
}

func (g *Game) restart(gopherN int) {
	g.muDraw.Lock()
	defer g.muDraw.Unlock()
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.done != nil && g.mode != ModeGameOver {
		close(g.done)
	}
	g.dynamicSPU = false
	g.mode = ModePlay
	g.gophers = make(map[int]*Gopher, gopherN)
	g.gophersX = 120
	g.stepID = 0
	for i := 0; i < gopherN; i++ {
		g.gophers[i] = NewGopher(i, 100, g.windowH*3/4-gopherImage.Bounds().Dy()/2-rand.Intn(g.windowH/2))
	}
	g.resetsNum++
	// g.gophers[2] = NewGopher(2, 200, 100)

	g.spawnTimer = 0
	g.score = 0
	g.stepsPerUpdate = 1
	g.inpChan = make(chan map[int]bool)
	g.statesChan = make(chan *State)
	g.done = make(chan bool)
	g.spawnDelay = 110
	g.gapY = 180
	g.pipes = make([]*Pipe, 0)
	g.pipesAhead = make([]*Pipe, 0)
	g.speed = 3
	g.base = NewBase(g.windowH, g.speed)
}

func (g *Game) GameOver() {
	g.mode = ModeGameOver

	// close(g.inpChan)
	close(g.done)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.windowW, g.windowH
}

func (g *Game) Step() {
	g.mu.Lock()
	defer func() {
		g.mu.Unlock()
		g.stepID++
	}()
	switch g.mode {
	case ModePlay:
		// process input
		select {
		case inp := <-g.inpChan:
			for id, jump := range inp {
				gopher, ok := g.gophers[id]
				if ok && jump {
					gopher.Jump()
				}
			}
		default:
			// fmt.Println("no input this tick...")
		}

		for _, gopher := range g.gophers {
			gopher.Move()
		}
		g.base.Move()

		// Update pipes
		// 1. remove old
		if len(g.pipes) > 0 && g.pipes[0].PosX()+g.pipes[0].Width() < 0 {
			g.pipes = g.pipes[1:]
		}
		// 2. move
		for _, pipe := range g.pipes {
			pipe.Move()
		}
		// 3. spawn new
		g.SpawnPipe()

		// 4. check pass
		if len(g.pipesAhead) > 0 {
			p0 := g.pipesAhead[0]
			if p0.Passed(g.gophersX) {
				g.score++
				g.pipesAhead = g.pipesAhead[1:]
			}
		}

		// check hit
		for _, pipe := range g.pipes {
			for _, gopher := range g.gophers {
				if pipe.Collide(gopher) || gopher.OffScreenY(g.windowH-tileSize) {
					delete(g.gophers, gopher.ID)
				}
			}
		}
		if len(g.gophers) == 0 {
			g.GameOver()
		}
	case ModeGameOver:
		// do nothing until reset
	}

	g.pushState()
}

func (g *Game) SpawnPipe() {
	if g.spawnTimer <= 0 {
		g.spawnTimer = g.spawnDelay
		newPipe := NewPipe(g.windowW, g.windowH, g.gapY, g.speed)
		g.pipes = append(g.pipes, newPipe)
		g.pipesAhead = append(g.pipesAhead, newPipe)
	}
	g.spawnTimer--
}

func (g *Game) Update() error {
	if g.dynamicSPU {
		for i := 0; i < g.stepsPerUpdate; i++ {
			g.Step()
		}

		if ebiten.ActualTPS() < 50 {
			if g.stepsPerUpdate > 1 {
				g.stepsPerUpdate--
			}
		} else {
			g.stepsPerUpdate++
		}
	} else {
		g.Step()
		g.Step()
		// g.Step()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.muDraw.Lock()
	defer g.muDraw.Unlock()
	screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})
	for _, pipe := range g.pipes {
		pipe.Draw(screen)
	}
	for _, gopher := range g.gophers {
		gopher.Draw(screen)
	}
	g.base.Draw(screen)

	var titleTexts []string
	var texts []string
	if g.mode == ModeGameOver {
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

	scoreStr := fmt.Sprintf("%04d", g.Score())
	text.Draw(screen, scoreStr, arcadeFont, g.windowW-len(scoreStr)*fontSize, fontSize, color.White)
	resetsStr := fmt.Sprintf("Gen: %d", g.resetsNum)
	text.Draw(screen, resetsStr, arcadeFont, 10, 2*fontSize, color.White)
	text.Draw(screen, scoreStr, arcadeFont, g.windowW-len(scoreStr)*fontSize, fontSize, color.White)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func (g *Game) Score() int {
	return g.score
}
