package gmgmap

import (
	"math/rand"
)

type bspArea struct {
	bspRoom
	isStreet         bool
	isConnected      bool
	isOnCriticalPath bool
}

func (s bspArea) dAlong() vec2 {
	if s.horizontal {
		return vec2{1, 0}
	}
	return vec2{0, 1}
}

func (s bspArea) dAcross() vec2 {
	if s.horizontal {
		return vec2{0, 1}
	}
	return vec2{1, 0}
}

type adjacencyMatrix [][]bool

func newAdjacencyMatrix(x int) adjacencyMatrix {
	a := make([][]bool, x)
	for i := 0; i < x; i++ {
		a[i] = make([]bool, x)
	}
	return a
}

func (a adjacencyMatrix) Connect(i, j int) {
	a[i][j] = true
	a[j][i] = true
}

func (a adjacencyMatrix) IsConnected(i, j int) bool {
	return a[i][j]
}

// NewBSPInterior - Create new BSP interior map
// Implementation of https://gamedev.stackexchange.com/questions/47917/procedural-house-with-rooms-generator/48216#48216
func NewBSPInterior(rr *rand.Rand, exportFunc func(*Map), width, height, splits, minRoomSize, corridorWidth int) *Map {
	corridorLevelDiffBlock := 1
	m := NewMap(width, height)
	var areas []bspArea

	// Split the map for a number of iterations, choosing alternating axis and random location
	hcount := rr.Intn(2)
	areas = append(areas, bspArea{bspRoomRoot(width, height), false, false, false})
	for i := 0; i < len(areas); i++ {
		if areas[i].level == splits {
			break
		}
		var r1, r2 bspRoom
		var err error = nil
		// Alternate splitting direction per level
		horizontal := ((hcount + areas[i].level) % 2) == 1
		if horizontal {
			r1, r2, err = areas[i].SplitHorizontal(rr, i, minRoomSize+corridorWidth/2)
		} else {
			r1, r2, err = areas[i].SplitVertical(rr, i, minRoomSize+corridorWidth/2)
		}
		if err == nil {
			// Resize rooms to allow space for street
			for j := 0; j < corridorWidth; j++ {
				if horizontal {
					if j%2 == 0 {
						r1.r.w--
					} else {
						r2.r.x++
						r2.r.w--
					}
				} else {
					if j%2 == 0 {
						r1.r.h--
					} else {
						r2.r.y++
						r2.r.h--
					}
				}
			}
			// Replace current area with a street
			areas[i].isStreet = true
			if horizontal {
				areas[i].r = rect{r1.r.x + r1.r.w, r1.r.y, corridorWidth, r1.r.h}
			} else {
				areas[i].r = rect{r1.r.x, r1.r.y + r1.r.h, r1.r.w, corridorWidth}
			}
			areas[i].horizontal = !horizontal
			areas[i].child1 = len(areas)
			areas = append(areas, bspArea{r1, false, false, false})
			areas[i].child2 = len(areas)
			areas = append(areas, bspArea{r2, false, false, false})
		}
	}
	// Try to split leaf rooms into more rooms, by longest axis
	for i := 0; i < len(areas); i++ {
		if areas[i].isStreet {
			continue
		}

		var r1, r2 bspRoom
		var err error = nil
		if areas[i].r.w > areas[i].r.h {
			r1, r2, err = areas[i].SplitHorizontal(rr, i, minRoomSize)
		} else {
			r1, r2, err = areas[i].SplitVertical(rr, i, minRoomSize)
		}
		if err == nil {
			// Resize rooms so they share a splitting wall
			if r1.horizontal {
				r1.r.w++
			} else {
				r1.r.h++
			}
			areas[i].child1 = len(areas)
			areas = append(areas, bspArea{r1, false, false, false})
			areas[i].child2 = len(areas)
			areas = append(areas, bspArea{r2, false, false, false})
		}
	}

	g := m.Layer("Ground")
	s := m.Layer("Structures")

	// Fill rooms
	for i := range areas {
		// Note: we should skip non-leaves, but for the sake of looking good
		// when iteratively generating, make those rooms anyway
		// The leaf nodes should be later in the areas collection
		//if !areas[i].IsLeaf() { continue }
		r := areas[i].r
		g.rectangleFilled(rect{r.x + 1, r.y + 1, r.w - 2, r.h - 2}, room)
		s.rectangleUnfilled(r, wall2)
		exportFunc(m)
	}

	// Set up adjacency matrix
	adjacency := newAdjacencyMatrix(len(areas))
	for i := range areas {
		if !areas[i].isStreet || areas[i].parent < 0 {
			continue
		}
		adjacency.Connect(i, areas[i].parent)
	}

	// Add door openings between rooms and corridors
	for i := range areas {
		// Skip non-leaves
		if !areas[i].IsLeaf() || areas[i].isStreet {
			continue
		}

		r := areas[i].r
		// Add openings to corridors (streets)
		for j := range areas {
			if !areas[j].isStreet {
				continue
			}

			streetR := areas[j].r

			// Check if room is adjacent to street and create opening
			if (r.x+r.w == streetR.x) || (streetR.x+streetR.w == r.x) {
				// Vertical opening
				overlapY := imax(r.y, streetR.y)
				overlapEndY := imin(r.y+r.h, streetR.y+streetR.h)
				if overlapEndY > overlapY {
					openY := overlapY + (overlapEndY-overlapY)/2
					if r.x+r.w == streetR.x {
						// Opening on right side of room
						s.setTile(r.x+r.w-1, openY, nothing)
						g.setTile(r.x+r.w-1, openY, room2)
					} else {
						// Opening on left side of room
						s.setTile(r.x, openY, nothing)
						g.setTile(r.x, openY, room2)
					}
					areas[i].isConnected = true
					adjacency.Connect(i, j)
				}
			} else if (r.y+r.h == streetR.y) || (streetR.y+streetR.h == r.y) {
				// Horizontal opening
				overlapX := imax(r.x, streetR.x)
				overlapEndX := imin(r.x+r.w, streetR.x+streetR.w)
				if overlapEndX > overlapX {
					openX := overlapX + (overlapEndX-overlapX)/2
					if r.y+r.h == streetR.y {
						// Opening on bottom side of room
						s.setTile(openX, r.y+r.h-1, nothing)
						g.setTile(openX, r.y+r.h-1, room2)
					} else {
						// Opening on top side of room
						s.setTile(openX, r.y, nothing)
						g.setTile(openX, r.y, room2)
					}
					areas[i].isConnected = true
					adjacency.Connect(i, j)
				}
			}
		}
		exportFunc(m)
	}

	// Find deepest leaf going down both branches; place stairs
	// This represents longest/critical path
	deepestRoom1 := findDeepestRoomFrom(areas, areas[0].child1)
	placeInsideRoom(s, areas[deepestRoom1].r, stairsUp)
	exportFunc(m)
	deepestRoom2 := findDeepestRoomFrom(areas, areas[0].child2)
	placeInsideRoom(s, areas[deepestRoom2].r, stairsDown)
	exportFunc(m)
	markParentStreets := func(area *bspArea) {
		street := area
		for {
			street.isOnCriticalPath = true
			street = &areas[street.parent]
			if street == &areas[0] {
				break
			}
		}
		street.isOnCriticalPath = true
	}
	markParentStreets(&areas[deepestRoom1])
	markParentStreets(&areas[deepestRoom2])

	// Fill streets
	for i := range areas {
		if !areas[i].isStreet {
			continue
		}
		g.rectangleFilled(areas[i].r, room2)
		// Remove the walls we added from non-leaf rooms before
		s.rectangleFilled(areas[i].r, nothing)
		// Check ends of street - cap or place door
		end1 := vec2{areas[i].r.x, areas[i].r.y}
		end2 := vec2{areas[i].r.x + areas[i].r.w - 1, areas[i].r.y + areas[i].r.h - 1}
		capStreet(g, s, areas, areas[i], end1, areas[i].dAcross(), areas[i].dAlong(), corridorWidth, corridorLevelDiffBlock)
		capStreet(g, s, areas, areas[i], end2, vec2{-areas[i].dAcross().x, -areas[i].dAcross().y}, vec2{-areas[i].dAlong().x, -areas[i].dAlong().y}, corridorWidth, corridorLevelDiffBlock)
		exportFunc(m)
	}

	// Use adjacency matrix to determine distance of all leaf nodes from critical path
	dCriticalPath := make([]int, len(areas))
	for i := range areas {
		if areas[i].isOnCriticalPath {
			dCriticalPath[i] = 1
		}
	}
	for {
		newConnections := 0
		for i := range areas {
			if dCriticalPath[i] > 0 {
				continue
			}
			if !areas[i].isStreet && !areas[i].IsLeaf() {
				continue
			}
			for j := range areas {
				if adjacency.IsConnected(i, j) && dCriticalPath[j] > 0 {
					dCriticalPath[i] = dCriticalPath[j] + 1
					newConnections++
					break
				}
			}
		}
		if newConnections == 0 {
			break
		}
	}

	// Key generation removed - no longer spawning keys for locked doors

	// Characters will be placed separately by the display layer, not during map generation

	return m
}

func capStreet(g, s *Layer, streets []bspArea, st bspArea, end, dAcross, dAlong vec2, corridorWidth, corridorLevelDiffBlock int) {
	// Check ends of street - if outside map, or next to much older street, block off with wall
	outside := vec2{end.x - dAlong.x, end.y - dAlong.y}
	capTile := floor
	capStructure := nothing
	if !g.isIn(outside.x, outside.y) {
		capTile = nothing
		capStructure = wall2
	} else {
		for i := range streets {
			if streets[i].r.isIn(outside.x, outside.y) {
				if st.level-streets[i].level > corridorLevelDiffBlock {
					capTile = nothing
					capStructure = wall2
				} else {
					capTile = room2
					capStructure = nothing
				}
				break
			}
		}
	}
	for i := 0; i < corridorWidth; i++ {
		g.setTile(end.x+dAcross.x*i, end.y+dAcross.y*i, capTile)
		s.setTile(end.x+dAcross.x*i, end.y+dAcross.y*i, capStructure)
	}
}

func findDeepestRoomFrom(areas []bspArea, child int) int {
	var pathStack []int
	pathStack = append(pathStack, child)
	deepestChild := -1
	maxDepth := 0
	for len(pathStack) > 0 {
		i := pathStack[len(pathStack)-1]
		r := areas[i]
		pathStack = pathStack[:len(pathStack)-1]
		if r.IsLeaf() {
			if r.level > maxDepth {
				maxDepth = r.level
				deepestChild = i
			}
		}
		if r.child1 >= 0 {
			pathStack = append(pathStack, r.child1)
		}
		if r.child2 >= 0 {
			pathStack = append(pathStack, r.child2)
		}
	}
	return deepestChild
}

func placeInsideRoom(s *Layer, r rect, t rune) {
	s.setTile(r.x+r.w/2, r.y+r.h/2, t)
}
