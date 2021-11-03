package main

func init() {
	Register(ExpandIslands, "expandIslands", "expand")
}

func ExpandIslands(s State) State {
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
	for l, r := range remaining {
		if r == 0 {
			continue
		}
		bodies := GroupBodies(s, islandTiles[l])
	bodyLoop:
		for _, b := range bodies {
			expansionOption := Pos{-1, -1}
			for p := range b {
				for _, np := range s.Around(p) {
					if !s.Get(np).FullyKnown() {
						if expansionOption.x != -1 {
							// Two+ options. Can't decide yet.
							continue bodyLoop
						}
						expansionOption = np
					}
				}
			}
			if expansionOption.x != -1 {
				s.Set(expansionOption, Cell(l))
			}
		}
	}
	return s
}
