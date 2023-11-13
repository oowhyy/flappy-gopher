package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Gopher struct {
	ID     int
	x      float64
	y      float64
	speedY float64
}

func NewGopher(ID int, x, y float64) *Gopher {
	return &Gopher{
		ID:     ID,
		x:      x,
		y:      y,
		speedY: 0,
	}
}

func (g *Gopher) Jump() {
	g.speedY = -6
}

func (g *Gopher) Move() {
	// g.x += 2
	g.y += g.speedY
	// Gravity
	g.speedY += 0.25
	if g.speedY > 96 {
		g.speedY = 96
	}
}

func (g *Gopher) PosX() int {
	return int(g.x)
}

func (g *Gopher) PosY() int {
	return int(g.y)
}

func (g *Gopher) Width() int {
	return gopherImage.Bounds().Dx()
}

func (g *Gopher) Height() int {
	return gopherImage.Bounds().Dy()
}

func (g *Gopher) OffScreenY(windowH int) bool {
	return g.PosY() < 0 || g.PosY()+g.Height() > windowH
}

func (g *Gopher) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	w, h := g.Width(), g.Height()
	// fmt.Println(g.PosY())
	// box := ebiten.NewImage(w, h)
	// box.Fill(color.Black)
	op.GeoM.Translate(-float64(w)/2.0, -float64(h)/2.0)
	op.GeoM.Rotate(float64(g.speedY) / 6.0 * math.Pi / 6)
	op.GeoM.Translate(float64(w)/2.0, float64(h)/2.0)
	op.GeoM.Translate(float64(g.x), float64(g.y))
	op.Filter = ebiten.FilterLinear
	// screen.DrawImage(box, op)
	screen.DrawImage(gopherImage, op)
}
