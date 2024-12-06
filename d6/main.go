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

var debug = flag.Bool("debug", false, "debug mode")

type Cell rune

const (
	unvisited  Cell = '.'
	visited    Cell = 'X'
	obstacle   Cell = '#'
	guardUp    Cell = '^'
	guardRight Cell = '>'
	guardDown  Cell = 'v'
	guardLeft  Cell = '<'
)

func NewCell(r rune) (Cell, error) {
	c := Cell(r)
	switch c {
	case unvisited, visited, obstacle, guardUp, guardRight, guardDown, guardLeft:
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

func (c Cell) NextGuard() Cell {
	switch c {
	case guardUp:
		return guardRight
	case guardRight:
		return guardDown
	case guardDown:
		return guardLeft
	case guardLeft:
		return guardUp
	default:
		log.Fatalf("next guard for %q invalid", c)
		return c
	}
}

type Pos struct {
	x, y int
}

func (p Pos) Up() Pos    { return Pos{p.x, p.y - 1} }
func (p Pos) Right() Pos { return Pos{p.x + 1, p.y} }
func (p Pos) Down() Pos  { return Pos{p.x, p.y + 1} }
func (p Pos) Left() Pos  { return Pos{p.x - 1, p.y} }

type Grid struct {
	cells         [][]Cell
	width, height int
	guard         *Pos
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

	if g.guard != nil {
		guard, ok := g.At(*g.guard)
		if ok && !guard.IsGuard() {
			log.Fatalf("invalid guard %c at %v", guard, *g.guard)
		}
	}
}

func (g *Grid) Render() {
	tm.MoveCursor(1, 1)

	for _, row := range g.cells {
		for _, cell := range row {
			tm.Printf("%c ", cell)
		}
		tm.Println()
	}

	tm.Flush()
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

	guard, ok := g.At(*g.guard)
	if !ok {
		return
	}

	var next Pos
	switch guard {
	case guardUp:
		next = g.guard.Up()
	case guardRight:
		next = g.guard.Right()
	case guardDown:
		next = g.guard.Down()
	case guardLeft:
		next = g.guard.Left()
	default:
		log.Fatalf("unexpected guard %c at %v", guard, *g.guard)
	}

	target, ok := g.At(next)
	if !ok {
		// Bye!
		g.Set(*g.guard, visited)
		g.guard = nil
	}
	switch target {
	case unvisited, visited:
		g.Set(*g.guard, visited)
		g.Set(next, guard)
		g.guard = &next
	case obstacle:
		guard = guard.NextGuard()
		g.Set(*g.guard, guard)
	}
}

func (g *Grid) Visited() int {
	ret := 0
	for _, row := range g.cells {
		for _, cell := range row {
			if cell == visited {
				ret++
			}
		}
	}
	return ret
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
	var guard *Pos

	s := bufio.NewScanner(f)
	y := 0
	for s.Scan() {
		line := s.Text()
		var row []Cell
		for x, l := range line {
			c := get(NewCell(l))
			if c.IsGuard() {
				guard = &Pos{x, y}
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
		guard:  guard,
	}
	grid.Validate()

	tm.Clear()
	for grid.guard != nil {
		if *debug {
			grid.Render()
			time.Sleep(time.Millisecond * 33)
		}
		grid.Iterate()
	}

	fmt.Printf("visited: %d\n", grid.Visited())
}
