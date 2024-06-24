// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	game "github.com/hajimehoshi/ebitenginegamejam2024/internal/game"
)

type GameScene struct {
	difficulty game.Difficulty
	field      *game.Field
	fieldCh    chan *game.Field
}

func NewGameScene(difficulty game.Difficulty) *GameScene {
	return &GameScene{
		difficulty: difficulty,
	}
}

func (g *GameScene) Update(gameContext GameContext) error {
	if g.field == nil && g.fieldCh == nil {
		g.fieldCh = make(chan *game.Field)
		go func() {
			g.fieldCh <- game.NewField(g.difficulty)
			close(g.fieldCh)
			g.fieldCh = nil
		}()
	}
	select {
	case f := <-g.fieldCh:
		g.field = f
	default:
	}
	if g.field == nil {
		return nil
	}

	g.field.Update()
	if g.field.IsGoalReached() {
		if len(inpututil.AppendJustPressedKeys(nil)) > 0 {
			gameContext.GoToTitle()
		}
	}

	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

	if g.field == nil {
		ebitenutil.DebugPrint(screen, "Generating a field...")
		return
	}
	g.field.Draw(screen)

	if g.field.IsGoalReached() {
		ebitenutil.DebugPrint(screen, "\n\nGOAL!")
	}
}
