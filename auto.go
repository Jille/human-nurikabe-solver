package main

import "fmt"

func init() {
	Register(autoPickStrategy, "autoPickStrategy", "auto")
}

func autoPickStrategy(s State) State {
	ordered := []string{
		"resolveUnknownLand",
		"surroundCompletedIslands",
		"avoidMerge",
		"expandIslands",
		"antiPool",
		"waterConnectable",
		"mendIslands",
		"unreachableByLand",
		"islandShaper",
	}
	for _, n := range ordered {
		if n == "listUsefulStrategies" || n == "autoPickStrategy" {
			continue
		}
		st := strategies[n]
		ss := st(s)
		if !ss.Equal(s) {
			fmt.Printf("==> Chose %s\n", n)
			return ss
		}
	}
	return s
}
