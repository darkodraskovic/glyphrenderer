package main

import (
	"fmt"
	"glyphrenderer/res"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	rows     = 96
	cols     = 56
	fontSize = 8

	logicalScreenWidth  = rows * fontSize
	logicalScreenHeight = cols * fontSize
	pixelsScale         = 2
)

type Game struct {
	Grid     *GlyphGrid
	Renderer *GlyphRenderer
	RNG      *rand.Rand
}

func NewGame(face *text.GoTextFace, palette *Palette) *Game {
	grid := NewGlyphGrid(logicalScreenWidth/fontSize, logicalScreenHeight/fontSize, palette)
	renderer := NewGlyphRenderer(grid, face, logicalScreenWidth, logicalScreenHeight)

	return &Game{
		Grid:     grid,
		Renderer: renderer,
		RNG:      NewRNG(),
	}
}

var frameCount int

func (g *Game) Update() error {
	FillRandom(g.Grid, g.RNG)
	g.Renderer.UpdateBGTexture()

	// if frameCount < 13 {
	g.Renderer.UpdateRuneBuffer()
	// }

	frameCount++

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Renderer.Draw(screen)

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrint(screen, "FPS: "+fmt.Sprintf("%.2f", fps))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// logical size
	return logicalScreenWidth, logicalScreenHeight
}

func main() {
	face := res.LoadFace("PressStart2P-Regular.ttf", fontSize)
	palette := NewPalette(baseColors)

	game := NewGame(face, palette)

	ebiten.SetWindowSize(logicalScreenWidth*pixelsScale, logicalScreenHeight*pixelsScale)
	ebiten.SetWindowTitle("Glyph Grid Test")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
