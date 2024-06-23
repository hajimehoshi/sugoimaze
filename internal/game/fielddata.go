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

type tile struct {
	wall   bool
	ladder bool
	sw     bool
	color  int // 0 is no color. 1 and more is depth+1.
}

type FieldData struct {
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
	tiles [][]tile

	tilesImage  *ebiten.Image
	playerImage *ebiten.Image
	wallImage   *ebiten.Image
	ladderImage *ebiten.Image
}

func NewFieldData(difficulty Difficulty) *FieldData {
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

	f := &FieldData{
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
	f.tiles = make([][]tile, f.height*roomYGridCount+2)
	for y := range f.height*roomYGridCount + 2 {
		f.tiles[y] = make([]tile, f.width*roomXGridCount+1)
	}

	img, err := png.Decode(bytes.NewReader(tilesPng))
	if err != nil {
		panic(err)
	}
	f.tilesImage = ebiten.NewImageFromImage(img)
	f.setTiles()

	f.playerImage = f.tilesImage.SubImage(image.Rect(1*GridSize, 0*GridSize, 2*GridSize, 1*GridSize)).(*ebiten.Image)
	f.wallImage = f.tilesImage.SubImage(image.Rect(2*GridSize, 0*GridSize, 3*GridSize, 1*GridSize)).(*ebiten.Image)
	f.ladderImage = f.tilesImage.SubImage(image.Rect(3*GridSize, 0*GridSize, 4*GridSize, 1*GridSize)).(*ebiten.Image)

	return f
}

func (f *FieldData) generateWalls() bool {
	f.rooms = make([][][]room, f.depth)
	for z := range f.depth {
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
				for z := range f.depth {
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
	roomXGridCount = 6
	roomYGridCount = 3
)

func (f *FieldData) setTiles() {
	for y := range f.tiles {
		for x := range f.tiles[y] {
			f.tiles[y][x] = tile{}
		}
	}

	// Set the outside walls.
	for x := range f.tiles[0] {
		f.tiles[0][x].wall = true

	}
	for y := range f.tiles {
		f.tiles[y][0].wall = true
	}

	for y := range f.height {
		for x := range f.width {
			f.setTilesForRoom(x, y)
		}
	}
}

func (f *FieldData) setTilesForRoom(roomX, roomY int) {
	const (
		edgeOffsetX = 1
		edgeOffsetY = 1
	)

	allPassableX := true
	allPassableY := true
	allWallX := true
	allWallY := true
	for z := range f.depth {
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
		for j := range roomYGridCount - 1 {
			x := roomX*roomXGridCount + roomXGridCount - 1 + edgeOffsetX
			y := roomY*roomYGridCount + j + edgeOffsetY
			f.tiles[y][x].wall = true
		}
	} else if !allPassableX {
		for z := range f.depth {
			room := f.rooms[z][roomY][roomX]
			if room.wallX == wallWall {
				continue
			}
			for j := range roomYGridCount - 1 {
				x := roomX*roomXGridCount + roomXGridCount - 1 + edgeOffsetX
				y := roomY*roomYGridCount + j + edgeOffsetY
				f.tiles[y][x].wall = true
				f.tiles[y][x].color = z + 1
			}
		}
	}

	for i := range roomXGridCount {
		x := roomX*roomXGridCount + i + edgeOffsetX
		y := (roomY+1)*roomYGridCount - 1 + edgeOffsetY
		f.tiles[y][x].wall = true
	}
	if !allWallY {
		x := roomX*roomXGridCount + 1 + (roomY % 2) + edgeOffsetX
		if allPassableY {
			for j := range roomYGridCount {
				y := roomY*roomYGridCount + j + edgeOffsetY
				f.tiles[y][x].ladder = true
			}
		} else {
			for z := range f.depth {
				room := f.rooms[z][roomY][roomX]
				if room.wallY == wallWall {
					continue
				}
				for j := range roomYGridCount {
					y := roomY*roomYGridCount + j + edgeOffsetY
					f.tiles[y][x].ladder = true
					f.tiles[y][x].color = z + 1
				}
			}
		}
	}

	if f.rooms[0][roomY][roomX].wallZ != wallWall {
		x := roomX*roomXGridCount + 3 + edgeOffsetX
		y := roomY*roomYGridCount + edgeOffsetY
		f.tiles[y][x].sw = true
	}
}

func (f *FieldData) hasSwitch(x, y int) bool {
	return f.tiles[y][x].sw
}

func (f *FieldData) passable(x, y int, currentDepth int) bool {
	if y < 0 || len(f.tiles) <= y || x < 0 || len(f.tiles[y]) <= x {
		return false
	}
	t := f.tiles[y][x]
	if t.wall {
		if t.color == 0 || t.color-1 != currentDepth {
			return false
		}
	}

	if t.ladder {
		if t.color == 0 || t.color-1 == currentDepth {
			return true
		}
	}

	if !f.tiles[y-1][x].wall {
		return false
	}
	return true
}

func (f *FieldData) Draw(screen *ebiten.Image, offsetX, offsetY int, currentDepth int) {
	for y := range f.tiles {
		for x := range f.tiles[y] {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*GridSize), float64(-(y+1)*GridSize))
			op.GeoM.Translate(float64(offsetX), float64(offsetY))

			t := f.tiles[y][x]
			if t.wall {
				img := f.wallImage
				if t.color != 0 && !t.ladder {
					d := t.color - 1
					var imgX int
					if currentDepth == d {
						imgX = 0
					} else {
						imgX = 1
					}
					imgY := 1 + d
					img = f.tilesImage.SubImage(image.Rect(imgX*GridSize, imgY*GridSize, (imgX+1)*GridSize, (imgY+1)*GridSize)).(*ebiten.Image)
				}
				screen.DrawImage(img, op)
			}
			if t.ladder {
				img := f.ladderImage
				if t.color != 0 {
					d := t.color - 1
					var imgX int
					if currentDepth == d {
						imgX = 4
					} else {
						imgX = 3
					}
					imgY := 1 + d
					img = f.tilesImage.SubImage(image.Rect(imgX*GridSize, imgY*GridSize, (imgX+1)*GridSize, (imgY+1)*GridSize)).(*ebiten.Image)
				}
				screen.DrawImage(img, op)
			}
			if f.tiles[y][x].sw {
				imgY := 1 + currentDepth
				switchImage := f.tilesImage.SubImage(image.Rect(2*GridSize, imgY*GridSize, 3*GridSize, (imgY+1)*GridSize)).(*ebiten.Image)
				screen.DrawImage(switchImage, op)
			}
		}
	}
}
