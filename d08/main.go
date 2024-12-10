package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	tm "github.com/buger/goterm"
)

var (
	debug = flag.Bool("debug", true, "show map")
	full  = flag.Bool("full", true, "full set of antinodes")
)

type Pos struct {
	x, y int
}

type Map struct {
	width, height int
	antennas      map[rune][]Pos
	antinodes     map[Pos]bool
}

func NewMap() *Map {
	return &Map{
		antennas:  make(map[rune][]Pos),
		antinodes: make(map[Pos]bool),
	}
}

func (m *Map) Print() {
	tm.Clear()

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			tm.MoveCursor(x*2+1, y+1)
			cell := tm.Color(".", tm.WHITE)
			if m.antinodes[Pos{x, y}] {
				cell = tm.Background(cell, tm.RED)
			}
			tm.Print(cell)
		}
	}

	for f, ps := range m.antennas {
		for _, p := range ps {
			tm.MoveCursor(p.x*2+1, p.y+1)
			cell := tm.Color(fmt.Sprintf("%c", f), tm.GREEN)
			if m.antinodes[p] {
				cell = tm.Background(cell, tm.RED)
			}
			tm.Print(cell)
		}
	}

	tm.MoveCursor(1, m.height+1)
	tm.Println()

	tm.Flush()
}

func (m *Map) FindAntinodes(full bool) {
	for _, ps := range m.antennas {
		for _, p0 := range ps {
			for _, p1 := range ps {
				if p0 == p1 {
					continue
				}
				dx := p1.x - p0.x
				dy := p1.y - p0.y

				if full {
					i, j := 0, 0
					for m.AddAntinode(p0.x-i, p0.y-j) {
						i += dx
						j += dy
					}
					i, j = 0, 0
					for m.AddAntinode(p1.x+i, p1.y+j) {
						i += dx
						j += dy
					}
				} else {
					m.AddAntinode(p0.x-dx, p0.y-dy)
					m.AddAntinode(p1.x+dx, p1.y+dy)
				}
			}
		}
	}
}

func (m *Map) AddAntinode(x, y int) bool {
	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return false
	}
	m.antinodes[Pos{x, y}] = true
	return true
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

	m := NewMap()

	s := bufio.NewScanner(f)
	y := 0
	for s.Scan() {
		line := s.Text()
		m.width = len(line)
		for x, c := range line {
			if c == '.' {
				continue
			}
			m.antennas[c] = append(m.antennas[c], Pos{x, y})
		}
		y++
	}
	m.height = y

	m.FindAntinodes(*full)

	if *debug {
		m.Print()
	}

	fmt.Printf("\nantinodes: %d\n", len(m.antinodes))
}
