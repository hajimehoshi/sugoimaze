// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package game

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed tiles.png
var tilesPng []byte

type Difficulty int

const (
	LevelEasy Difficulty = iota
	LevelNormal
	LevelHard
)

type wall int

const (
	wallWall wall = iota
	wallPassable
	wallOneWayForward
	wallOneWayBackward
)

type room struct {
	wallX     wall
	wallY     wall
	wallZ     wall
	pathCount int
}

type Field struct {
	width  int
	height int
	depth  int
	startX int
	startY int
	startZ int
	goalX  int
	goalY  int
	goalZ  int

	rooms [][][]room

	tilesImage  *ebiten.Image
	playerImage *ebiten.Image
	wallImage   *ebiten.Image
	ladderImage *ebiten.Image

	currentDepth int
}

func NewField(difficulty Difficulty) *Field {
	var width int
	var depth int
	var height int

	switch difficulty {
	case LevelEasy:
		width = 9
		height = 5
		depth = 2
	case LevelNormal:
		width = 11
		height = 10
		depth = 2 // 3 or more is impossible to represent in 2D.
	case LevelHard:
		width = 13
		height = 15
		depth = 2
	default:
		panic("not reached")
	}

	f := &Field{
		width:  width,
		height: height,
		depth:  depth,
		startX: 0,
		startY: 0,
		startZ: 0,
		goalX:  width - 1,
		goalY:  height - 1,
		goalZ:  depth - 1,
	}

	for !f.generateWalls() {
	}

	img, err := png.Decode(bytes.NewReader(tilesPng))
	if err != nil {
		panic(err)
	}
	f.tilesImage = ebiten.NewImageFromImage(img)

	f.playerImage = f.tilesImage.SubImage(image.Rect(1*GridSize, 0*GridSize, 2*GridSize, 1*GridSize)).(*ebiten.Image)
	f.wallImage = f.tilesImage.SubImage(image.Rect(2*GridSize, 0*GridSize, 3*GridSize, 1*GridSize)).(*ebiten.Image)
	f.ladderImage = f.tilesImage.SubImage(image.Rect(3*GridSize, 0*GridSize, 4*GridSize, 1*GridSize)).(*ebiten.Image)

	return f
}

func (f *Field) generateWalls() bool {
	f.rooms = make([][][]room, f.depth)
	for z := 0; z < f.depth; z++ {
		f.rooms[z] = make([][]room, f.height)
		for y := 0; y < f.height; y++ {
			f.rooms[z][y] = make([]room, f.width)
		}
	}

	// Generate the correct path.
	x, y, z := f.startX, f.startY, f.startZ
	count := 1
	f.rooms[z][y][x].pathCount = count
	for x != f.goalX || y != f.goalY || z != f.goalZ {
		var nextX, nextY, nextZ int

		for i := 0; i < 100; i++ {
			x, y, z := x, y, z

			var zChanged bool
			switch d := rand.IntN(4 + f.depth); d {
			case 0:
				if x <= 0 {
					continue
				}
				x--
			case 1:
				if x >= f.width-1 {
					continue
				}
				x++
			case 2:
				if y <= 0 {
					continue
				}
				y--
				// TODO: One way
			case 3:
				if y >= f.height-1 {
					continue
				}
				y++
			default:
				z = (z + (d - 4)) % f.depth
				zChanged = true
			}

			// The next room is already visited.
			if zChanged {
				for z := 0; z < f.depth; z++ {
					if f.rooms[z][y][x].pathCount != 0 {
						continue
					}
				}
			} else {
				if f.rooms[z][y][x].pathCount != 0 {
					continue
				}
			}

			nextX, nextY, nextZ = x, y, z
			break
		}

		if nextX == x && nextY == y && nextZ == z {
			return false
		}

		switch {
		case x < nextX:
			f.rooms[z][y][x].wallX = wallPassable
		case x > nextX:
			f.rooms[z][y][x-1].wallX = wallPassable
		case y < nextY:
			f.rooms[z][y][x].wallY = wallPassable
		case y > nextY:
			f.rooms[z][y-1][x].wallY = wallPassable
		case z != nextZ:
			for z := 0; z < f.depth-1; z++ {
				f.rooms[z][y][x].wallZ = wallPassable
			}
		}

		count++
		if z != nextZ {
			origZ := z
			for z := 0; z < f.depth; z++ {
				f.rooms[z][nextY][nextX].pathCount = count + abs(origZ-z)
			}
		} else {
			f.rooms[nextZ][nextY][nextX].pathCount = count
		}
		x, y, z = nextX, nextY, nextZ
	}

	f.rooms[f.goalZ][f.goalY][f.goalX].wallY = wallPassable

	// TODO: Add branches

	return true
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

const GridSize = 16

const (
	roomXGridCount = 10
	roomYGridCount = 4
)

func (f *Field) Draw(screen *ebiten.Image, offsetX, offsetY int) {
	// Draw the outside walls.
	for y := 0; y < f.height; y++ {
		for j := 0; j < roomYGridCount; j++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(0, float64(-(y*roomYGridCount+j)*GridSize))
			op.GeoM.Translate(float64(offsetX), float64(offsetY))
			op.GeoM.Translate(0, -GridSize)
			op.GeoM.Translate(0, -GridSize)
			screen.DrawImage(f.wallImage, op)
		}
	}
	for x := 0; x < f.width; x++ {
		for i := 0; i < roomXGridCount; i++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64((x*roomXGridCount+i)*GridSize), 0)
			op.GeoM.Translate(float64(offsetX), float64(offsetY))
			op.GeoM.Translate(0, -GridSize)
			screen.DrawImage(f.wallImage, op)
		}
	}

	// Draw the rooms.
	for y := 0; y < f.height; y++ {
		for x := 0; x < f.width; x++ {
			f.drawRoom(screen, offsetX+GridSize, offsetY-GridSize, x, y)
		}
	}
}

func (f *Field) drawRoom(screen *ebiten.Image, offsetX, offsetY int, roomX, roomY int) {
	allPassableX := true
	allPassableY := true
	allWallX := true
	allWallY := true
	for z := 0; z < f.depth; z++ {
		room := f.rooms[z][roomY][roomX]
		if room.wallX != wallPassable {
			allPassableX = false
		}
		if room.wallX != wallWall {
			allWallX = false
		}
		if room.wallY != wallPassable {
			allPassableY = false
		}
		if room.wallY != wallWall {
			allWallY = false
		}
	}
	if allWallX {
		for j := 1; j < roomYGridCount; j++ {
			op := &ebiten.DrawImageOptions{}
			x := roomX*roomXGridCount + 9
			y := -(roomY+1)*roomYGridCount + j
			op.GeoM.Translate(float64(x*GridSize), float64(y*GridSize))
			op.GeoM.Translate(float64(offsetX), float64(offsetY))
			screen.DrawImage(f.wallImage, op)
		}
	} else if !allPassableX {
		for z := 0; z < f.depth; z++ {
			room := f.rooms[z][roomY][roomX]
			if room.wallX == wallWall {
				continue
			}
			var colorWallImage *ebiten.Image
			imgY := 1 + z
			if f.currentDepth == z {
				// Passable
				colorWallImage = f.tilesImage.SubImage(image.Rect(0*GridSize, imgY*GridSize, 1*GridSize, (imgY+1)*GridSize)).(*ebiten.Image)
			} else {
				// Wall
				colorWallImage = f.tilesImage.SubImage(image.Rect(1*GridSize, imgY*GridSize, 2*GridSize, (imgY+1)*GridSize)).(*ebiten.Image)
			}
			for j := 1; j < roomYGridCount; j++ {
				op := &ebiten.DrawImageOptions{}
				x := roomX*roomXGridCount + 5 + z
				y := -(roomY+1)*roomYGridCount + j
				op.GeoM.Translate(float64(x*GridSize), float64(y*GridSize))
				op.GeoM.Translate(float64(offsetX), float64(offsetY))
				screen.DrawImage(colorWallImage, op)
			}
		}
	}

	for i := 0; i < roomXGridCount; i++ {
		op := &ebiten.DrawImageOptions{}
		x := roomX*roomXGridCount + i
		y := -(roomY + 1) * roomYGridCount
		op.GeoM.Translate(float64(x*GridSize), float64(y*GridSize))
		op.GeoM.Translate(float64(offsetX), float64(offsetY))
		screen.DrawImage(f.wallImage, op)
	}
	if !allWallY {
		x := roomX*roomXGridCount + 1 + (roomY % 2)
		if allPassableY {
			for j := 0; j < roomYGridCount; j++ {
				op := &ebiten.DrawImageOptions{}
				y := -(roomY+1)*roomYGridCount + j
				op.GeoM.Translate(float64(x*GridSize), float64(y*GridSize))
				op.GeoM.Translate(float64(offsetX), float64(offsetY))
				screen.DrawImage(f.ladderImage, op)
			}
		} else {
			for z := 0; z < f.depth; z++ {
				room := f.rooms[z][roomY][roomX]
				if room.wallY == wallWall {
					continue
				}
				var colorLadderImage *ebiten.Image
				imgY := 1 + z
				if f.currentDepth == z {
					// Passable
					colorLadderImage = f.tilesImage.SubImage(image.Rect(4*GridSize, imgY*GridSize, 5*GridSize, (imgY+1)*GridSize)).(*ebiten.Image)
				} else {
					// Non passable
					colorLadderImage = f.tilesImage.SubImage(image.Rect(3*GridSize, imgY*GridSize, 4*GridSize, (imgY+1)*GridSize)).(*ebiten.Image)
				}
				for j := 0; j < roomYGridCount; j++ {
					op := &ebiten.DrawImageOptions{}
					y := -(roomY+1)*roomYGridCount + j
					op.GeoM.Translate(float64(x*GridSize), float64(y*GridSize))
					op.GeoM.Translate(float64(offsetX), float64(offsetY))
					screen.DrawImage(colorLadderImage, op)
				}
			}
		}
	}

	if f.rooms[0][roomY][roomX].wallZ != wallWall {
		imgY := (1 + f.currentDepth)
		switchImage := f.tilesImage.SubImage(image.Rect(2*GridSize, imgY*GridSize, 3*GridSize, (imgY+1)*GridSize)).(*ebiten.Image)
		op := &ebiten.DrawImageOptions{}
		x := roomX*roomXGridCount + 4
		y := -(roomY+1)*roomYGridCount + 3
		op.GeoM.Translate(float64(x*GridSize), float64(y*GridSize))
		op.GeoM.Translate(float64(offsetX), float64(offsetY))
		screen.DrawImage(switchImage, op)
	}
}
