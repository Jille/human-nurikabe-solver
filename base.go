package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
)

type void struct{}

type Pos struct{ x, y int }

func (p Pos) Distance(o Pos) int {
	return abs(p.x-o.x) + abs(p.y-o.y)
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

type Cell byte

const (
	Unknown  Cell = ' '
	Water    Cell = '~'
	SomeLand Cell = '='
)

func (c Cell) MaybeLand() bool {
	return c != Water
}

func (c Cell) MaybeWater() bool {
	return c == Unknown || c == Water
}

func (c Cell) Known() bool {
	return c != Unknown
}

func (c Cell) FullyKnown() bool {
	return c != Unknown && c != SomeLand
}

func (c Cell) Land() bool {
	return c != Unknown && c != Water
}

func (c Cell) KnownLand() bool {
	return c != Unknown && c != Water && c != SomeLand
}

func (c Cell) Water() bool {
	return c == Water
}

func (c Cell) Char() byte {
	return byte(c)
}

type State struct {
	board [][]Cell
	sizes map[byte]int
}

func (s State) Get(p Pos) Cell {
	return s.board[p.y][p.x]
}

func (s State) Set(p Pos, c Cell) {
	s.board[p.y][p.x] = c
}

func (s State) Around(p Pos) []Pos {
	ret := make([]Pos, 0, 4)
	if p.x > 0 {
		ret = append(ret, Pos{p.x - 1, p.y})
	}
	if p.y > 0 {
		ret = append(ret, Pos{p.x, p.y - 1})
	}
	if p.x < len(s.board[0])-1 {
		ret = append(ret, Pos{p.x + 1, p.y})
	}
	if p.y < len(s.board)-1 {
		ret = append(ret, Pos{p.x, p.y + 1})
	}
	return ret
}

func (s State) SmartAround(p, target Pos) []Pos {
	var closer []Pos
	var equal []Pos
	var farther []Pos
	d := p.Distance(target)
	for _, ap := range s.Around(p) {
		ad := ap.Distance(target)
		if d == ad {
			equal = append(equal, ap)
		} else if d < ad {
			farther = append(farther, ap)
		} else {
			closer = append(closer, ap)
		}
	}
	ret := make([]Pos, 0, 4)
	ret = append(ret, closer...)
	ret = append(ret, equal...)
	ret = append(ret, farther...)
	return ret
}

func (s State) UndiscoveredIslands() map[byte]int {
	remaining := map[byte]int{}
	for l, s := range s.sizes {
		remaining[l] = s
	}
	for _, r := range s.board {
		for _, c := range r {
			if !c.Land() || c == SomeLand {
				continue
			}
			remaining[c.Char()]--
		}
	}
	return remaining
}

func (c State) Clone() State {
	n := State{
		sizes: c.sizes,
	}
	for _, r := range c.board {
		n.board = append(n.board, append([]Cell(nil), r...))
	}
	return n
}

func (c State) Equal(o State) bool {
	if len(c.board) != len(o.board) {
		return false
	}
	if len(c.board[0]) != len(o.board[0]) {
		return false
	}
	for y, r := range c.board {
		for x, c := range r {
			if c != o.board[y][x] {
				return false
			}
		}
	}
	return true
}

func (c State) Finished() bool {
	for _, r := range c.board {
		for _, c := range r {
			if !c.FullyKnown() {
				return false
			}
		}
	}
	return true
}

func (s State) Print(w io.Writer) {
	width := len(s.board[0])
	fmt.Fprintf(w, "+"+strings.Repeat("-", width*3)+"+\n")
	for _, r := range s.board {
		fmt.Fprintf(w, "|")
		for _, c := range r {
			if !c.Known() {
				fmt.Fprintf(w, "   ")
			} else if c.Water() {
				fmt.Fprintf(w, " ~ ")
			} else if c == SomeLand {
				fmt.Fprintf(w, "? ?")
			} else {
				fmt.Fprintf(w, "%c%2d", c.Char(), s.sizes[c.Char()])
			}
		}
		fmt.Fprintf(w, "|\n")
	}
	fmt.Fprintf(w, "+"+strings.Repeat("-", width*3)+"+\n")
}

func ReadGame(r io.Reader) (State, error) {
	scanner := bufio.NewScanner(r)
	s := State{
		sizes: map[byte]int{},
	}
	scanner.Scan() // Discard first line
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "+") {
			break
		}
		l := strings.Trim(scanner.Text(), "|")
		var row []Cell
		for len(l) >= 3 {
			switch l[:3] {
			case "   ":
				row = append(row, Unknown)
			case " ~ ":
				row = append(row, Water)
			case "? ?":
				row = append(row, SomeLand)
			default:
				row = append(row, Cell(l[0]))
				sz, err := strconv.Atoi(strings.TrimSpace(l[1:3]))
				if err != nil {
					return s, err
				}
				s.sizes[l[0]] = sz
			}
			l = l[3:]
		}
		s.board = append(s.board, row)
	}
	return s, scanner.Err()
}

type StrategyFunc func(State) State

var strategies = map[string]StrategyFunc{}
var aliases = map[string]StrategyFunc{}

func Register(f StrategyFunc, name string, shortcuts ...string) {
	strategies[name] = f
	aliases[strings.ToLower(name)] = f
	for _, a := range shortcuts {
		aliases[strings.ToLower(a)] = f
	}
}

func listUsefulStrategies(s State) State {
	fmt.Println("Useful strategies:")
	for n, st := range strategies {
		if n == "listUsefulStrategies" || n == "autoPickStrategy" {
			continue
		}
		ss := st(s)
		if !ss.Equal(s) {
			fmt.Printf("* %s\n", n)
		}
	}
	return s
}

func init() {
	Register(listUsefulStrategies, "listUsefulStrategies", "list")
}

func main() {
	flag.Parse()
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	s, err := ReadGame(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	s.Print(os.Stdout)

	for _, n := range flag.Args() {
		st, ok := aliases[strings.ToLower(n)]
		if !ok {
			log.Fatalf("Unknown strategy %q", n)
		}
		ns := st(s)
		if ns.Equal(s) {
			fmt.Printf("%q was not very effective\n", n)
			continue
		}
		s = ns
		s.Print(os.Stdout)
		if s.Finished() {
			fmt.Println("Whoo!")
			break
		}
	}

	/*
		s = AvoidMerge(s)
		s.Print(os.Stdout)
		s = SurroundCompletedIslands(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
		s = WaterConnectable(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
		s = SurroundCompletedIslands(s)
		s.Print(os.Stdout)
		s = WaterConnectable(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
		s = AvoidMerge(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
		s = SurroundCompletedIslands(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
		s = WaterConnectable(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
		s = WaterConnectable(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
		s = SurroundCompletedIslands(s)
		s.Print(os.Stdout)
		s = WaterConnectable(s)
		s.Print(os.Stdout)
		s = UnreachableByLand(s)
		s.Print(os.Stdout)
		s = AntiPool(s)
		s.Print(os.Stdout)
		s = UnreachableByLand(s)
		s.Print(os.Stdout)
		s = WaterConnectable(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
		s = ExpandIslands(s)
		s.Print(os.Stdout)
	*/
}
