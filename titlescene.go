// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	gamepkg "github.com/hajimehoshi/sugoimaze/internal/game"
)

type TitleScene struct {
	cursorIndex int
}

func (t *TitleScene) Update(game GameContext) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		t.cursorIndex++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		t.cursorIndex--
	}
	if t.cursorIndex < 0 {
		t.cursorIndex = 0
	}
	if t.cursorIndex > int(gamepkg.LevelSugoi) {
		t.cursorIndex = int(gamepkg.LevelSugoi)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		game.GoToGame(gamepkg.Difficulty(t.cursorIndex))
	}
	return nil
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
	msg := "The Sugoi Maze Building\n\n"
	for i, difficulty := range []gamepkg.Difficulty{gamepkg.LevelTutorial, gamepkg.LevelEasy, gamepkg.LevelNormal, gamepkg.LevelHard, gamepkg.LevelSugoi} {
		if i == t.cursorIndex {
			msg += " -> "
		} else {
			msg += "    "
		}
		msg += difficulty.String() + "\n"
	}
	msg += `
Controls:
  - Arrow keys, WASD: Move
  - Space, Enter:     Toggle switches, etc.
`
	ebitenutil.DebugPrint(screen, msg)
}
