package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"golang.org/x/sync/semaphore"
)

func init() {
	Register(MendIslands, "mendIslands", "mend")
}

func MendIslands(s State) State {
	s = s.Clone()
	remaining := s.UndiscoveredIslands()
	islandTiles := map[byte][]Pos{}
	for y, r := range s.board {
		for x, c := range r {
			if !c.Land() || c == SomeLand {
				continue
			}
			islandTiles[c.Char()] = append(islandTiles[c.Char()], Pos{x, y})
		}
	}
	var mtx sync.Mutex
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	newIslands := map[Pos]byte{}
	for l, r := range remaining {
		if r == 0 {
			continue
		}
		bodies := GroupBodies(s, islandTiles[l])
		for i, b1 := range bodies {
			for _, b2 := range bodies[i+1:] {
				b1 := b1
				b2 := b2
				l := l
				r := r
				wg.Add(1)
				sem.Acquire(context.Background(), 1)
				go func() {
					nw := findIslandPaths(s, l, b1.Any(), b2.Any(), r+1)
					sem.Release(1)
					mtx.Lock()
					for p := range nw {
						newIslands[p] = l
					}
					mtx.Unlock()
					wg.Done()
				}()
			}
		}
	}

	wg.Wait()
	for p, c := range newIslands {
		s.Set(p, Cell(c))
	}
	return s
}

func findIslandPaths(s State, island byte, dst, src Pos, rem int) map[Pos]void {
	fmt.Printf("findIslandPaths(%v -> %v for %c in %d)\n", src, dst, island, rem)
	firstPath := islandDFS(s, island, src, dst, rem, map[Pos]int{})
	// fmt.Printf("First path (%v -> %v): %v\n", src, dst, firstPath)
	unavoidable := map[Pos]void{}
	for _, p := range firstPath {
		unavoidable[p] = void{}
	}
	for _, p := range firstPath {
		if _, f := unavoidable[p]; !f {
			// We already found a path without this one.
			continue
		}
		path := islandDFS(s, island, src, dst, rem, map[Pos]int{p: 999999999})
		if len(path) == 0 {
			// This one is unavoidable
			continue
		}
		for p := range unavoidable {
			if !inPath(path, p) {
				delete(unavoidable, p)
			}
		}
	}
	return unavoidable
}

func islandDFS(s State, island byte, p, dst Pos, rem int, beenThere map[Pos]int) []Pos {
	beenThere[p] = rem
	if rem == 0 {
		return nil
	}
	for _, np := range s.SmartAround(p, dst) {
		if s.Get(np).Water() || (s.Get(np).KnownLand() && s.Get(np).Char() != island) {
			continue
		}
		if d, ok := beenThere[np]; ok {
			if d >= rem {
				continue
			}
		}
		if np == dst {
			return []Pos{dst}
		}
		sub := 1
		if s.Get(np).Char() == island {
			sub = 0
		}
		path := islandDFS(s, island, np, dst, rem-sub, beenThere)
		if len(path) > 0 {
			return append(path, np)
		}
	}
	return nil
}
