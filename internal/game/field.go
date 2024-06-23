// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Field struct {
	data *FieldData
}

func NewField(difficulty Difficulty) *Field {
	return &Field{
		data: NewFieldData(difficulty),
	}
}

func (f *Field) Draw(screen *ebiten.Image, offsetX, offsetY int) {
	f.data.Draw(screen, offsetX, offsetY)
}
