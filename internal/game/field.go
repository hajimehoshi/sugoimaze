// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package game

import (
	"image/color"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/oklab"
)

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
		depth = 3
	case LevelHard:
		width = 13
		height = 15
		depth = 4
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
			switch d := rand.IntN(4 + (f.depth - 1)); d {
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

	// TODO: Add branches

	return true
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (f *Field) Draw(screen *ebiten.Image, offsetX, offsetY int) {
	for y := 0; y < f.height; y++ {
		for x := 0; x < f.width; x++ {
			f.drawRoom(screen, offsetX, offsetY, x, y)
		}
	}
}

func (f *Field) drawRoom(screen *ebiten.Image, offsetX, offsetY int, roomX, roomY int) {
	const (
		roomWidth  = 32
		roomHeight = 32
	)

	x := roomX*roomWidth + offsetX
	y := -roomY*roomHeight + offsetY

	// https://ja.wikipedia.org/wiki/PCCS
	var (
		red    = oklab.OklchModel.Convert(color.RGBA{R: 0xee, G: 0x00, B: 0x26, A: 0xff}).(oklab.Oklch)
		blue   = oklab.OklchModel.Convert(color.RGBA{R: 0x0f, G: 0x21, B: 0x8b, A: 0xff}).(oklab.Oklch)
		green  = oklab.OklchModel.Convert(color.RGBA{R: 0x33, G: 0xa2, B: 0x3d, A: 0xff}).(oklab.Oklch)
		yellow = oklab.OklchModel.Convert(color.RGBA{R: 0xff, G: 0xe6, B: 0x00, A: 0xff}).(oklab.Oklch)
	)
	wallColors := []color.Color{red, blue, green, yellow}

	allWallX := true
	allWallY := true
	for z := 0; z < f.depth; z++ {
		room := f.rooms[z][roomY][roomX]
		if room.wallX != wallWall {
			allWallX = false
		}
		if room.wallY != wallWall {
			allWallY = false
		}
	}
	if allWallX {
		vector.StrokeLine(screen, float32(x+roomWidth), float32(y), float32(x+roomWidth), float32(y-roomHeight), 1, color.White, false)
	}
	if allWallY {
		vector.StrokeLine(screen, float32(x), float32(y-roomHeight), float32(x+roomWidth), float32(y-roomHeight), 1, color.White, false)
	}

	for z := 0; z < f.depth; z++ {
		room := f.rooms[z][roomY][roomX]
		if !allWallX && room.wallX == wallWall {
			vector.StrokeLine(screen, float32(x+roomWidth+z), float32(y), float32(x+roomWidth+z), float32(y-roomHeight), 1, wallColors[z], false)
		}
		if !allWallY && room.wallY == wallWall {
			vector.StrokeLine(screen, float32(x), float32(y-roomHeight-z), float32(x+roomWidth), float32(y-roomHeight-z), 1, wallColors[z], false)
		}
	}

	if f.rooms[0][roomY][roomX].wallZ != wallWall {
		vector.DrawFilledRect(screen, float32(x), float32(y-roomHeight), float32(roomWidth/2), float32(roomHeight/2), color.White, false)
	}
}
