package main

import (
	"context"
	"runtime"
	"sync"

	"golang.org/x/sync/semaphore"
)

func init() {
	Register(WaterConnectable, "waterConnectable", "wc", "water")
}

func WaterConnectable(s State) State {
	s = s.Clone()
	var waters []Pos
	for y, r := range s.board {
		for x, c := range r {
			if c.Water() {
				waters = append(waters, Pos{x, y})
			}
		}
	}
	var mtx sync.Mutex
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	newWater := map[Pos]void{}
	waterBodies := GroupBodies(s, waters)
	// fmt.Printf("Water bodies: %v\n", waterBodies)
	for i, w1 := range waterBodies {
		for _, w2 := range waterBodies[i+1:] {
			w1 := w1
			w2 := w2
			wg.Add(1)
			sem.Acquire(context.Background(), 1)
			go func() {
				nw := findWaterPaths(s, w1.Any(), w2.Any())
				sem.Release(1)
				mtx.Lock()
				for p := range nw {
					newWater[p] = void{}
				}
				mtx.Unlock()
				wg.Done()
			}()
		}
	}
	wg.Wait()
	for p := range newWater {
		s.Set(p, Water)
	}
	return s
}

/*
func findWaterPaths(s State, dst, src Pos) map[Pos]void {
	// fmt.Printf("findWaterPaths(%v -> %v)\n", src, dst)
	var unavoidable map[Pos]void
	cb := func(path []Pos) bool {
		// fmt.Printf("Path: %v\n", path)
		if unavoidable == nil {
			unavoidable = map[Pos]void{}
			for _, p := range path {
				unavoidable[p] = void{}
			}
		} else {
			for p := range unavoidable {
				if !inPath(path, p) {
					delete(unavoidable, p)
				}
			}
		}
		// fmt.Printf("Remaining unavoidables: %v\n", unavoidable)
		return len(unavoidable) > 0
	}
	BFS(s, src, dst, cb)
	return unavoidable
}

func BFS(s State, src, dst Pos, cb func(path []Pos) bool) {
	dfsMode := false
	queue := [][]Pos{[]Pos{src}}
	for len(queue) > 0 {
		var path []Pos
		if dfsMode {
			path = queue[len(queue)-1]
			queue = queue[:len(queue)-1]
		} else {
			path = queue[0]
			queue = queue[1:]
		}
		p := path[0]
		for _, np := range s.Around(p) {
			if s.Get(np).Land() {
				continue
			}
			if np == dst {
				if !cb(path) {
					return
				}
				dfsMode = !dfsMode
				break
			}
			if inPath(path, np) {
				continue
			}
			queue = append(queue, append([]Pos{np}, path...))
		}
	}
}
*/

func inPath(path []Pos, p Pos) bool {
	for _, pp := range path {
		if p == pp {
			return true
		}
	}
	return false
}

func findWaterPaths(s State, dst, src Pos) map[Pos]void {
	// fmt.Printf("findWaterPaths(%v -> %v)\n", src, dst)
	firstPath := DFS(s, src, dst, map[Pos]void{})
	unavoidable := map[Pos]void{}
	for _, p := range firstPath {
		unavoidable[p] = void{}
	}
	for _, p := range firstPath {
		if _, f := unavoidable[p]; !f {
			// We already found a path without this one.
			continue
		}
		path := DFS(s, src, dst, map[Pos]void{p: void{}})
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

func DFS(s State, p, dst Pos, beenThere map[Pos]void) []Pos {
	beenThere[p] = void{}
	for _, np := range s.SmartAround(p, dst) {
		if s.Get(np).Land() {
			continue
		}
		if _, f := beenThere[np]; f {
			continue
		}
		if np == dst {
			return []Pos{dst}
		}
		path := DFS(s, np, dst, beenThere)
		if len(path) > 0 {
			return append(path, np)
		}
	}
	return nil
}
