// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type GameContext interface {
	GoToGame()
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
	return 640, 480
}

func (g *Game) GoToGame() {
	g.scene = &GameScene{}
}

func main() {
	ebiten.SetWindowTitle("Ebitengine Game Jam 2024")
	if err := ebiten.RunGame(NewGame()); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
