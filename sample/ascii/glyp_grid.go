package main

import (
	"glyphrenderer/res"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//
// ─── MODEL ──────────────────────────────────────────────
//

// Glyph = a single character on the grid with foreground + background colors.
type Glyph struct {
	Ch    rune
	FGIdx int
	BGIdx int
}

// GlyphGrid = pure logical state of glyphs + palette.
type GlyphGrid struct {
	Cols, Rows int
	Glyphs     []Glyph
	Palette    *Palette
}

func NewGlyphGrid(cols, rows int, palette *Palette) *GlyphGrid {
	return &GlyphGrid{
		Cols:    cols,
		Rows:    rows,
		Glyphs:  make([]Glyph, cols*rows),
		Palette: palette,
	}
}

//
// ─── ATLAS ──────────────────────────────────────────────
//

// GlyphAtlas holds the font atlas and per-rune subimages.
type GlyphAtlas struct {
	Image     *ebiten.Image
	GlyphImgs map[rune]*ebiten.Image
	GlyphW    int
	GlyphH    int
}

// GenerateGlyphAtlas builds an ASCII atlas from a text face.
func GenerateGlyphAtlas(face *text.GoTextFace, w, h int) *GlyphAtlas {
	atlasImg, glyphRects := GenerateASCIIAtlas(face, w, h)
	textW, textH := text.Measure("M", face, 0)
	glyphW := int(textW)
	glyphH := int(textH)

	glyphImgs := make(map[rune]*ebiten.Image, len(glyphRects))
	for r, rect := range glyphRects {
		glyphImgs[r] = atlasImg.SubImage(rect).(*ebiten.Image)
	}

	return &GlyphAtlas{
		Image:     atlasImg,
		GlyphImgs: glyphImgs,
		GlyphW:    glyphW,
		GlyphH:    glyphH,
	}
}

//
// ─── BACKGROUND BUFFER ──────────────────────────────────
//

// BackgroundBuffer manages low-res per-glyph background colors.
// One pixel = one glyph cell. The image is scaled up when drawn.
type BackgroundBuffer struct {
	Pixels []byte        // RGBA buffer, cols*rows*4
	Tex    *ebiten.Image // low-res texture (cols x rows)

	Cols, Rows int
	GlyphW     int
	GlyphH     int
}

// NewBackgroundBuffer builds a low-res background buffer.
// The buffer is cols x rows pixels, scaled up by glyphW x glyphH at draw.
func NewBackgroundBuffer(cols, rows, glyphW, glyphH int) *BackgroundBuffer {
	bufSize := cols * rows * 4
	return &BackgroundBuffer{
		Pixels: make([]byte, bufSize),
		Tex:    ebiten.NewImage(cols, rows),
		Cols:   cols,
		Rows:   rows,
		GlyphW: glyphW,
		GlyphH: glyphH,
	}
}

//
// ─── RENDERER ───────────────────────────────────────────
//

// GlyphRenderer draws a GlyphGrid using an atlas + background buffer.
type GlyphRenderer struct {
	Atlas    *GlyphAtlas
	BgBuffer *BackgroundBuffer
	Grid     *GlyphGrid

	preGeo   []ebiten.GeoM       // one per grid cell
	preColor []ebiten.ColorScale // one per palette color
	op       *ebiten.DrawImageOptions

	RuneBuffer       *ebiten.Image
	RuneBufferScaled *ebiten.Image
	RuneBufferRT     *ebiten.Image

	shader *ebiten.Shader
}

func NewGlyphRenderer(grid *GlyphGrid, face *text.GoTextFace, w, h int) *GlyphRenderer {
	atlas := GenerateGlyphAtlas(face, w, h)
	bgBuffer := NewBackgroundBuffer(grid.Cols, grid.Rows, atlas.GlyphW, atlas.GlyphH)

	preGeo := make([]ebiten.GeoM, grid.Rows*grid.Cols)
	for row := 0; row < grid.Rows; row++ {
		for col := 0; col < grid.Cols; col++ {
			idx := row*grid.Cols + col
			var m ebiten.GeoM
			m.Translate(float64(col*atlas.GlyphW), float64(row*atlas.GlyphH))
			preGeo[idx] = m
		}
	}

	preColor := make([]ebiten.ColorScale, len(grid.Palette.Colors))
	for i, c := range grid.Palette.Colors {
		var cs ebiten.ColorScale
		cs.ScaleWithColor(c)
		preColor[i] = cs
	}

	return &GlyphRenderer{
		Atlas:    atlas,
		BgBuffer: bgBuffer,
		Grid:     grid,

		preGeo:           preGeo,
		preColor:         preColor,
		op:               &ebiten.DrawImageOptions{},
		RuneBuffer:       ebiten.NewImage(grid.Cols, grid.Rows),
		RuneBufferScaled: ebiten.NewImage(grid.Cols*atlas.GlyphW, grid.Rows*atlas.GlyphH),
		RuneBufferRT:     ebiten.NewImage(grid.Cols*atlas.GlyphW, grid.Rows*atlas.GlyphH),
		shader:           res.LoadShader("glyph_01.kage"),
		// shader: res.LoadShader("draw_texture.kage"),
	}
}

// UpdateBGTexture fills the low-res background buffer based on glyph BG colors.
// One pixel = one glyph cell. The image is scaled up when drawn.
func (r *GlyphRenderer) UpdateBGTexture() {
	for row := 0; row < r.Grid.Rows; row++ {
		for col := 0; col < r.Grid.Cols; col++ {
			idx := row*r.Grid.Cols + col
			glyph := r.Grid.Glyphs[idx]

			// Look up the background color from the palette
			c := r.Grid.Palette.Colors[glyph.BGIdx]
			cr, cg, cb, ca := c.RGBA()

			// Destination pixel (1px per glyph cell)
			base := (row*r.BgBuffer.Cols + col) * 4
			r.BgBuffer.Pixels[base+0] = uint8(cr >> 8)
			r.BgBuffer.Pixels[base+1] = uint8(cg >> 8)
			r.BgBuffer.Pixels[base+2] = uint8(cb >> 8)
			r.BgBuffer.Pixels[base+3] = uint8(ca >> 8)
		}
	}

	// Upload to GPU once per frame
	r.BgBuffer.Tex.WritePixels(r.BgBuffer.Pixels)
}

// UpdateRuneBuffer writes per-cell glyph + color data into the rune buffer.
// Each texel encodes: R=glyphIdx, G=fgIdx, B=effect (0), A=255.
func (r *GlyphRenderer) UpdateRuneBuffer() {
	pixels := make([]byte, r.Grid.Cols*r.Grid.Rows*4)

	for row := 0; row < r.Grid.Rows; row++ {
		for col := 0; col < r.Grid.Cols; col++ {
			idx := row*r.Grid.Cols + col
			glyph := r.Grid.Glyphs[idx]

			// Map rune → atlas index (for now, just cast rune to byte)
			glyphIdx := byte(glyph.Ch - startASCII)
			fgIdx := byte(glyph.FGIdx)

			base := idx * 4
			pixels[base+0] = glyphIdx
			pixels[base+1] = fgIdx
			pixels[base+2] = 0   // effect
			pixels[base+3] = 255 // alpha
		}
	}

	r.RuneBuffer.WritePixels(pixels)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(r.Atlas.GlyphW), float64(r.Atlas.GlyphH))
	r.RuneBufferScaled.Clear()
	r.RuneBufferScaled.DrawImage(r.RuneBuffer, op)

	// Foreground via shader
	sOp := &ebiten.DrawRectShaderOptions{}
	sOp.Uniforms = map[string]any{
		"Cols":   float32(r.Grid.Cols),
		"Rows":   float32(r.Grid.Rows),
		"GlyphW": float32(r.Atlas.GlyphW),
		"GlyphH": float32(r.Atlas.GlyphH),
	}
	sOp.Images[0] = r.Atlas.Image
	sOp.Images[1] = r.RuneBufferScaled

	r.RuneBufferRT.Clear()
	r.RuneBufferRT.DrawRectShader(r.RuneBufferRT.Bounds().Dx(), r.RuneBufferRT.Bounds().Dy(), r.shader, sOp)

}

// Draw renders both background + foreground glyphs to the target image.
func (r *GlyphRenderer) Draw(dst *ebiten.Image) {
	// Background
	r.op.GeoM.Reset()
	r.op.ColorScale.Reset()
	r.op.GeoM.Scale(float64(r.Atlas.GlyphW), float64(r.Atlas.GlyphH))
	dst.DrawImage(r.BgBuffer.Tex, r.op)

	// // Foreground
	// for row := 0; row < r.Grid.Rows; row++ {
	// 	for col := 0; col < r.Grid.Cols; col++ {
	// 		idx := row*r.Grid.Cols + col
	// 		glyph := r.Grid.Glyphs[idx]

	// 		img := r.Atlas.GlyphImgs[glyph.Ch]
	// 		if img == nil {
	// 			continue
	// 		}

	// 		r.op.GeoM = r.preGeo[idx]
	// 		r.op.ColorScale = r.preColor[glyph.FGIdx]

	// 		dst.DrawImage(img, r.op)
	// 	}
	// }

	// // Foreground via shader
	// op := &ebiten.DrawRectShaderOptions{}
	// op.Uniforms = map[string]any{
	// 	"Cols":   float32(r.Grid.Cols),
	// 	"Rows":   float32(r.Grid.Rows),
	// 	"GlyphW": float32(r.Atlas.GlyphW),
	// 	"GlyphH": float32(r.Atlas.GlyphH),
	// }
	// op.Images[0] = r.Atlas.Image
	// op.Images[1] = r.RuneBufferScaled

	// dst.DrawRectShader(dst.Bounds().Dx(), dst.Bounds().Dy(), r.shader, op)

	dst.DrawImage(r.RuneBufferRT, nil)
}

//
// ─── UTILITIES ──────────────────────────────────────────
//

// FillRandom assigns random runes and colors to all glyphs in the grid.
func FillRandom(grid *GlyphGrid, rng *rand.Rand) {
	for i := range grid.Glyphs {
		grid.Glyphs[i] = Glyph{
			Ch:    rune(rng.Intn(runeCount) + startASCII),
			FGIdx: rng.Intn(len(grid.Palette.Colors)),
			BGIdx: rng.Intn(len(grid.Palette.Colors)),
		}
	}
}

// PrebuildBGRects creates one solid-colored rectangle per palette color.
// Each rectangle is a []byte buffer of size glyphW*glyphH*4 (RGBA8).
func PrebuildBGRects(glyphW, glyphH int, colors []color.Color) [][]byte {
	rects := make([][]byte, len(colors))
	for i, col := range colors {
		cr, cg, cb, ca := col.RGBA()
		cr8 := uint8(cr >> 8)
		cg8 := uint8(cg >> 8)
		cb8 := uint8(cb >> 8)
		ca8 := uint8(ca >> 8)

		buf := make([]byte, glyphW*glyphH*4)
		for px := 0; px < glyphW*glyphH; px++ {
			base := px * 4
			buf[base+0] = cr8
			buf[base+1] = cg8
			buf[base+2] = cb8
			buf[base+3] = ca8
		}
		rects[i] = buf
	}
	return rects
}

//
// ─── EXAMPLE RNG CONSTRUCTOR ────────────────────────────
//

func NewRNG() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
