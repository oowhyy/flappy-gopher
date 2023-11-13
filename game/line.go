package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func DrawLine(screen *ebiten.Image, x1, y1, x2, y2 int) {
	pix := ebiten.NewImage(1, 10)
	pix.Fill(color.RGBA{255, 0, 0, 100})
	op := &ebiten.DrawImageOptions{}
	dx := float64(x1) - float64(x2)
	dy := float64(y1) - float64(y2)
	atan := math.Atan2(dy, dx) + math.Pi
	op.GeoM.Scale(math.Sqrt(dx*dx+dy*dy), 1)
	op.GeoM.Rotate(atan)
	op.GeoM.Translate(float64(x1), float64(y1))
	screen.DrawImage(pix, op)

}
