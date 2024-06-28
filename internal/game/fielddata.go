// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package game

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
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

type passage int

const (
	passageWall passage = iota
	passagePassable
	passageOneWayForward
	passageOneWayBackward
)

type room struct {
	passageX  passage
	passageY  passage
	passageZ  passage
	passageW  passage
	pathCount int
}

type tile struct {
	walls     []bool
	ladders   []bool
	upward    bool
	downward  bool
	switches  []bool
	door      bool
	doorUpper bool
	goal      bool

	// 0 is no color. 1 and more is depth+1.
	wallColors   []int
	ladderColors []int
	doorColor    int
}

type FieldData struct {
	width  int
	height int
	depth0 int
	depth1 int
	startX int
	startY int
	startZ int
	startW int
	goalX  int
	goalY  int
	goalZ  int
	goalW  int

	colorPalette [2]int

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
	doorImage                   *ebiten.Image
	colorPassableWallImages     [4]*ebiten.Image
	colorUnpassableWallImages   [4]*ebiten.Image
	colorPassableLadderImages   [4]*ebiten.Image
	colorUnpassableLadderImages [4]*ebiten.Image
	colorUpwardImage            [4]*ebiten.Image
	colorDownwardImage          [4]*ebiten.Image
	colorUpwardDisabledImage    [4]*ebiten.Image
	colorDownwardDisabledImage  [4]*ebiten.Image
	switchImages                [4]*ebiten.Image
	colorDoorImages             [4]*ebiten.Image
	colorDoorDisabledImages     [4]*ebiten.Image
}

func NewFieldData(difficulty Difficulty) *FieldData {
	var width int
	var height int
	var depth0 int
	var depth1 int

	switch difficulty {
	case LevelEasy:
		width = 5
		height = 5
		depth0 = 2
		depth1 = 1
	case LevelNormal:
		width = 8
		height = 8
		depth0 = 2
		depth1 = 1
	case LevelHard:
		width = 11
		height = 11
		depth0 = 2
		depth1 = 1
	case LevelSugoi:
		width = 14
		height = 14
		depth0 = 2
		depth1 = 2
	default:
		panic("not reached")
	}

	f := &FieldData{
		width:  width,
		height: height,
		depth0: depth0,
		depth1: depth1,
		startX: 0,
		startY: 0,
		startZ: 0,
		startW: 0,
		goalX:  width - 1,
		goalY:  height - 1,
		goalZ:  depth0 - 1,
		goalW:  depth1 - 1,
	}
	f.colorPalette = [2]int{1, 3}

	var rooms [][][][]room
	for {
		if rooms = f.generateRooms(); rooms != nil {
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
	f.doorImage = f.tilesImage.SubImage(image.Rect(0*GridSize, 5*GridSize, 1*GridSize, 7*GridSize)).(*ebiten.Image)
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
	for i := range f.colorUpwardImage {
		f.colorUpwardImage[i] = f.tilesImage.SubImage(image.Rect(5*GridSize, (i+1)*GridSize, 6*GridSize, (i+2)*GridSize)).(*ebiten.Image)
	}
	for i := range f.colorDownwardImage {
		f.colorDownwardImage[i] = f.tilesImage.SubImage(image.Rect(6*GridSize, (i+1)*GridSize, 7*GridSize, (i+2)*GridSize)).(*ebiten.Image)
	}
	for i := range f.colorUpwardDisabledImage {
		f.colorUpwardDisabledImage[i] = f.tilesImage.SubImage(image.Rect(7*GridSize, (i+1)*GridSize, 8*GridSize, (i+2)*GridSize)).(*ebiten.Image)
	}
	for i := range f.colorDownwardDisabledImage {
		f.colorDownwardDisabledImage[i] = f.tilesImage.SubImage(image.Rect(8*GridSize, (i+1)*GridSize, 9*GridSize, (i+2)*GridSize)).(*ebiten.Image)
	}
	for i := range f.switchImages {
		f.switchImages[i] = f.tilesImage.SubImage(image.Rect(2*GridSize, (i+1)*GridSize, 3*GridSize, (i+2)*GridSize)).(*ebiten.Image)
	}
	for i := range f.colorDoorImages {
		f.colorDoorImages[i] = f.tilesImage.SubImage(image.Rect((2*i+2)*GridSize, 5*GridSize, (2*i+3)*GridSize, 7*GridSize)).(*ebiten.Image)
	}
	for i := range f.colorDoorDisabledImages {
		f.colorDoorDisabledImages[i] = f.tilesImage.SubImage(image.Rect((2*i+1)*GridSize, 5*GridSize, (2*i+2)*GridSize, 7*GridSize)).(*ebiten.Image)
	}

	return f
}

func (f *FieldData) generateRooms() [][][][]room {
	rooms := make([][][][]room, f.depth1)
	for w := range f.depth1 {
		rooms[w] = make([][][]room, f.depth0)
		for z := range f.depth0 {
			rooms[w][z] = make([][]room, f.height)
			for y := 0; y < f.height; y++ {
				rooms[w][z][y] = make([]room, f.width)
			}
		}
	}

	// Generate the correct path.
	x, y, z, w := f.startX, f.startY, f.startZ, f.startW
	rooms[w][z][y][x].pathCount = 1
	newRooms := f.tryAddPathWithOneWay(rooms, x, y, z, w, func(x, y, z, w int, rooms [][][][]room, count int) bool {
		return x == f.goalX && y == f.goalY && z == f.goalZ && w == f.goalW
	})
	if newRooms == nil {
		return nil
	}
	rooms = newRooms
	rooms[f.goalW][f.goalZ][f.goalY][f.goalX].passageY = passagePassable

	// Add branches.
	var count int
	for !f.areEnoughRoomsVisited(rooms) {
		var startX, startY, startZ, startW int
		for {
			startX, startY, startZ, startW = rand.IntN(f.width), rand.IntN(f.height), rand.IntN(f.depth0), rand.IntN(f.depth1)
			if rooms[startW][startZ][startY][startX].pathCount != 0 {
				break
			}
		}
		startCount := rooms[startW][startZ][startY][startX].pathCount
		newRooms := f.tryAddPathWithOneWay(rooms, startX, startY, startZ, startW, func(x, y, z, w int, rooms [][][][]room, count int) bool {
			if x == startX && y == startY && z == startZ && w == startW {
				return false
			}
			if rooms[w][z][y][x].pathCount == 0 {
				return false
			}
			// A branch must not be a shortcut.
			// Also, a good branch should go back to a position close to the start position.
			// Multiply a constant to make better branches.
			if startCount <= rooms[w][z][y][x].pathCount*5/4 {
				return false
			}
			return true
		})
		if newRooms == nil {
			count++
			if count > 1000 {
				return nil
			}
			continue
		}
		rooms = newRooms
		count = 0
	}

	return rooms
}

func (f *FieldData) tryAddPathWithOneWay(rooms [][][][]room, x, y, z, w int, isGoal func(x, y, z, w int, rooms [][][][]room, count int) bool) [][][][]room {
	// Clone rooms.
	origRooms := rooms
	rooms = make([][][][]room, len(origRooms))
	for w := range f.depth1 {
		rooms[w] = make([][][]room, len(origRooms[w]))
		for z := range f.depth0 {
			rooms[w][z] = make([][]room, len(origRooms[w][z]))
			for y := range f.height {
				rooms[w][z][y] = append([]room{}, origRooms[w][z][y]...)
			}
		}
	}

	var oneWayExists bool

	count := rooms[w][z][y][x].pathCount

	for !isGoal(x, y, z, w, rooms, count) {
		var goalReached bool
		var nextX, nextY, nextZ, nextW int
		var oneWay bool
		var found bool

	retry:
		for range 100 {
			origX, origY, origZ, origW := x, y, z, w
			nextX, nextY, nextZ, nextW = x, y, z, w
			oneWay = false

			switch d := rand.IntN(12 + (f.depth0 - 1) + (f.depth1 - 1)); d {
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
			case 12:
				nextZ = (nextZ + 1) % f.depth0
			case 13:
				nextW = (nextW + 1) % f.depth1
			}

			// visited indicates whether the next room is already visited.
			var visited bool
			switch {
			case origZ != nextZ:
				for z := range f.depth0 {
					if z == origZ {
						continue
					}
					if rooms[nextW][nextZ][nextY][nextX].pathCount != 0 {
						visited = true
						break
					}
				}
			case origW != nextW:
				for w := range f.depth1 {
					if w == origW {
						continue
					}
					if rooms[nextW][nextZ][nextY][nextX].pathCount != 0 {
						visited = true
						break
					}
				}
			case origY != nextY:
				allWall := true
				allWallOrOneWay := true
				for z := range f.depth0 {
					if origY < nextY {
						// There is a conflicted one-way passage.
						if rooms[origW][z][origY][origX].passageY == passageOneWayBackward {
							continue retry
						}
						if rooms[origW][z][origY][origX].passageY != passageWall {
							allWall = false
							if rooms[origW][z][origY][origX].passageY != passageOneWayForward {
								allWallOrOneWay = false
							}
						}
					}
					if origY > nextY {
						// There is a conflicted one-way passage.
						if rooms[origW][z][nextY][nextX].passageY == passageOneWayForward {
							continue retry
						}
						if rooms[origW][z][nextY][nextX].passageY != passageWall {
							allWall = false
							if rooms[origW][z][nextY][nextX].passageY != passageOneWayBackward {
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
					if isGoal(nextX, nextY, nextZ, nextW, rooms, count+1) {
						oneWay = true
						goalReached = true
						found = true
						break
					}
				}
				fallthrough
			default:
				if rooms[nextW][nextZ][nextY][nextX].pathCount != 0 {
					visited = true
				}
			}

			if !visited {
				found = true
				break
			}

			if isGoal(nextX, nextY, nextZ, nextW, rooms, count+1) {
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
			rooms[w][z][y][x].passageX = passagePassable
		case x > nextX:
			rooms[w][z][y][nextX].passageX = passagePassable
		case y < nextY:
			if oneWay {
				for z := range f.depth0 {
					if z == nextZ && w == nextW {
						rooms[w][z][y][x].passageY = passageOneWayForward
						continue
					}
					if rooms[w][z][y][x].passageY == passageOneWayBackward {
						panic("not reached")
					}
					if rooms[w][z][y][x].passageY == passagePassable {
						panic("not reached")
					}
				}
			} else {
				for z := range f.depth0 {
					if z == nextZ && w == nextW {
						rooms[w][z][y][x].passageY = passagePassable
						continue
					}
					if rooms[w][z][y][x].passageY == passageOneWayForward {
						panic("not reached")
					}
					if rooms[w][z][y][x].passageY == passageOneWayBackward {
						panic("not reached")
					}
				}
			}
		case y > nextY:
			if oneWay {
				for z := range f.depth0 {
					if z == nextZ && w == nextW {
						rooms[w][z][nextY][x].passageY = passageOneWayBackward
						continue
					}
					if rooms[w][z][nextY][x].passageY == passageOneWayForward {
						panic("not reached")
					}
					if rooms[w][z][nextY][x].passageY == passagePassable {
						panic("not reached")
					}
				}
			} else {
				for z := range f.depth0 {
					if z == nextZ && w == nextW {
						rooms[w][z][nextY][x].passageY = passagePassable
						continue
					}
					if rooms[w][z][nextY][x].passageY == passageOneWayForward {
						panic("not reached")
					}
					if rooms[w][z][nextY][x].passageY == passageOneWayBackward {
						panic("not reached")
					}
				}
			}
		case z != nextZ:
			// The last Z's passage is always wall
			for z := range f.depth0 - 1 {
				rooms[w][z][y][x].passageZ = passagePassable
			}
		case w != nextW:
			// The last W's passage is always wall
			for w := range f.depth1 - 1 {
				rooms[w][z][y][x].passageW = passagePassable
			}
		}

		if z != nextZ {
			origZ := z
			for z := range f.depth0 {
				rooms[nextW][z][nextY][nextX].pathCount = count + abs(origZ-z)
			}
		} else if w != nextW {
			origW := w
			for w := range f.depth1 {
				rooms[w][nextZ][nextY][nextX].pathCount = count + abs(origW-w)
			}
		} else {
			rooms[nextW][nextZ][nextY][nextX].pathCount = count + 1
		}
		count++

		if goalReached {
			break
		}

		x, y, z, w = nextX, nextY, nextZ, nextW
	}

	if !oneWayExists {
		return nil
	}
	return rooms
}

func (f *FieldData) areEnoughRoomsVisited(rooms [][][][]room) bool {
	var visited int
	threshold := (f.width * f.height * f.depth0 * f.depth1) * 8 / 10
	for w := range f.depth1 {
		for z := range f.depth0 {
			for y := range f.height {
				for x := range f.width {
					if rooms[w][z][y][x].pathCount > 0 {
						visited++
						if visited >= threshold {
							return true
						}
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
	roomYGridCount = 3
)

func (f *FieldData) roomXGridCount() int {
	switch f.depth1 {
	case 1:
		return 6
	case 2:
		return 8
	default:
		panic("not reached")
	}
}

func (f *FieldData) setTiles(rooms [][][][]room) {
	roomXGridCount := f.roomXGridCount()

	width := f.width*roomXGridCount + 1
	height := f.height*roomYGridCount + 2

	f.tiles = make([][]tile, height)
	for y := range f.tiles {
		f.tiles[y] = make([]tile, width)
		for x := range f.tiles[y] {
			f.tiles[y][x].walls = make([]bool, f.depth1)
			f.tiles[y][x].ladders = make([]bool, f.depth1)
			f.tiles[y][x].switches = make([]bool, f.depth1)
			f.tiles[y][x].wallColors = make([]int, f.depth1)
			f.tiles[y][x].ladderColors = make([]int, f.depth1)
		}
	}

	// Set the outside walls.
	for x := range f.tiles[0] {
		for i := range f.tiles[0][x].walls {
			f.tiles[0][x].walls[i] = true
		}
	}
	for y := range f.tiles {
		for i := range f.tiles[y][0].walls {
			f.tiles[y][0].walls[i] = true
		}
	}
	for i := range f.tiles[height-1][width-1].walls {
		f.tiles[height-1][width-1].walls[i] = true
	}

	// Set the goal.
	f.tiles[height-1][width-roomXGridCount-1].goal = true

	for y := range f.height {
		for x := range f.width {
			f.setTilesForRoom(rooms, x, y)
		}
	}
}

func (f *FieldData) setTilesForRoom(rooms [][][][]room, roomX, roomY int) {
	const (
		edgeOffsetX = 1
		edgeOffsetY = 1
	)
	roomXGridCount := f.roomXGridCount()

	// Add walls.
	colors, oks := f.wallColors(rooms, roomX, roomY)
	for j := range roomYGridCount - 1 {
		x := roomX*roomXGridCount + roomXGridCount - 1 + edgeOffsetX
		x -= f.depth1 - 1
		y := roomY*roomYGridCount + j + edgeOffsetY
		allNonColorWall := true
		for w := range f.depth1 {
			if !oks[w] {
				allNonColorWall = false
				break
			}
			if colors[w] != 0 {
				allNonColorWall = false
				break
			}
		}
		if allNonColorWall {
			for w := range f.depth1 {
				f.tiles[y][x+(f.depth1-1)].walls[w] = true
			}
		} else {
			for w := range f.depth1 {
				if !oks[w] {
					continue
				}
				f.tiles[y][x+w].walls[w] = true
				f.tiles[y][x+w].wallColors[w] = colors[w]
			}
		}
	}

	// Add a ceiling.
	for i := range roomXGridCount {
		x := roomX*roomXGridCount + i + edgeOffsetX
		y := (roomY+1)*roomYGridCount - 1 + edgeOffsetY
		for w := range f.depth1 {
			f.tiles[y][x].walls[w] = true
		}
	}

	// Add ladders.
	passageYs := make([]passage, f.depth1)
	for i := range passageYs {
		passageYs[i] = passageWall
	}
	for w := range f.depth1 {
		for z := range f.depth0 {
			room := rooms[w][z][roomY][roomX]
			if room.passageY == passageWall {
				continue
			}
			if passageYs[w] == passageWall {
				passageYs[w] = room.passageY
				continue
			}
			if passageYs[w] != room.passageY {
				panic("not reached")
			}
			passageYs[w] = room.passageY
		}
	}
	colors, oks = f.ladderColors(rooms, roomX, roomY)
	for j := range roomYGridCount {
		y := roomY*roomYGridCount + j + edgeOffsetY
		for w := range f.depth1 {
			if !oks[w] {
				continue
			}
			x := roomX*roomXGridCount + 1 + ((roomY + w) % 2) + edgeOffsetX
			f.tiles[y][x].ladders[w] = true
			f.tiles[y][x].ladderColors[w] = colors[w]
			if passageYs[w] == passageOneWayForward {
				f.tiles[y][x].upward = true
			}
			if passageYs[w] == passageOneWayBackward {
				f.tiles[y][x].downward = true
			}
		}
	}

	// Add switches.
	for w := range f.depth1 {
		if rooms[w][0][roomY][roomX].passageZ != passageWall {
			x := roomX*roomXGridCount + 3 + w + edgeOffsetX
			y := roomY*roomYGridCount + edgeOffsetY
			f.tiles[y][x].switches[w] = true
		}
	}

	// Add doors.
	if color, ok := f.doorColor(rooms, roomX, roomY); ok {
		x := roomX*roomXGridCount + 5 + edgeOffsetX
		y := roomY*roomYGridCount + edgeOffsetY
		f.tiles[y][x].door = true
		f.tiles[y][x].doorColor = color
		f.tiles[y+1][x].doorUpper = true
		f.tiles[y+1][x].doorColor = color
	}
}

func (f *FieldData) wallColors(rooms [][][][]room, roomX, roomY int) (colors []int, oks []bool) {
	colors = make([]int, f.depth1)
	oks = make([]bool, f.depth1)
	for w := range f.depth1 {
		x0 := rooms[w][0][roomY][roomX].passageX == passageWall
		x1 := rooms[w][1][roomY][roomX].passageX == passageWall
		if !x0 && x1 {
			colors[w] = 1 // TODO: Add (w * f.depth0)?
		}
		if x0 && !x1 {
			colors[w] = 2
		}
		if x0 || x1 {
			oks[w] = true
		}
	}
	return
}

func (f *FieldData) ladderColors(rooms [][][][]room, roomX, roomY int) (colors []int, oks []bool) {
	colors = make([]int, f.depth1)
	oks = make([]bool, f.depth1)
	for w := range f.depth1 {
		y0 := rooms[w][0][roomY][roomX].passageY != passageWall
		y1 := rooms[w][1][roomY][roomX].passageY != passageWall
		if y0 && !y1 {
			colors[w] = 1
		}
		if !y0 && y1 {
			colors[w] = 2
		}
		if y0 || y1 {
			oks[w] = true
		}
	}
	return
}

func (f *FieldData) doorColor(rooms [][][][]room, roomX, roomY int) (color int, ok bool) {
	w0 := rooms[0][0][roomY][roomX].passageW != passageWall
	w1 := rooms[0][1][roomY][roomX].passageW != passageWall
	if w0 && !w1 {
		color = 1
	}
	if !w0 && w1 {
		color = 2
	}
	if w0 || w1 {
		ok = true
	}
	return
}

func (f *FieldData) hasSwitch(x, y int, currentDepth1 int) bool {
	return f.tiles[y][x].switches[currentDepth1]
}

func (f *FieldData) hasDoor(x, y int, currentDepth0 int) bool {
	if !f.tiles[y][x].door {
		return false
	}
	return f.tiles[y][x].doorColor == 0 || f.tiles[y][x].doorColor-1 == currentDepth0
}

func (f *FieldData) passable(nextX, nextY int, prevY int, currentDepth0 int, currentDepth1 int) bool {
	if nextY < 0 || len(f.tiles) <= nextY || nextX < 0 || len(f.tiles[nextY]) <= nextX {
		return false
	}
	if !f.canBeInTile(nextX, nextY, currentDepth0, currentDepth1) {
		return false
	}
	if !f.canStandOnTile(nextX, nextY-1, currentDepth0, currentDepth1) {
		return false
	}
	if nextY > prevY && !f.canGoUp(nextX, nextY, currentDepth0, currentDepth1) {
		return false
	}
	if nextY < prevY && !f.canGoDown(nextX, nextY, currentDepth0, currentDepth1) {
		return false
	}
	return true
}

func (f *FieldData) canBeInTile(x, y int, currentDepth0 int, currentDepth1 int) bool {
	if y < 0 || len(f.tiles) <= y || x < 0 || len(f.tiles[y]) <= x {
		return false
	}
	t := f.tiles[y][x]
	// A ladder is passable, even though this is a wall.
	if t.ladders[currentDepth1] {
		if t.ladderColors[currentDepth1] == 0 || t.ladderColors[currentDepth1]-1 == currentDepth0 {
			return true
		}
	}
	// A wall is not passable.
	if t.walls[currentDepth1] {
		if t.wallColors[currentDepth1] == 0 || (t.wallColors[currentDepth1]-1 != currentDepth0) {
			return false
		}
	}
	return true
}

func (f *FieldData) canStandOnTile(x, y int, currentDepth0 int, currentDepth1 int) bool {
	if y < 0 || len(f.tiles) <= y || x < 0 || len(f.tiles[y]) <= x {
		return false
	}
	t := f.tiles[y][x]
	// A player can stand on a ladder.
	if t.ladders[currentDepth1] {
		if t.ladderColors[currentDepth1] == 0 || t.ladderColors[currentDepth1]-1 == currentDepth0 {
			return true
		}
	}
	// A player can stand on a wall.
	if t.walls[currentDepth1] {
		if t.wallColors[currentDepth1] == 0 || (t.wallColors[currentDepth1]-1 != currentDepth0) {
			return true
		}
	}
	return false
}

func (f *FieldData) canGoUp(x, y int, currentDepth0 int, currentDepth1 int) bool {
	if y < 0 || len(f.tiles) <= y || x < 0 || len(f.tiles[y]) <= x {
		return false
	}
	t := f.tiles[y][x]
	if !t.ladders[currentDepth1] {
		return true
	}
	if t.ladderColors[currentDepth1] > 0 && t.ladderColors[currentDepth1]-1 != currentDepth0 {
		return true
	}
	return !t.downward
}

func (f *FieldData) canGoDown(x, y int, currentDepth0 int, currentDepth1 int) bool {
	if y < 0 || len(f.tiles) <= y || x < 0 || len(f.tiles[y]) <= x {
		return false
	}
	t := f.tiles[y][x]
	if !t.ladders[currentDepth1] {
		return true
	}
	if t.ladderColors[currentDepth1] > 0 && t.ladderColors[currentDepth1]-1 != currentDepth0 {
		return true
	}
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

func (f *FieldData) Draw(screen *ebiten.Image, offsetX, offsetY int, currentDepth0, currentDepth1 int) {
	for y := range f.tiles {
		for x := range f.tiles[y] {
			dx := x*GridSize + offsetX
			dy := -(y+1)*GridSize + offsetY

			if dx < -GridSize || dx >= screen.Bounds().Dx() || dy < -GridSize || dy >= screen.Bounds().Dy() {
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(dx), float64(dy))

			const transparent = 0.25
			t := f.tiles[y][x]
			for w := range f.depth1 {
				if t.walls[w] {
					img := f.wallImage
					if t.wallColors[w] != 0 {
						c := t.wallColors[w] - 1
						if currentDepth0 == c {
							img = f.colorPassableWallImages[f.colorPalette[c]]
						} else {
							img = f.colorUnpassableWallImages[f.colorPalette[c]]
						}
					}
					op.ColorScale = ebiten.ColorScale{}
					if currentDepth1 != w {
						op.ColorScale.ScaleAlpha(transparent)
					}
					screen.DrawImage(img, op)
				}
			}
			for w := range f.depth1 {
				if t.ladders[w] {
					c := -1
					idx := -1
					if t.ladderColors[w] != 0 {
						c = t.ladderColors[w] - 1
						idx = f.colorPalette[c]
					}
					var img *ebiten.Image
					switch {
					case !t.upward && !t.downward:
						if c < 0 {
							img = f.ladderImage
						} else if currentDepth0 == c {
							img = f.colorPassableLadderImages[idx]
						} else {
							img = f.colorUnpassableLadderImages[idx]
						}
					case t.upward:
						if c < 0 {
							img = f.upwardImage
						} else if currentDepth0 == c {
							img = f.colorUpwardImage[idx]
						} else {
							img = f.colorUpwardDisabledImage[idx]
						}
					case t.downward:
						if c < 0 {
							img = f.downwardImage
						} else if currentDepth0 == c {
							img = f.colorDownwardImage[idx]
						} else {
							img = f.colorDownwardDisabledImage[idx]
						}
					}
					op.ColorScale = ebiten.ColorScale{}
					if currentDepth1 != w {
						op.ColorScale.ScaleAlpha(transparent)
					}
					screen.DrawImage(img, op)
				}
			}
			for w := range f.depth1 {
				if t.switches[w] {
					switchImage := f.switchImages[f.colorPalette[currentDepth0]]
					op.ColorScale = ebiten.ColorScale{}
					if currentDepth1 != w {
						op.ColorScale.ScaleAlpha(transparent)
					}
					screen.DrawImage(switchImage, op)
				}
			}
			if t.doorUpper {
				img := f.doorImage
				if t.doorColor != 0 {
					c := t.doorColor - 1
					if c == currentDepth0 {
						img = f.colorDoorImages[f.colorPalette[c]]
					} else {
						img = f.colorDoorDisabledImages[f.colorPalette[c]]
					}
				}
				op.ColorScale = ebiten.ColorScale{}
				screen.DrawImage(img, op)
			}
			if t.goal {
				screen.DrawImage(f.goalImage, op)
			}
		}
	}
}

var doorImage = ebiten.NewImage(16, 16)

func init() {
	doorImage.Fill(color.White)
}
