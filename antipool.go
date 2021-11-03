package main

func init() {
	Register(AntiPool, "antiPool", "pool")
}

func AntiPool(s State) State {
	s = s.Clone()
	for y, r := range s.board {
		if y == 0 {
			continue
		}
	cell:
		for x := range r {
			if x == 0 {
				continue
			}
			waters := 0
			var lastUnknown Pos
			for _, p := range []Pos{{x - 1, y - 1}, {x - 1, y}, {x, y - 1}, {x, y}} {
				if s.Get(p).Land() {
					continue cell
				} else if s.Get(p).Water() {
					waters++
				} else {
					lastUnknown = p
				}
			}
			if waters == 3 {
				s.Set(lastUnknown, '=')
			}
		}
	}
	return s
}
