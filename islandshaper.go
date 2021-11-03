package main

import (
	"sort"

	"gogen.quis.cx/bytelib"
)

func init() {
	Register(IslandShaper, "islandShaper", "shape", "shaper")
}

func IslandShaper(s State) State {
	s = s.Clone()
	rm := reachabilityMap(s)
	poolToSolvers := map[Pos][]byte{}
	for _, pool := range potentialPools(s) {
		poolToSolvers[pool] = poolSolvers(pool, rm)
	}
	reachableByIsland := map[byte][]Pos{}
	for p, rs := range rm {
		for _, l := range rs {
			reachableByIsland[l] = append(reachableByIsland[l], p)
		}
	}
	for l, rem := range s.UndiscoveredIslands() {
		if rem == 0 {
			continue
		}
		var myPools []Pos
		for p, solvers := range poolToSolvers {
			if len(solvers) == 1 && solvers[0] == l {
				myPools = append(myPools, p)
			}
		}
		for _, p := range shapeMyIsland(s, islandParts(s, l), reachableByIsland[l], rem, myPools) {
			s.Set(p, Cell(l))
		}
	}
	return s
}

func islandParts(s State, island byte) []Pos {
	var ret []Pos
	for y, r := range s.board {
		for x, c := range r {
			if c.KnownLand() && c.Char() == island {
				ret = append(ret, Pos{x, y})
			}
		}
	}
	return ret
}

func poolSolvers(pool Pos, rm map[Pos][]byte) []byte {
	var ret []byte
	for _, p := range []Pos{{pool.x, pool.y}, {pool.x, pool.y + 1}, {pool.x + 1, pool.y}, {pool.x + 1, pool.y + 1}} {
		ret = append(ret, rm[p]...)
	}
	return bytelib.Unique(ret)
}

func potentialPools(s State) []Pos {
	var ret []Pos
	for y, r := range s.board {
		if y == 0 {
			continue
		}
	cell:
		for x := range r {
			if x == 0 {
				continue
			}
			for _, p := range []Pos{{x - 1, y - 1}, {x - 1, y}, {x, y - 1}, {x, y}} {
				if s.Get(p).Land() {
					continue cell
				}
			}
			ret = append(ret, Pos{x - 1, y - 1})
		}
	}
	return ret
}

func shapeMyIsland(s State, parts []Pos, reachable []Pos, rem int, pools []Pos) []Pos {
	totalOptions := posSliceToMap(reachable)
	firstOptions := optionsAround(s, parts, totalOptions)
	firstProposal := proposeIsland(s, posSliceToMap(parts), firstOptions, totalOptions, rem, pools)
	unavoidable := map[Pos]void{}
	for _, p := range firstProposal {
		unavoidable[p] = void{}
	}
	for _, p := range firstProposal {
		if _, f := unavoidable[p]; !f {
			// We already found a shape without this one.
			continue
		}
		delete(totalOptions, p)
		firstOptions = optionsAround(s, parts, totalOptions)
		proposal := proposeIsland(s, posSliceToMap(parts), firstOptions, totalOptions, rem, pools)
		totalOptions[p] = void{}
		if len(proposal) == 0 {
			// This one is unavoidable
			continue
		}
		for p := range unavoidable {
			if !inPath(proposal, p) {
				delete(unavoidable, p)
			}
		}
	}
	return posMapToSlice(unavoidable)
}

func proposeIsland(s State, parts, options, totalOptions map[Pos]void, rem int, pools []Pos) []Pos {
	for _, p := range posMapToSlice(options) {
		if _, found := parts[p]; found {
			continue
		}
		parts[p] = void{}
		if rem == 1 {
			ok := true
		poolLoop:
			for _, pool := range pools {
				for _, pp := range []Pos{{pool.x, pool.y}, {pool.x, pool.y + 1}, {pool.x + 1, pool.y}, {pool.x + 1, pool.y + 1}} {
					if _, f := parts[pp]; f {
						continue poolLoop
					}
				}
				ok = false
				break
			}
			if ok {
				ps := posMapToSlice(parts)
				if isConnected(ps) {
					return ps
				}
			}
			delete(parts, p)
			continue
		}
		delete(options, p)
		var undo []Pos
		for _, np := range s.Around(p) {
			if _, found := options[np]; !found {
				if _, valid := totalOptions[np]; !valid {
					continue
				}
				options[np] = void{}
				undo = append(undo, np)
			}
		}
		proposal := proposeIsland(s, parts, options, totalOptions, rem-1, pools)
		// proposal := proposeIsland(s, parts, optionsAround(s, posMapToSlice(parts), totalOptions), totalOptions, rem-1, pools)
		if len(proposal) > 0 {
			return proposal
		}
		delete(parts, p)
		options[p] = void{}
		for _, op := range undo {
			delete(options, op)
		}
	}
	return nil
}

func isConnected(ps []Pos) bool {
	unconnected := append([]Pos(nil), ps[1:]...)
	reachable := make([]Pos, 1, len(ps))
	reachable[0] = ps[0]
outer:
	for progressing := true; progressing; {
		progressing = false
		for i, up := range unconnected {
			for _, rp := range reachable {
				if up.Distance(rp) == 1 {
					progressing = true
					reachable = append(reachable, up)
					unconnected[i] = unconnected[len(unconnected)-1]
					unconnected = unconnected[:len(unconnected)-1]
					continue outer
				}
			}
		}
	}
	return len(unconnected) == 0
}

func posMapToSlice(m map[Pos]void) []Pos {
	ret := make([]Pos, 0, len(m))
	for p := range m {
		ret = append(ret, p)
	}
	return ret
}

func posSliceToMap(ps []Pos) map[Pos]void {
	ret := map[Pos]void{}
	for _, p := range ps {
		ret[p] = void{}
	}
	return ret
}

func optionsAround(s State, parts []Pos, totalOptions map[Pos]void) map[Pos]void {
	options := map[Pos]void{}
	for _, p := range parts {
		for _, np := range s.Around(p) {
			if _, valid := totalOptions[np]; valid {
				options[np] = void{}
			}
		}
	}
	for _, p := range parts {
		delete(options, p)
	}
	return options
}

func sortedCopy(ps []Pos) []Pos {
	c := make([]Pos, len(ps))
	copy(c, ps)
	sort.Slice(c, func(i, j int) bool {
		if c[i].y != c[j].y {
			return c[i].y < c[j].y
		}
		return c[i].x < c[j].x
	})
	return c
}
