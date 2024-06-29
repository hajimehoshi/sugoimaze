// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"

	"github.com/hajimehoshi/sugoimaze/internal/game"
)

//go:embed game.ogg
var gameOgg []byte

type GameContext interface {
	PlayBGM(name string) error
	StopBGM()
	GoToGame(difficulty game.Difficulty)
	GoToTitle()
}

type Scene interface {
	Update(gameContext GameContext) error
	Draw(screen *ebiten.Image)
}

type Game struct {
	scene            Scene
	audioContext     *audio.Context
	bgmPlayers       map[string]*audio.Player
	currentBGMPlayer *audio.Player
}

func NewGame() *Game {
	return &Game{
		scene:        &TitleScene{},
		audioContext: audio.NewContext(48000),
	}
}

func (g *Game) AudioContext() *audio.Context {
	return g.audioContext
}

func (g *Game) PlayBGM(name string) error {
	player, ok := g.bgmPlayers[name]
	if ok {
		player.Play()
		return nil
	}
	if g.bgmPlayers == nil {
		g.bgmPlayers = map[string]*audio.Player{}
	}
	if name != "game" {
		return fmt.Errorf("sugoimaze: unknown BGM name: %s", name)
	}
	stream, err := vorbis.DecodeWithoutResampling(bytes.NewReader(gameOgg))
	if err != nil {
		return err
	}
	player, err = g.audioContext.NewPlayer(stream)
	if err != nil {
		return err
	}
	g.bgmPlayers[name] = player
	g.currentBGMPlayer = player
	player.Play()
	return nil
}

func (g *Game) StopBGM() {
	if g.currentBGMPlayer != nil {
		g.currentBGMPlayer.Pause()
		g.currentBGMPlayer.Rewind()
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
	ebiten.SetWindowTitle("The Sugoi Maze Building")
	ebiten.SetWindowSize(640, 640)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(NewGame()); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
