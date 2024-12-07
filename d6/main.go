package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	tm "github.com/buger/goterm"
)

var (
	debug    = flag.Bool("debug", false, "debug mode")
	pristine = flag.Bool("pristine", false, "don't modify the grid")
)

type Cell rune

const (
	unvisited      Cell = '.'
	visitedUp      Cell = '1'
	visitedDown    Cell = 'l'
	visitedLeft    Cell = '-'
	visitedRight   Cell = '_'
	obstruction    Cell = '#'
	newObstruction Cell = 'O'
	guardUp        Cell = '^'
	guardRight     Cell = '>'
	guardDown      Cell = 'v'
	guardLeft      Cell = '<'
)

func NewCell(r rune) (Cell, error) {
	c := Cell(r)
	switch c {
	case unvisited, obstruction, guardUp, guardRight, guardDown, guardLeft:
		return c, nil
	default:
		return c, fmt.Errorf("unknown cell %q", r)
	}
}

func (c Cell) IsGuard() bool {
	switch c {
	case guardUp, guardRight, guardDown, guardLeft:
		return true
	default:
		return false
	}
}

func (c Cell) IsVisited() bool {
	switch c {
	case visitedUp, visitedDown, visitedLeft, visitedRight:
		return true
	default:
		return false
	}
}

type Pos struct {
	x, y int
}

func (p Pos) Up() Pos    { return Pos{p.x, p.y - 1} }
func (p Pos) Right() Pos { return Pos{p.x + 1, p.y} }
func (p Pos) Down() Pos  { return Pos{p.x, p.y + 1} }
func (p Pos) Left() Pos  { return Pos{p.x - 1, p.y} }

type Guard struct {
	p       Pos
	c       Cell
	turning bool
}

func (g *Guard) Turn() {
	g.turning = true
	switch g.c {
	case guardUp:
		g.c = guardRight
	case guardRight:
		g.c = guardDown
	case guardDown:
		g.c = guardLeft
	case guardLeft:
		g.c = guardUp
	default:
		log.Fatalf("next guard for %q invalid", g.c)
	}
}

func (g *Guard) Direction() Cell {
	switch g.c {
	case guardUp:
		return visitedUp
	case guardDown:
		return visitedDown
	case guardLeft:
		return visitedLeft
	case guardRight:
		return visitedRight
	default:
		log.Fatalf("invalid guard %q", g.c)
		return unvisited
	}
}

func (g *Guard) Visited(grid *Grid) {
	c := g.Direction()
	if g.turning {
		g.turning = false
	}
	grid.Set(g.p, c)
}

type Grid struct {
	cells         [][]Cell
	width, height int
	guard         *Guard
}

func (g *Grid) ResetGuard(guard *Guard) {
	g.guard = &Guard{
		p: guard.p,
		c: guard.c,
	}
}

func (g *Grid) Reset() {
	for y, row := range g.cells {
		for x, cell := range row {
			if cell.IsVisited() || cell == newObstruction {
				pos := Pos{x, y}
				g.Set(pos, unvisited)
			}
		}
	}
}

func (g *Grid) Validate() {
	if got, want := len(g.cells), g.height; got != want {
		log.Fatalf("wrong height, want %d, got %d", want, got)
	}

	for i, row := range g.cells {
		if got, want := len(row), g.width; got != want {
			log.Fatalf("row %d: wrong width, want %d, got %d", i, want, got)
		}
	}
}

func (g *Grid) Render() {
	tm.MoveCursor(1, 1)

	for _, row := range g.cells {
		for _, cell := range row {
			switch cell {
			case visitedUp, visitedDown:
				cell = '|'
			case visitedLeft, visitedRight:
				cell = '-'
			}
			tm.Printf("%c ", cell)
		}
		tm.Println()
	}
	if g.guard != nil {
		tm.MoveCursor(g.guard.p.x*2+1, g.guard.p.y+1)
		tm.Printf("%c", g.guard.c)
	}

	tm.Flush()
}

func (g *Grid) RenderPlain() {
	for y, row := range g.cells {
		for x, cell := range row {
			pos := Pos{x, y}
			if g.guard != nil && g.guard.p == pos {
				fmt.Printf("%c ", g.guard.c)
			} else {
				switch cell {
				case visitedUp, visitedDown:
					cell = '|'
				case visitedLeft, visitedRight:
					cell = '-'
				}
				fmt.Printf("%c ", cell)
			}
		}
		fmt.Println()
	}
}

func (g *Grid) At(p Pos) (_ Cell, ok bool) {
	if p.y > g.height-1 || p.y < 0 || p.x > g.width-1 || p.x < 0 {
		return Cell(' '), false
	}
	return g.cells[p.y][p.x], true
}

func (g *Grid) Set(p Pos, c Cell) {
	g.cells[p.y][p.x] = c
}

func (g *Grid) Iterate() {
	if g.guard == nil {
		return
	}

	var next Pos
	switch g.guard.c {
	case guardUp:
		next = g.guard.p.Up()
	case guardRight:
		next = g.guard.p.Right()
	case guardDown:
		next = g.guard.p.Down()
	case guardLeft:
		next = g.guard.p.Left()
	default:
		log.Fatalf("unexpected guard %c at %v", g.guard.c, g.guard.p)
	}

	target, ok := g.At(next)
	if !ok {
		// Bye!
		g.guard.Visited(g)
		g.guard = nil
		return
	}
	switch target {
	case unvisited, visitedUp, visitedDown, visitedLeft, visitedRight:
		g.guard.Visited(g)
		g.guard.p = next
	case obstruction, newObstruction:
		g.guard.Turn()
	default:
		log.Fatalf("unexpected target cell %c", target)
	}
}

func (g *Grid) Visited() []Pos {
	var ret []Pos
	for x, row := range g.cells {
		for y, cell := range row {
			if cell.IsVisited() {
				ret = append(ret, Pos{x, y})
			}
		}
	}
	return ret
}

func (g *Grid) Stuck() bool {
	if g.guard == nil {
		return false
	}
	current, ok := g.At(g.guard.p)
	if !ok {
		return false
	}
	switch current {
	case g.guard.Direction():
		return true
	default:
		return false
	}
}

func get[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func main() {
	flag.Parse()

	filename := "example.txt"
	if args := flag.Args(); len(args) > 0 {
		filename = args[0]
	}

	f := get(os.Open(filename))
	defer f.Close()

	var cells [][]Cell
	guard := &Guard{}

	s := bufio.NewScanner(f)
	y := 0
	for s.Scan() {
		line := s.Text()
		var row []Cell
		for x, l := range line {
			c := get(NewCell(l))
			if c.IsGuard() {
				guard.c = c
				guard.p.x = x
				guard.p.y = y
				c = unvisited
			}
			row = append(row, c)
		}
		if len(row) > 0 {
			y++
			cells = append(cells, row)
		}
	}
	if len(cells) == 0 {
		log.Fatal("empty grid")
	}

	grid := &Grid{
		cells:  cells,
		width:  len(cells[0]),
		height: len(cells),
	}
	grid.Validate()

	run(grid, guard)
	visited := grid.Visited()
	fmt.Printf("\n\nvisited: %d\n", len(visited))

	if !*pristine {
		var obstructions []Pos

		grid.Reset()

		for y := 0; y < grid.height; y++ {
			for x := 0; x < grid.width; x++ {
				pos := Pos{x, y}
				if c, ok := grid.At(pos); !ok || c != unvisited {
					continue
				}
				grid.Set(pos, newObstruction)
				run(grid, guard)
				if grid.Stuck() {
					// fmt.Printf("\nfound looping obstruction at %v\n", pos)
					// grid.RenderPlain()
					// fmt.Printf("\n")
					obstructions = append(obstructions, pos)
				}
				grid.Reset()
			}
		}
		if *debug {
			grid.Reset()
			grid.ResetGuard(guard)
			for _, obstruction := range obstructions {
				grid.Set(obstruction, newObstruction)
			}
			grid.Render()
		}
		fmt.Printf("\n\n\nobstructions: %d\n", len(obstructions))
	}
}

func run(grid *Grid, guard *Guard) {
	grid.ResetGuard(guard)
	tm.Clear()
	for grid.guard != nil && !grid.Stuck() {
		if *debug {
			grid.Render()
			time.Sleep(time.Millisecond * 1)
		}
		grid.Iterate()
	}
	if *debug {
		grid.Render()
	}
}
