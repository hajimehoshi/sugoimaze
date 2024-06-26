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
	LevelSugoi
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
	wall     bool
	ladder   bool
	upward   bool
	downward bool
	sw       bool
	goal     bool
	color    int // 0 is no color. 1 and more is depth+1.
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

	tiles [][]tile

	tilesImage                  *ebiten.Image
	playerImage                 *ebiten.Image
	wallImage                   *ebiten.Image
	ladderImage                 *ebiten.Image
	goalImage                   *ebiten.Image
	upwardImage                 *ebiten.Image
	downwardImage               *ebiten.Image
	upwardDisabledImage         *ebiten.Image
	downwardDisabledImage       *ebiten.Image
	colorPassableWallImages     [4]*ebiten.Image
	colorUnpassableWallImages   [4]*ebiten.Image
	colorPassableLadderImages   [4]*ebiten.Image
	colorUnpassableLadderImages [4]*ebiten.Image
	switchImages                [4]*ebiten.Image
}

func NewFieldData(difficulty Difficulty) *FieldData {
	var width int
	var depth int
	var height int

	switch difficulty {
	case LevelEasy:
		width = 5
		height = 5
		depth = 2
	case LevelNormal:
		width = 8
		height = 8
		depth = 2
	case LevelHard:
		width = 11
		height = 11
		depth = 2
	case LevelSugoi:
		width = 14
		height = 14
		depth = 2 // TODO: Add more dimensions.
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

	var rooms [][][]room
	for {
		if rooms = f.generateWalls(); rooms != nil {
			break
		}
	}
	f.setTiles(rooms)

	img, err := png.Decode(bytes.NewReader(tilesPng))
	if err != nil {
		panic(err)
	}
	f.tilesImage = ebiten.NewImageFromImage(img)

	f.playerImage = f.tilesImage.SubImage(image.Rect(1*GridSize, 0*GridSize, 2*GridSize, 1*GridSize)).(*ebiten.Image)
	f.wallImage = f.tilesImage.SubImage(image.Rect(2*GridSize, 0*GridSize, 3*GridSize, 1*GridSize)).(*ebiten.Image)
	f.ladderImage = f.tilesImage.SubImage(image.Rect(3*GridSize, 0*GridSize, 4*GridSize, 1*GridSize)).(*ebiten.Image)
	f.goalImage = f.tilesImage.SubImage(image.Rect(4*GridSize, 0*GridSize, 5*GridSize, 1*GridSize)).(*ebiten.Image)
	f.upwardImage = f.tilesImage.SubImage(image.Rect(5*GridSize, 0*GridSize, 6*GridSize, 1*GridSize)).(*ebiten.Image)
	f.downwardImage = f.tilesImage.SubImage(image.Rect(6*GridSize, 0*GridSize, 7*GridSize, 1*GridSize)).(*ebiten.Image)
	f.upwardDisabledImage = f.tilesImage.SubImage(image.Rect(7*GridSize, 0*GridSize, 8*GridSize, 1*GridSize)).(*ebiten.Image)
	f.downwardDisabledImage = f.tilesImage.SubImage(image.Rect(8*GridSize, 0*GridSize, 9*GridSize, 1*GridSize)).(*ebiten.Image)
	for i := range f.colorPassableWallImages {
		f.colorPassableWallImages[i] = f.tilesImage.SubImage(image.Rect(0*GridSize, (i+1)*GridSize, 1*GridSize, (i+2)*GridSize)).(*ebiten.Image)
	}
	for i := range f.colorUnpassableWallImages {
		f.colorUnpassableWallImages[i] = f.tilesImage.SubImage(image.Rect(1*GridSize, (i+1)*GridSize, 2*GridSize, (i+2)*GridSize)).(*ebiten.Image)
	}
	for i := range f.colorPassableLadderImages {
		f.colorPassableLadderImages[i] = f.tilesImage.SubImage(image.Rect(4*GridSize, (i+1)*GridSize, 5*GridSize, (i+2)*GridSize)).(*ebiten.Image)
	}
	for i := range f.colorUnpassableLadderImages {
		f.colorUnpassableLadderImages[i] = f.tilesImage.SubImage(image.Rect(3*GridSize, (i+1)*GridSize, 4*GridSize, (i+2)*GridSize)).(*ebiten.Image)
	}
	for i := range f.switchImages {
		f.switchImages[i] = f.tilesImage.SubImage(image.Rect(2*GridSize, (i+1)*GridSize, 3*GridSize, (i+2)*GridSize)).(*ebiten.Image)
	}

	return f
}

func (f *FieldData) generateWalls() [][][]room {
	rooms := make([][][]room, f.depth)
	for z := range f.depth {
		rooms[z] = make([][]room, f.height)
		for y := 0; y < f.height; y++ {
			rooms[z][y] = make([]room, f.width)
		}
	}

	// Generate the correct path.
	x, y, z := f.startX, f.startY, f.startZ
	rooms[z][y][x].pathCount = 1
	newRooms := f.tryAddPathWithOneWay(rooms, x, y, z, func(x, y, z int, rooms [][][]room, count int) bool {
		return x == f.goalX && y == f.goalY && z == f.goalZ
	})
	if newRooms == nil {
		return nil
	}
	rooms = newRooms
	rooms[f.goalZ][f.goalY][f.goalX].wallY = wallPassable

	// Add branches.
	for !f.areEnoughRoomsVisited(rooms) {
		var startX, startY, startZ int
		for {
			startX, startY, startZ = rand.IntN(f.width), rand.IntN(f.height), rand.IntN(f.depth)
			if rooms[startZ][startY][startX].pathCount != 0 {
				break
			}
		}
		startCount := rooms[startZ][startY][startX].pathCount
		newRooms := f.tryAddPathWithOneWay(rooms, startX, startY, startZ, func(x, y, z int, rooms [][][]room, count int) bool {
			if x == startX && y == startY && z == startZ {
				return false
			}
			if rooms[z][y][x].pathCount == 0 {
				return false
			}
			// A branch must not be a shortcut.
			// Also, a good branch should go back to a position close to the start position.
			// Multiply a constant to make better branches.
			if startCount <= rooms[z][y][x].pathCount*5/4 {
				return false
			}
			return true
		})
		if newRooms == nil {
			continue
		}
		rooms = newRooms
	}

	return rooms
}

func (f *FieldData) tryAddPathWithOneWay(rooms [][][]room, x, y, z int, isGoal func(x, y, z int, rooms [][][]room, count int) bool) [][][]room {
	// Clone rooms.
	origRooms := rooms
	rooms = append([][][]room{}, origRooms...)
	for z := range f.depth {
		rooms[z] = append([][]room{}, origRooms[z]...)
		for y := range f.height {
			rooms[z][y] = append([]room{}, origRooms[z][y]...)
		}
	}

	var oneWayExists bool

	count := rooms[z][y][x].pathCount

	for !isGoal(x, y, z, rooms, count) {
		var goalReached bool
		var nextX, nextY, nextZ int
		var oneWay bool
		var found bool

	retry:
		for range 100 {
			origX, origY, origZ := x, y, z
			nextX, nextY, nextZ = x, y, z
			oneWay = false

			switch d := rand.IntN(12 + f.depth - 1); d {
			case 0, 1, 2:
				if nextX <= 0 {
					continue
				}
				nextX--
			case 3, 4, 5:
				if nextX >= f.width-1 {
					continue
				}
				nextX++
			case 6, 7, 8:
				if nextY <= 0 {
					continue
				}
				nextY--
			case 9, 10, 11:
				if nextY >= f.height-1 {
					continue
				}
				nextY++
			default:
				nextZ = (nextZ + (d - 12) + 1) % f.depth
			}

			// visited indicates whether the next room is already visited.
			var visited bool
			switch {
			case origZ != nextZ:
				for z := range f.depth {
					if z == origZ {
						continue
					}
					if rooms[nextZ][nextY][nextX].pathCount != 0 {
						visited = true
						break
					}
				}
			case origY != nextY:
				allWall := true
				allWallOrOneWay := true
				for z := range f.depth {
					if origY < nextY {
						// There is a conflicted one-way passage.
						if rooms[z][origY][origX].wallY == wallOneWayBackward {
							continue retry
						}
						if rooms[z][origY][origX].wallY != wallWall {
							allWall = false
							if rooms[z][origY][origX].wallY != wallOneWayForward {
								allWallOrOneWay = false
							}
						}
					}
					if origY > nextY {
						// There is a conflicted one-way passage.
						if rooms[z][nextY][nextX].wallY == wallOneWayForward {
							continue retry
						}
						if rooms[z][nextY][nextX].wallY != wallWall {
							allWall = false
							if rooms[z][nextY][nextX].wallY != wallOneWayBackward {
								allWallOrOneWay = false
							}
						}
					}
				}
				if allWall {
					oneWay = rand.IntN(5) == 0
				} else if allWallOrOneWay {
					oneWay = true
				}
				if allWallOrOneWay {
					// A branch must have a one-way passage.
					// Just before the goal, the passage should be one-way so that branches are created more easily.
					if isGoal(nextX, nextY, nextZ, rooms, count+1) {
						oneWay = true
						goalReached = true
						found = true
						break
					}
				}
				fallthrough
			default:
				if rooms[nextZ][nextY][nextX].pathCount != 0 {
					visited = true
				}
			}

			if !visited {
				found = true
				break
			}

			if isGoal(nextX, nextY, nextZ, rooms, count+1) {
				goalReached = true
				found = true
				break
			}
		}

		// Give up when no new path is created.
		if !found {
			return nil
		}

		if oneWay {
			oneWayExists = true
		}

		switch {
		case x < nextX:
			rooms[z][y][x].wallX = wallPassable
		case x > nextX:
			rooms[z][y][nextX].wallX = wallPassable
		case y < nextY:
			if oneWay {
				for z := range f.depth {
					if z == nextZ {
						rooms[z][y][x].wallY = wallOneWayForward
						continue
					}
					if rooms[z][y][x].wallY == wallOneWayBackward {
						panic("not reached")
					}
					if rooms[z][y][x].wallY == wallPassable {
						panic("not reached")
					}
				}
			} else {
				for z := range f.depth {
					if z == nextZ {
						rooms[z][y][x].wallY = wallPassable
					}
					if rooms[z][y][x].wallY == wallOneWayForward {
						panic("not reached")
					}
					if rooms[z][y][x].wallY == wallOneWayBackward {
						panic("not reached")
					}
				}
			}
		case y > nextY:
			if oneWay {
				for z := range f.depth {
					if z == nextZ {
						rooms[z][nextY][x].wallY = wallOneWayBackward
						continue
					}
					if rooms[z][nextY][x].wallY == wallOneWayForward {
						panic("not reached")
					}
					if rooms[z][nextY][x].wallY == wallPassable {
						panic("not reached")
					}
				}
			} else {
				for z := range f.depth {
					if z == nextZ {
						rooms[z][nextY][x].wallY = wallPassable
					}
					if rooms[z][nextY][x].wallY == wallOneWayForward {
						panic("not reached")
					}
					if rooms[z][nextY][x].wallY == wallOneWayBackward {
						panic("not reached")
					}
				}
			}
		case z != nextZ:
			for z := 0; z < f.depth-1; z++ {
				rooms[z][y][x].wallZ = wallPassable
			}
		}

		count++
		if z != nextZ {
			origZ := z
			for z := range f.depth {
				rooms[z][nextY][nextX].pathCount = count + abs(origZ-z)
			}
		} else {
			rooms[nextZ][nextY][nextX].pathCount = count
		}

		if goalReached {
			break
		}

		x, y, z = nextX, nextY, nextZ
	}

	if !oneWayExists {
		return nil
	}
	return rooms
}

func (f *FieldData) areEnoughRoomsVisited(rooms [][][]room) bool {
	var visited int
	threshold := (f.width * f.height * f.depth) * 8 / 10
	for z := range f.depth {
		for y := range f.height {
			for x := range f.width {
				if rooms[z][y][x].pathCount > 0 {
					visited++
					if visited >= threshold {
						return true
					}
				}
			}
		}
	}
	return false
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

func (f *FieldData) setTiles(rooms [][][]room) {
	width := f.width*roomXGridCount + 1
	height := f.height*roomYGridCount + 2

	f.tiles = make([][]tile, height)
	for y := range f.tiles {
		f.tiles[y] = make([]tile, width)
	}

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
	f.tiles[height-1][width-1].wall = true

	// Set the goal.
	f.tiles[height-1][width-6].goal = true

	for y := range f.height {
		for x := range f.width {
			f.setTilesForRoom(rooms, x, y)
		}
	}
}

func (f *FieldData) setTilesForRoom(rooms [][][]room, roomX, roomY int) {
	const (
		edgeOffsetX = 1
		edgeOffsetY = 1
	)

	allPassableOrOneWayX := true
	allPassableOrOneWayY := true
	allWallX := true
	allWallY := true
	for z := range f.depth {
		room := rooms[z][roomY][roomX]
		if room.wallX != wallPassable && room.wallX != wallOneWayForward && room.wallX != wallOneWayBackward {
			allPassableOrOneWayX = false
		}
		if room.wallX != wallWall {
			allWallX = false
		}
		if room.wallY != wallPassable && room.wallY != wallOneWayForward && room.wallY != wallOneWayBackward {
			allPassableOrOneWayY = false
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
	} else if !allPassableOrOneWayX {
		for z := range f.depth {
			room := rooms[z][roomY][roomX]
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
		if allPassableOrOneWayY {
			wallY := wallWall
			// Check all the wallY are the same.
			for z := range f.depth {
				room := rooms[z][roomY][roomX]
				if wallY == wallWall {
					wallY = room.wallY
					continue
				}
				if wallY != room.wallY {
					panic("not reached")
				}
			}
			for j := range roomYGridCount {
				y := roomY*roomYGridCount + j + edgeOffsetY
				f.tiles[y][x].ladder = true
				if wallY == wallOneWayForward {
					f.tiles[y][x].upward = true
				}
				if wallY == wallOneWayBackward {
					f.tiles[y][x].downward = true
				}
			}
		} else {
			for z := range f.depth {
				room := rooms[z][roomY][roomX]
				if room.wallY == wallWall {
					continue
				}
				for j := range roomYGridCount {
					y := roomY*roomYGridCount + j + edgeOffsetY
					f.tiles[y][x].ladder = true
					f.tiles[y][x].color = z + 1
					if room.wallY == wallOneWayForward {
						f.tiles[y][x].upward = true
					}
					if room.wallY == wallOneWayBackward {
						f.tiles[y][x].downward = true
					}
				}
			}
		}
	}

	if rooms[0][roomY][roomX].wallZ != wallWall {
		x := roomX*roomXGridCount + 3 + edgeOffsetX
		y := roomY*roomYGridCount + edgeOffsetY
		f.tiles[y][x].sw = true
	}
}

func (f *FieldData) hasSwitch(x, y int) bool {
	return f.tiles[y][x].sw
}

func (f *FieldData) passable(nextX, nextY int, prevX, prevY int, currentDepth int) bool {
	if nextY < 0 || len(f.tiles) <= nextY || nextX < 0 || len(f.tiles[nextY]) <= nextX {
		return false
	}
	if !f.canBeInTile(nextX, nextY, currentDepth) {
		return false
	}
	if !f.canStandOnTile(nextX, nextY-1, currentDepth) {
		return false
	}
	if nextY > prevY && !f.canGoUp(nextX, nextY) {
		return false
	}
	if nextY < prevY && !f.canGoDown(nextX, nextY) {
		return false
	}
	return true
}

func (f *FieldData) canBeInTile(x, y int, currentDepth int) bool {
	if y < 0 || len(f.tiles) <= y || x < 0 || len(f.tiles[y]) <= x {
		return false
	}
	t := f.tiles[y][x]
	if t.ladder {
		if t.color == 0 || t.color-1 == currentDepth {
			return true
		}
	}
	if t.wall {
		if t.color == 0 || t.color-1 != currentDepth {
			return false
		}
	}
	return true
}

func (f *FieldData) canStandOnTile(x, y int, currentDepth int) bool {
	if y < 0 || len(f.tiles) <= y || x < 0 || len(f.tiles[y]) <= x {
		return false
	}
	t := f.tiles[y][x]
	if t.ladder {
		if t.color == 0 || t.color-1 == currentDepth {
			return true
		}
	}
	if t.wall {
		if t.color == 0 || t.color-1 != currentDepth {
			return true
		}
	}
	return false
}

func (f *FieldData) canGoUp(x, y int) bool {
	if y < 0 || len(f.tiles) <= y || x < 0 || len(f.tiles[y]) <= x {
		return false
	}
	t := f.tiles[y][x]
	return !t.downward
}

func (f *FieldData) canGoDown(x, y int) bool {
	if y < 0 || len(f.tiles) <= y || x < 0 || len(f.tiles[y]) <= x {
		return false
	}
	t := f.tiles[y][x]
	return !t.upward
}

func (f *FieldData) isGoal(x, y int) bool {
	return f.tiles[y][x].goal
}

func (f *FieldData) floorNumber(y int) int {
	return (y-1)/roomYGridCount + 1
}

func (f *FieldData) floorCount() int {
	return f.height + 1
}

func (f *FieldData) Draw(screen *ebiten.Image, offsetX, offsetY int, currentDepth int) {
	for y := range f.tiles {
		for x := range f.tiles[y] {
			dx := x*GridSize + offsetX
			dy := -(y+1)*GridSize + offsetY

			if dx < -GridSize || dx >= screen.Bounds().Dx() || dy < -GridSize || dy >= screen.Bounds().Dy() {
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(dx), float64(dy))

			t := f.tiles[y][x]
			if t.wall {
				img := f.wallImage
				if t.color != 0 && !t.ladder {
					d := t.color - 1
					if currentDepth == d {
						img = f.colorPassableWallImages[d]
					} else {
						img = f.colorUnpassableWallImages[d]
					}
				}
				screen.DrawImage(img, op)
			}
			if t.ladder {
				d := t.color - 1
				img := f.ladderImage
				if t.color != 0 {
					if currentDepth == d {
						img = f.colorPassableLadderImages[d]
					} else {
						img = f.colorUnpassableLadderImages[d]
					}
				}
				screen.DrawImage(img, op)
				if t.upward {
					if t.color == 0 || currentDepth == d {
						screen.DrawImage(f.upwardImage, op)
					} else {
						screen.DrawImage(f.upwardDisabledImage, op)
					}
				}
				if t.downward {
					if t.color == 0 || currentDepth == d {
						screen.DrawImage(f.downwardImage, op)
					} else {
						screen.DrawImage(f.downwardDisabledImage, op)
					}
				}
			}
			if t.sw {
				switchImage := f.switchImages[currentDepth]
				screen.DrawImage(switchImage, op)
			}
			if t.goal {
				screen.DrawImage(f.goalImage, op)
			}
		}
	}
}
