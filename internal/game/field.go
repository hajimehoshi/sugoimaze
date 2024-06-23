// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package game

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
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

func (f *Field) Draw(screen *ebiten.Image) {
	offsetX := 0
	offsetY := 240
	f.data.Draw(screen, offsetX, offsetY, f.currentDepth)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(f.playerX*GridSize), float64(-(f.playerY+1)*GridSize))
	op.GeoM.Translate(float64(offsetX), float64(offsetY))
	screen.DrawImage(f.playerImage, op)
}
