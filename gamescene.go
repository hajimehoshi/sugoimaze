// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	game "github.com/hajimehoshi/sugoimaze/internal/game"
)

type GameScene struct {
	bgmStarted bool
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
	if !g.bgmStarted && g.field != nil {
		gameContext.PlayBGM("game")
		g.bgmStarted = true
	}

	if g.field == nil && g.fieldCh == nil {
		g.fieldCh = make(chan *game.Field)
		// Wait one second at least to show the message.
		t := time.NewTimer(time.Second)
		go func() {
			f := game.NewField(g.difficulty)
			<-t.C
			t.Stop()
			g.fieldCh <- f
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
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			gameContext.GoToTitle()
		}
	}

	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

	if g.field == nil {
		ebitenutil.DebugPrint(screen, "Currently under construction.\nPlease wait a moment.")
		return
	}
	g.field.Draw(screen)

	if g.field.IsGoalReached() {
		ebitenutil.DebugPrint(screen, "\n\nGOAL!")
	}
}
