package gmgmap

import "math/rand"

func NewBSP(rr *rand.Rand, width, height, iterations, minRoomSize, connectionIterations int) *Map {
	m := NewMap(width, height)

	var areas []bspRoom
	areas = append(areas, bspRoomRoot(width, height))
	for i := 0; i < len(areas); i++ {
		if areas[i].level == iterations {
			break
		}
		if r1, r2, err := areas[i].Split(rr, i, minRoomSize, 0); err == nil {
			areas[i].child1 = len(areas)
			areas = append(areas, r1)
			areas[i].child2 = len(areas)
			areas = append(areas, r2)
		}
	}

	g := m.Layer("Ground")
	s := m.Layer("Structures")
	for i := range areas {
		if !areas[i].IsLeaf() {
			continue
		}
		var r rect
		if areas[i].r.w == minRoomSize {
			r.w = minRoomSize
			r.x = areas[i].r.x
		} else {
			r.w = rr.Intn(areas[i].r.w-minRoomSize) + minRoomSize
			r.x = rr.Intn(areas[i].r.w-r.w) + areas[i].r.x
			xmid := areas[i].r.x + areas[i].r.w/2
			r.x = iclamp(r.x, xmid-(r.w-2), xmid-1)
		}
		if areas[i].r.h == minRoomSize {
			r.h = minRoomSize
			r.y = areas[i].r.y
		} else {
			r.h = rr.Intn(areas[i].r.h-minRoomSize) + minRoomSize
			r.y = rr.Intn(areas[i].r.h-r.h) + areas[i].r.y
			ymid := areas[i].r.y + areas[i].r.h/2
			r.y = iclamp(r.y, ymid-(r.h-2), ymid-1)
		}
		g.rectangleFilled(rect{r.x + 1, r.y + 1, r.w - 2, r.h - 2}, room)
		s.rectangleUnfilled(r, wall2)
	}

	for i := len(areas) - 1; i >= 0; i-- {
		a := areas[i]
		if a.parent < 0 {
			continue
		}
		siblingIndex := areas[a.parent].child2
		if siblingIndex == i || siblingIndex < 0 {
			continue
		}
		sibling := areas[siblingIndex]
		if a.r.x < sibling.r.x {
			aymin, aymax := getYRange(*g, a.r)
			symin, symax := getYRange(*g, sibling.r)
			y := irand(rr, imax(aymin, symin), imin(aymax, symax))
			addStraightCorridor(g, s, a.r.x+a.r.w, y, 1, 0, room2, wall2)
		} else {
			axmin, axmax := getXRange(*g, a.r)
			sxmin, sxmax := getXRange(*g, sibling.r)
			x := irand(rr, imax(axmin, sxmin), imin(axmax, sxmax))
			addStraightCorridor(g, s, x, a.r.y+a.r.h, 0, 1, room2, wall2)
		}
	}

	for n := 0; n < connectionIterations; n++ {
		i := rr.Intn(len(areas))
		a := areas[i]
		if !a.IsLeaf() {
			continue
		}
		dx := rr.Intn(1)*2 - 1
		dy := rr.Intn(1)*2 - 1
		if rr.Intn(1) > 0 {
			dx = 0
		} else {
			dy = 0
		}
		c1 := areas[a.parent].child1
		c2 := areas[a.parent].child2
		if c1 >= 0 && c2 >= 0 {
			sibling := areas[c1+c2-i]
			if (a.r.x < sibling.r.x && dx == 1) ||
				(a.r.x > sibling.r.x && dx == -1) ||
				(a.r.y < sibling.r.y && dy == 1) ||
				(a.r.y > sibling.r.y && dy == -1) {
				continue
			}
		}
		x := a.r.x + a.r.w/2
		if dx > 0 {
			x = a.r.x + a.r.w
		} else if dx < 0 {
			x = a.r.x
		}
		y := a.r.y + a.r.h/2
		if dy > 0 {
			x = a.r.y + a.r.h
		} else if dy < 0 {
			x = a.r.y
		}
		if canDrawInDirection(*g, x, y, dx, dy) {
			addStraightCorridor(g, s, x, y, dx, dy, room2, wall2)
		}
	}

	return m
}

func getXRange(l Layer, r rect) (int, int) {
	var minx, maxx int
	for minx = r.x; minx < r.x+r.w; minx++ {
		if !l.isClear(minx, r.y, 1, r.h) {
			break
		}
	}
	for maxx = r.x + r.w - 1; maxx > minx; maxx-- {
		if !l.isClear(maxx, r.y, 1, r.h) {
			break
		}
	}
	return minx, maxx
}
func getYRange(l Layer, r rect) (int, int) {
	var miny, maxy int
	for miny = r.y; miny < r.y+r.h; miny++ {
		if !l.isClear(r.x, miny, r.w, 1) {
			break
		}
	}
	for maxy = r.y + r.h - 1; maxy > miny; maxy-- {
		if !l.isClear(r.x, maxy, r.w, 1) {
			break
		}
	}
	return miny, maxy
}

func addStraightCorridor(g, s *Layer, startX, startY, dx, dy int, tile, wall rune) {
	drawInDirection(g, s, startX, startY, dx, dy, tile, wall)
	drawInDirection(g, s, startX, startY, -dx, -dy, tile, wall)
}

func drawInDirection(g, s *Layer, startX, startY, dx, dy int, tile, wall rune) {
	drawEnd := false
	for x := startX; !drawEnd; x += dx {
		for y := startY; !drawEnd; y += dy {
			g.setTile(x, y, tile)
			s.setTile(x, y, nothing)
			if dx == 0 {
				if s.isIn(x+1, y) && g.getTile(x+1, y) == nothing {
					s.setTile(x+1, y, wall)
				}
				if s.isIn(x-1, y) && g.getTile(x-1, y) == nothing {
					s.setTile(x-1, y, wall)
				}
			} else {
				if s.isIn(x, y+1) && g.getTile(x, y+1) == nothing {
					s.setTile(x, y+1, wall)
				}
				if s.isIn(x, y-1) && g.getTile(x, y-1) == nothing {
					s.setTile(x, y-1, wall)
				}
			}
			if hasNeighbouringTile(*g, x, y, dx, dy) {
				drawEnd = true
			}
			if dy == 0 {
				break
			}
		}
		if dx == 0 {
			break
		}
	}
}

func canDrawInDirection(g Layer, startX, startY, dx, dy int) bool {
	drawEnd := false
	for x := startX; !drawEnd; x += dx {
		for y := startY; !drawEnd; y += dy {
			if !g.isIn(x, y) {
				return false
			}
			if hasNeighbouringTile(g, x, y, dx, dy) {
				drawEnd = true
			}
			if dy == 0 {
				break
			}
		}
		if dx == 0 {
			break
		}
	}
	return true
}

func hasNeighbouringTile(l Layer, x, y, dx, dy int) bool {
	if dx != 0 {
		if hasTile(l, x, y+1) {
			return true
		}
		if hasTile(l, x, y-1) {
			return true
		}
		if hasTile(l, x+dx, y) {
			return true
		}
	} else {
		if hasTile(l, x+1, y) {
			return true
		}
		if hasTile(l, x-1, y) {
			return true
		}
		if hasTile(l, x, y+dy) {
			return true
		}
	}
	return false
}

func hasTile(l Layer, x, y int) bool {
	return l.isIn(x, y) && l.getTile(x, y) != nothing
}
