package main

func init() {
	Register(AvoidMerge, "avoidMerge")
}

func AvoidMerge(s State) State {
	s = s.Clone()
	for y, r := range s.board {
		for x, c := range r {
			if c.Known() {
				continue
			}
			neighbours := map[byte]void{}
			for _, np := range s.Around(Pos{x, y}) {
				if s.Get(np).KnownLand() {
					neighbours[s.Get(np).Char()] = void{}
				}
			}
			if len(neighbours) >= 2 {
				s.Set(Pos{x, y}, Water)
			}
		}
	}
	return s
}
