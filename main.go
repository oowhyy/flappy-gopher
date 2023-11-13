// Copyright 2018 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	_ "embed"
	"flag"
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

var flagCRT = flag.Bool("crt", false, "enable the CRT effect")

//go:embed crt.go
var crtGo []byte

const (
	windowWidth  = 640
	windowHeight = 480

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
}

func init() {
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
}

func NewGame(crt bool) ebiten.Game {
	g := &Game{}
	g.init()
	if crt {
		return &GameWithCRTEffect{Game: g}
	}
	return g
}

func (g *Game) init() {
	g.gophers = make(map[int]*Gopher, 0)
	g.gophers[1] = NewGopher(1, 120, 100)
	g.gophers[2] = NewGopher(2, 200, 100)
	g.SpawnDelay = 110
	g.GapY = 180
	g.Pipes = make([]*Pipe, 0)
	g.Speed = 3
	g.Base = NewBase(g.Speed)
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
	return windowWidth, windowHeight
}

func (g *Game) Step() {
	switch g.mode {
	case ModeTitle:
		if g.isKeyJustPressed() {
			g.mode = ModeGame
		}
	case ModeGame:
		if g.isKeyJustPressed() {
			for _, gopher := range g.gophers {
				gopher.Jump()
			}
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
			newPipe := NewPipe(g.GapY, g.Speed)
			g.Pipes = append(g.Pipes, newPipe)
		}
		g.Timer--

		// check hit
		for _, pipe := range g.Pipes {
			for _, gopher := range g.gophers {
				if pipe.Collide(gopher) || gopher.OffScreenY(windowHeight-tileSize) {
					delete(g.gophers, gopher.ID)
				}
				g.score += pipe.Score(gopher)
			}
		}
		if len(g.gophers) == 0 {
			g.mode = ModeGameOver
			g.gameoverCount = 30
		}

	case ModeGameOver:
		if g.gameoverCount > 0 {
			g.gameoverCount--
		}
		if g.gameoverCount == 0 && g.isKeyJustPressed() {
			g.init()
			g.mode = ModeTitle
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
		x := (windowWidth - len(l)*titleFontSize) / 2
		text.Draw(screen, l, titleArcadeFont, x, (i+4)*titleFontSize, color.White)
	}
	for i, l := range texts {
		x := (windowWidth - len(l)*fontSize) / 2
		text.Draw(screen, l, arcadeFont, x, (i+4)*fontSize, color.White)
	}

	if g.mode == ModeTitle {
		msg := []string{
			"Go Gopher by Renee French is",
			"licenced under CC BY 3.0.",
		}
		for i, l := range msg {
			x := (windowWidth - len(l)*smallFontSize) / 2
			text.Draw(screen, l, smallArcadeFont, x, windowHeight-4+(i-1)*smallFontSize, color.White)
		}
	}

	scoreStr := fmt.Sprintf("%04d", g.Score())
	text.Draw(screen, scoreStr, arcadeFont, windowWidth-len(scoreStr)*fontSize, fontSize, color.White)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func (g *Game) Score() int {
	return g.score
}

// func (g *Game) drawTiles(screen *ebiten.Image) {
// 	const (
// 		nx           = screenWidth / tileSize
// 		ny           = screenHeight / tileSize
// 		pipeTileSrcX = 128
// 		pipeTileSrcY = 192
// 	)
// 	// op2 := &ebiten.DrawImageOptions{}
// 	// op2.GeoM.Reset()
// 	// op2.GeoM.Translate(float64(g.cameraX), float64(g.cameraY))
// 	// screen.DrawImage(tilesImage, op2)

// 	rectTop := image.Rect(pipeTileSrcX, pipeTileSrcY, pipeTileSrcX+pipeWidth, pipeTileSrcY+tileSize)
// 	rectShaft := image.Rect(pipeTileSrcX, pipeTileSrcY+tileSize, pipeTileSrcX+pipeWidth, pipeTileSrcY+pipeWidth)

// 	op := &ebiten.DrawImageOptions{}
// 	for i := -2; i < nx+1; i++ {
// 		// ground
// 		op.GeoM.Reset()
// 		op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
// 			float64((ny-1)*tileSize-floorMod(g.cameraY, tileSize)))
// 		screen.DrawImage(tilesImage.SubImage(image.Rect(0, 0, tileSize, tileSize)).(*ebiten.Image), op)

// 		// pipe
// 		if tileY, ok := g.pipeAt(floorDiv(g.cameraX, tileSize) + i); ok {
// 			for j := 0; j < tileY; j++ {
// 				op.GeoM.Reset()
// 				op.GeoM.Scale(1, -1)
// 				op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
// 					float64(j*tileSize-floorMod(g.cameraY, tileSize)))
// 				op.GeoM.Translate(0, tileSize)
// 				var r image.Rectangle
// 				if j == tileY-1 {
// 					r = rectTop
// 				} else {
// 					r = rectShaft
// 				}
// 				screen.DrawImage(tilesImage.SubImage(r).(*ebiten.Image), op)
// 			}
// 			for j := tileY + g.GapY; j < screenHeight/tileSize-1; j++ {
// 				op.GeoM.Reset()
// 				op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
// 					float64(j*tileSize-floorMod(g.cameraY, tileSize)))
// 				var r image.Rectangle
// 				if j == tileY+g.GapY {
// 					r = rectTop
// 				} else {
// 					r = rectShaft
// 				}
// 				screen.DrawImage(tilesImage.SubImage(r).(*ebiten.Image), op)
// 			}
// 		}
// 	}
// }

type GameWithCRTEffect struct {
	ebiten.Game

	crtShader *ebiten.Shader
}

func (g *GameWithCRTEffect) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
	if g.crtShader == nil {
		s, err := ebiten.NewShader(crtGo)
		if err != nil {
			panic(fmt.Sprintf("flappy: failed to compiled the CRT shader: %v", err))
		}
		g.crtShader = s
	}

	os := offscreen.Bounds().Size()

	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = offscreen
	op.GeoM = geoM
	screen.DrawRectShader(os.X, os.Y, g.crtShader, op)
}

func main() {
	flag.Parse()
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Flappy Gopher (Ebitengine Demo)")
	if err := ebiten.RunGame(NewGame(*flagCRT)); err != nil {
		panic(err)
	}
}
