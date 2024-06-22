// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import "github.com/hajimehoshi/ebiten/v2"

type GameScene struct {
}

func (g *GameScene) Update(game GameContext) error {
	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
}
