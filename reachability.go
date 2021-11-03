package main

func init() {
	Register(UnreachableByLand, "unreachableByLand", "unreachable")
	Register(ResolveUnknownLand, "resolveUnknownLand", "resolve")
}

func UnreachableByLand(s State) State {
	s = s.Clone()
	rm := reachabilityMap(s)
	for y, r := range s.board {
		for x, c := range r {
			if c.Known() {
				continue
			}
			options := rm[Pos{x, y}]
			if len(options) == 0 {
				s.Set(Pos{x, y}, Water)
			}
		}
	}
	return s
}

func ResolveUnknownLand(s State) State {
	s = s.Clone()
	rm := reachabilityMap(s)
	for y, r := range s.board {
		for x, c := range r {
			if c != SomeLand {
				continue
			}
			options := rm[Pos{x, y}]
			if len(options) == 1 {
				s.Set(Pos{x, y}, Cell(options[0]))
			}
		}
	}
	return s
}

func reachabilityMap(s State) map[Pos][]byte {
	perIsland := map[byte]map[Pos]int{}
	remaining := s.UndiscoveredIslands()
	for l := range remaining {
		perIsland[l] = map[Pos]int{}
	}
	for y, r := range s.board {
		for x, c := range r {
			if !c.KnownLand() {
				continue
			}
			l := c.Char()
			landDFS(s, l, Pos{x, y}, remaining[l], perIsland[l])
		}
	}
	ret := map[Pos][]byte{}
	for l, m := range perIsland {
		for p := range m {
			ret[p] = append(ret[p], l)
		}
	}
	return ret
}

func landDFS(s State, island byte, p Pos, rem int, been map[Pos]int) {
	been[p] = rem
	if rem == 0 {
		return
	}
around:
	for _, np := range s.Around(p) {
		if s.Get(np).FullyKnown() {
			continue
		}
		if d, ok := been[np]; ok {
			if d >= rem {
				continue
			}
		}
		for _, nnp := range s.Around(np) {
			if s.Get(nnp).KnownLand() && s.Get(nnp).Char() != island {
				continue around
			}
		}
		landDFS(s, island, np, rem-1, been)
	}
}
