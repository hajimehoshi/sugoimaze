// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package game

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Field struct {
	data         *FieldData
	playerX      int
	playerY      int
	currentDepth int

	playerImage *ebiten.Image
}

func NewField(difficulty Difficulty) *Field {
	f := &Field{
		data:    NewFieldData(difficulty),
		playerX: 1,
		playerY: 1,
	}

	f.playerImage = f.data.tilesImage.SubImage(image.Rect(1*GridSize, 0*GridSize, 2*GridSize, 1*GridSize)).(*ebiten.Image)

	return f
}

func (f *Field) Update() {
	x, y := f.playerX, f.playerY
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		y++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		y--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		x--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		x++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if f.data.hasSwitch(x, y) {
			f.currentDepth++
			f.currentDepth %= f.data.depth
		}
	}
	if !f.data.passable(x, y, f.currentDepth) {
		return
	}
	f.playerX, f.playerY = x, y
}

func (f *Field) Draw(screen *ebiten.Image) {
	cx := screen.Bounds().Dx() / 2
	cy := screen.Bounds().Dy() / 3 * 2
	offsetX := cx - f.playerX*GridSize
	offsetY := cy + f.playerY*GridSize
	f.data.Draw(screen, offsetX, offsetY, f.currentDepth)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(f.playerX*GridSize), float64(-((f.playerY + 1) * GridSize)))
	op.GeoM.Translate(float64(offsetX), float64(offsetY))
	screen.DrawImage(f.playerImage, op)
}
