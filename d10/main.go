package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	debug = flag.Bool("debug", false, "debug mode")
)

type Pos struct {
	x, y int
}

func (p Pos) String() string {
	return fmt.Sprintf("(%d, %d)", p.x, p.y)
}

type Cell struct {
	height int
	p      Pos
}

func (c *Cell) String() string {
	return fmt.Sprintf("%d %s", c.height, c.p)
}

type Map struct {
	cells         map[Pos]*Cell
	height, width int
	trailheads    []Trailhead
}

func NewMap() *Map {
	return &Map{
		cells:  make(map[Pos]*Cell),
		height: -1,
		width:  -1,
	}
}

func (m *Map) Print() {
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			fmt.Printf("%d", m.cells[Pos{x, y}].height)
		}
		fmt.Println()
	}
}

func (m *Map) Route() {
	for _, cell := range m.cells {
		if cell.height != 0 {
			continue
		}
		trails := m.FindTrailsFrom(cell)
		if len(trails) > 0 {
			m.trailheads = append(m.trailheads, Trailhead{
				p:      cell.p,
				trails: trails,
			})
		}
	}
}

func (m *Map) FindTrailsFrom(head *Cell) []Trail {
	candidates := []Trail{Trail{head}}
	var trails []Trail

	for len(candidates) > 0 {
		if *debug {
			fmt.Printf("candidates: %v\n", candidates)
		}
		c := candidates[0]
		candidates = candidates[1:]
		tail := c[len(c)-1]
		neighbours := m.Neighbours(tail, func(n *Cell) bool {
			return n.height == tail.height+1
		})
		if *debug {
			fmt.Printf("c: %v\ntail: %v\nneighbours: %v\n", c, tail, neighbours)
		}
		for _, n := range neighbours {
			trail := append(Trail{}, c...)
			trail = append(trail, n)
			if trail.Valid() {
				if *debug {
					fmt.Printf("trail valid: %s\n", trail)
				}
				trails = append(trails, trail)
			} else if trail.Incomplete() {
				if *debug {
					fmt.Printf("trail incomplete: %s\n", trail)
				}
				candidates = append(candidates, trail)
			} else {
				log.Fatalf("bad trail %v", trail)
			}
		}
	}

	return trails
}

func (m *Map) Neighbours(cell *Cell, match func(*Cell) bool) []*Cell {
	var ret []*Cell
	nesw := []*Cell{
		m.cells[Pos{cell.p.x, cell.p.y - 1}],
		m.cells[Pos{cell.p.x + 1, cell.p.y}],
		m.cells[Pos{cell.p.x, cell.p.y + 1}],
		m.cells[Pos{cell.p.x - 1, cell.p.y}],
	}
	for _, c := range nesw {
		if c != nil && match(c) {
			ret = append(ret, c)
		}
	}
	return ret
}

type Trailhead struct {
	p      Pos
	trails []Trail
}

func (th Trailhead) Score() int {
	summits := make(map[Pos]bool, len(th.trails))
	for _, t := range th.trails {
		summits[t[len(t)-1].p] = true
	}
	return len(summits)
}

type Trail []*Cell

func (t Trail) String() string {
	var b strings.Builder
	b.WriteString("{ ")
	for i, step := range t {
		b.WriteString(step.String())
		if i < len(t)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString(" }")
	return b.String()
}

func (t Trail) Valid() bool {
	if len(t) != 10 {
		return false
	}
	return t.validSteps()
}

func (t Trail) Incomplete() bool {
	if len(t) == 0 || len(t) > 9 {
		return false
	}
	return t.validSteps()
}

func (t Trail) validSteps() bool {
	for i, step := range t {
		if step.height != i {
			return false
		}
	}
	return true
}

func main() {
	flag.Parse()
	filename := "example.txt"
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
	}
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	m := NewMap()

	s := bufio.NewScanner(f)
	y := 0
	for s.Scan() {
		line := s.Text()
		m.width = len(line)
		for x, e := range line {
			if e < '0' || e > '9' {
				log.Fatalf("invalid entry %q", e)
			}
			height := int(e - '0')
			pos := Pos{x, y}
			m.cells[pos] = &Cell{
				height: height,
				p:      Pos{x, y},
			}
		}
		y++
	}
	m.height = y
	if *debug {
		m.Print()
	}
	m.Route()
	score, ratings := 0, 0
	fmt.Printf("trailheads: %d\n", len(m.trailheads))
	for _, th := range m.trailheads {
		for _, t := range th.trails {
			if !t.Valid() {
				log.Fatalf("invalid trail %v in trailhead %v", t, th)
			}
		}
		s, r := th.Score(), len(th.trails)
		if *debug {
			fmt.Printf("trailhead %s: score %d/rating %d\n", th.p, s, r)
			for _, trail := range th.trails {
				fmt.Println(trail)
			}
		}
		score += s
		ratings += r
	}
	fmt.Printf("total score: %d\ntotal ratings: %d\n", score, ratings)
}
