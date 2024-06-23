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
	dx           int
	dy           int
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
	const v = 3

	if f.dx != 0 || f.dy != 0 {
		if f.dx > 0 {
			f.dx += v
		} else if f.dx < 0 {
			f.dx -= v
		}
		if f.dy > 0 {
			f.dy += v
		} else if f.dy < 0 {
			f.dy -= v
		}
		if f.dx >= GridSize {
			f.playerX++
			f.dx = 0
		}
		if f.dx <= -GridSize {
			f.playerX--
			f.dx = 0
		}
		if f.dy >= GridSize {
			f.playerY++
			f.dy = 0
		}
		if f.dy <= -GridSize {
			f.playerY--
			f.dy = 0
		}
		return
	}

	x, y := f.playerX, f.playerY
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		y++
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		y--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		x--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
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
	if x > f.playerX {
		f.dx = v
	}
	if x < f.playerX {
		f.dx = -v
	}
	if y > f.playerY {
		f.dy = v
	}
	if y < f.playerY {
		f.dy = -v
	}
}

func (f *Field) Draw(screen *ebiten.Image) {
	cx := screen.Bounds().Dx() / 2
	cy := screen.Bounds().Dy() / 3 * 2
	offsetX := cx - (f.playerX*GridSize + f.dx)
	offsetY := cy + (f.playerY*GridSize + f.dy)
	f.data.Draw(screen, offsetX, offsetY, f.currentDepth)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(f.playerX*GridSize+f.dx), float64(-((f.playerY+1)*GridSize + f.dy)))
	op.GeoM.Translate(float64(offsetX), float64(offsetY))
	screen.DrawImage(f.playerImage, op)
}
