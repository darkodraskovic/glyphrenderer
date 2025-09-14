package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	startASCII = 32
	endASCII   = 126
	runeCount  = endASCII - startASCII + 1
)

// GenerateASCIIAtlas builds a glyph atlas padded to exactly screenW × screenH.
// Glyphs are packed row by row without skipping rows.
// Returns the atlas and a map of rune -> glyph rectangle.
func GenerateASCIIAtlas(face *text.GoTextFace, screenW, screenH int) (*ebiten.Image, map[rune]image.Rectangle) {
	cellW, cellH := text.Measure("M", face, 0)
	cellWInt := int(cellW)
	cellHInt := int(cellH)

	atlas := ebiten.NewImage(screenW, screenH)
	glyphMap := make(map[rune]image.Rectangle)

	opts := &text.DrawOptions{}
	opts.ColorScale.ScaleWithColor(color.White)

	ch := startASCII
	for c := 0; ch <= endASCII; c++ {
		x := float64(c * cellWInt)
		y := 0.0

		opts.GeoM.Reset()
		opts.GeoM.Translate(x, y)
		text.Draw(atlas, string(rune(ch)), face, opts)

		rect := image.Rect(
			c*cellWInt,
			0,
			(c+1)*cellWInt,
			cellHInt,
		)
		glyphMap[rune(ch)] = rect

		ch++
	}

	return atlas, glyphMap
}
