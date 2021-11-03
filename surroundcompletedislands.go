package main

func init() {
	Register(SurroundCompletedIslands, "surroundCompletedIslands", "surround", "complete")
}

func SurroundCompletedIslands(s State) State {
	s = s.Clone()
	found := map[byte]int{}
	for _, r := range s.board {
		for _, c := range r {
			if !c.KnownLand() {
				continue
			}
			found[c.Char()]++
		}
	}
	completed := map[byte]bool{}
	for l, got := range found {
		if got >= s.sizes[l] {
			completed[l] = true
		}
	}
	if len(completed) == 0 {
		return s
	}
	for y, r := range s.board {
		for x, c := range r {
			if !c.Land() {
				continue
			}
			if completed[c.Char()] {
				for _, np := range s.Around(Pos{x, y}) {
					if s.Get(np) == Unknown {
						s.Set(np, Water)
					}
				}
			}
		}
	}
	return s
}
