package main

func init() {
	Register(ExpandIslands, "expandIslands", "expand")
}

func ExpandIslands(state State) State {
	state = state.Clone()
	remainingLands := state.UndiscoveredIslands()
	islandTiles := map[byte][]Pos{}
	expansions := map[byte][]Pos{}

	for y, row := range state.board {
		for x, cell := range row {
			if !cell.Land() || cell == SomeLand {
				continue
			}
			islandTiles[cell.Char()] = append(islandTiles[cell.Char()], Pos{x, y})
		}
	}
	for land, remaining := range remainingLands {
		if remaining == 0 {
			continue
		}
		bodies := GroupBodies(state, islandTiles[land])
	bodyLoop:
		for _, body := range bodies {
			expansionOption := Pos{-1, -1}
			for pos := range body {
				for _, newPos := range state.Around(pos) {
					if !state.Get(newPos).FullyKnown() {
						if expansionOption.x != -1 {
							// Two+ options. Can't decide yet.
							continue bodyLoop
						}
						expansionOption = newPos
					}
				}
			}
			if expansionOption.x != -1 {
				expansions[land] = append(expansions[land], expansionOption)
			}
		}
	}

	for land, positions := range expansions {
		for _, pos := range positions {
			state.Set(pos, Cell(land))
		}
	}

	return state
}
