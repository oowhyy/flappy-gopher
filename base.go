package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Base struct {
	baseImg *ebiten.Image
	y       int
	speed   int
	delta   int
}

func NewBase(speed int) *Base {
	baseimg := tilesImage.SubImage(image.Rect(0, 0, tileSize, tileSize)).(*ebiten.Image)
	return &Base{
		baseImg: baseimg,
		speed:   speed,
		y:       windowHeight - baseimg.Bounds().Dy(),
	}
}

func (b *Base) TileWidth() int {
	return b.baseImg.Bounds().Dx()
}

func (b *Base) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(b.delta), float64(b.y))
	for i := 0; i < screen.Bounds().Dx()/b.TileWidth()+2; i++ {
		screen.DrawImage(b.baseImg, op)
		op.GeoM.Translate(float64(b.TileWidth()), 0)
	}
}

func (b *Base) Move() {
	b.delta += b.speed
	if b.delta >= b.TileWidth() {
		b.delta -= b.TileWidth()
	}
}
