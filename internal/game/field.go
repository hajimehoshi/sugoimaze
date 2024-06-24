// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package game

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Field struct {
	data         *FieldData
	playerX      int
	playerY      int
	dx           int
	dy           int
	currentDepth int
	goalReached  bool

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

func (f *Field) IsGoalReached() bool {
	return f.goalReached
}

func (f *Field) Update() {
	if f.goalReached {
		return
	}

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
		if f.data.isGoal(f.playerX, f.playerY) {
			f.goalReached = true
		}
		return
	}

	prevX, prevY := f.playerX, f.playerY
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if f.data.hasSwitch(prevX, prevY) {
			f.currentDepth++
			f.currentDepth %= f.data.depth
		}
	}

	nextX, nextY := prevX, prevY
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		nextY++
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		nextY--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		nextX--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		nextX++
	}
	if !f.data.passable(nextX, nextY, prevX, prevY, f.currentDepth) {
		return
	}
	if nextX > f.playerX {
		f.dx = v
	}
	if nextX < f.playerX {
		f.dx = -v
	}
	if nextY > f.playerY {
		f.dy = v
	}
	if nextY < f.playerY {
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

	ebitenutil.DebugPrint(screen, fmt.Sprintf("%dF / %dF", f.data.floorNumber(f.playerY), f.data.floorCount()))
}
