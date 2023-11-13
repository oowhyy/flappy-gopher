package main

import (
	"image"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	pipeTileSrcX = 128
	pipeTileSrcY = 192
	tileSize     = 32
	PipeWidth    = tileSize * 2
)

func init() {
	rectTop := image.Rect(pipeTileSrcX, pipeTileSrcY, pipeTileSrcX+PipeWidth, pipeTileSrcY+tileSize)
	rectShaft := image.Rect(pipeTileSrcX, pipeTileSrcY+tileSize, pipeTileSrcX+PipeWidth, pipeTileSrcY+PipeWidth)
	topImg = tilesImage.SubImage(rectTop).(*ebiten.Image)
	shaftImg = tilesImage.SubImage(rectShaft).(*ebiten.Image)
}

var (
	topImg   *ebiten.Image
	shaftImg *ebiten.Image
)

type Pipe struct {
	x        int
	topY     int
	speed    int
	gap      int
	topImg   *ebiten.Image
	shaftImg *ebiten.Image
	passed   bool
	scored   bool
}

func NewPipe(gap int, speed int) *Pipe {
	topY := rand.Intn(windowHeight-gap-2*tileSize) + tileSize
	return &Pipe{
		x:        windowWidth,
		topY:     topY,
		speed:    speed,
		gap:      gap,
		topImg:   topImg,
		shaftImg: shaftImg,
	}
}

func (p *Pipe) Draw(screen *ebiten.Image) {
	// top part
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1, -1)
	topShaftH := p.topY - p.topImg.Bounds().Dy()
	topShaftTiling := toTiling(p.shaftImg, p.Width(), topShaftH)
	op.GeoM.Translate(float64(p.x), float64(topShaftH))
	screen.DrawImage(topShaftTiling, op)
	op.GeoM.Translate(0, tileSize)
	screen.DrawImage(p.topImg, op)

	//bottom part
	deltaY := p.topY + p.gap
	op.GeoM.Reset()
	botShaftH := windowHeight - deltaY - tileSize
	botShaftTiling := toTiling(p.shaftImg, p.Width(), botShaftH)
	op.GeoM.Translate(float64(p.x), float64(deltaY))
	screen.DrawImage(p.topImg, op)
	op.GeoM.Translate(0, tileSize)
	screen.DrawImage(botShaftTiling, op)
}

func (p *Pipe) Collide(gopher *Gopher) bool {
	// not reached
	if p.PosX() > gopher.PosX()+gopher.Width() {
		return false
	}
	// already passed
	if p.PosX()+p.Width() < gopher.PosX() {
		return false
	}
	hitBot := gopher.PosY()+gopher.Height() > p.PosBotY()
	hitTop := gopher.PosY() < p.PosTopY()
	return hitTop || hitBot
}

func (p *Pipe) Passed(gopher *Gopher) bool {
	if p.passed {
		return true
	}
	// already passed
	if p.PosX()+p.Width() < gopher.PosX() {
		p.passed = true
	}
	return p.passed
}
func (p *Pipe) Score(gopher *Gopher) int {
	if p.scored {
		return 0
	}
	if p.Passed(gopher) {
		p.passed = true
		p.scored = true
		return 1
	}
	return 0
}

func (p *Pipe) Move() {
	p.x -= p.speed
}

func (p *Pipe) Width() int {
	return PipeWidth
}

func (p *Pipe) PosX() int {
	return p.x
}

func (p *Pipe) PosTopY() int {
	return p.topY
}
func (p *Pipe) PosBotY() int {
	return p.topY + p.gap
}

func toTiling(img *ebiten.Image, targetWidth, targetHeight int) *ebiten.Image {
	if targetHeight < 1 || targetWidth < 1 {
		return ebiten.NewImage(1, 1)
	}
	res := ebiten.NewImage(targetWidth, targetHeight)
	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()

	for i := 0; i < (targetHeight+imgH-1)/imgH; i++ {
		for j := 0; j < (targetWidth+imgW-1)/imgW; j++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(j*imgW), float64(i*imgH))
			res.DrawImage(img, op)
		}
	}
	return res
}
