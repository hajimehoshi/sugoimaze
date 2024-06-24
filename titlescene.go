// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	gamepkg "github.com/hajimehoshi/ebitenginegamejam2024/internal/game"
)

type TitleScene struct {
	cursorIndex int
}

func (t *TitleScene) Update(game GameContext) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		t.cursorIndex++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		t.cursorIndex--
	}
	if t.cursorIndex < 0 {
		t.cursorIndex = 0
	}
	if t.cursorIndex > 2 {
		t.cursorIndex = 2
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		var difficulty gamepkg.Difficulty
		switch t.cursorIndex {
		case 0:
			difficulty = gamepkg.LevelEasy
		case 1:
			difficulty = gamepkg.LevelNormal
		case 2:
			difficulty = gamepkg.LevelHard
		}
		game.GoToGame(difficulty)
	}
	return nil
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
	msg := "Gopher Maze Bldg.\n\n"
	for i, level := range []string{"Easy", "Normal", "Hard"} {
		if i == t.cursorIndex {
			msg += "-> "
		} else {
			msg += "   "
		}
		msg += level + "\n"
	}
	ebitenutil.DebugPrint(screen, msg)
}
