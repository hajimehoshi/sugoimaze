// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/ebitenginegamejam2024/internal/game"
)

type GameContext interface {
	GoToGame(difficulty game.Difficulty)
	GoToTitle()
}

type Scene interface {
	Update(gameContext GameContext) error
	Draw(screen *ebiten.Image)
}

type Game struct {
	scene Scene
}

func NewGame() *Game {
	return &Game{
		scene: &TitleScene{},
	}
}

func (g *Game) Update() error {
	if err := g.scene.Update(g); err != nil {
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth / 2, outsideHeight / 2
}

func (g *Game) GoToGame(level game.Difficulty) {
	g.scene = NewGameScene(level)
}

func (g *Game) GoToTitle() {
	g.scene = &TitleScene{}
}

func main() {
	ebiten.SetWindowTitle("Ebitengine Game Jam 2024")
	ebiten.SetWindowSize(640, 640)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(NewGame()); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
