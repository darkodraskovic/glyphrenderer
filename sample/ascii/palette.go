package main

import (
	"image/color"
	"sort"
)

type Palette struct {
	Colors []color.Color  // indexed colors for rendering
	Names  map[string]int // name → index lookup
}

// Predefined named colors.
var baseColors = map[string]color.Color{
	"white":     color.RGBA{255, 255, 255, 255},
	"black":     color.RGBA{0, 0, 0, 255},
	"red":       color.RGBA{255, 0, 0, 255},
	"green":     color.RGBA{0, 255, 0, 255},
	"blue":      color.RGBA{0, 0, 255, 255},
	"yellow":    color.RGBA{255, 255, 0, 255},
	"cyan":      color.RGBA{0, 255, 255, 255},
	"magenta":   color.RGBA{255, 0, 255, 255},
	"gray":      color.RGBA{128, 128, 128, 255},
	"darkred":   color.RGBA{128, 0, 0, 255},
	"darkgreen": color.RGBA{0, 128, 0, 255},
	"darkblue":  color.RGBA{0, 0, 128, 255},
	"olive":     color.RGBA{128, 128, 0, 255},
	"teal":      color.RGBA{0, 128, 128, 255},
	"purple":    color.RGBA{128, 0, 128, 255},
	"silver":    color.RGBA{192, 192, 192, 255},
}

// NewPalette builds a Palette with deterministic ordering of indices.
func NewPalette(colorMap map[string]color.Color) *Palette {
	names := make([]string, 0, len(colorMap))
	for name := range colorMap {
		names = append(names, name)
	}
	sort.Strings(names)

	colors := make([]color.Color, len(names))
	nameToIdx := make(map[string]int, len(names))
	for i, name := range names {
		colors[i] = colorMap[name]
		nameToIdx[name] = i
	}

	return &Palette{
		Colors: colors,
		Names:  nameToIdx,
	}
}
