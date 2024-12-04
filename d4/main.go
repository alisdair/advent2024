package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	debug      = flag.Bool("debug", false, "enable debug logging")
	forceColor = flag.Bool("force-color", false, "force color output")
)

func get[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

type coord struct {
	x int
	y int
}

type cell struct {
	c coord
	b byte
}

func (c coord) String() string {
	return fmt.Sprintf("(%d, %d)", c.x, c.y)
}

type match struct {
	cells  []cell
	finder string
}

func (m *match) Reset() {
	m.cells = nil
}

func (m *match) Add(c cell) {
	m.cells = append(m.cells, c)
}

func (m *match) Equals(o *match) bool {
	if len(m.cells) != len(o.cells) {
		return false
	}
	for i := range m.cells {
		if m.cells[i] != o.cells[i] {
			return false
		}
	}
	return true
}

func (m *match) Contains(needle coord) bool {
	for _, c := range m.cells {
		if c.c == needle {
			return true
		}
	}
	return false
}

func (m *match) String() string {
	var b strings.Builder
	// b.WriteString(m.finder)
	// b.WriteString(": ")
	// for _, c := range m.cells {
	// 	b.WriteString(fmt.Sprintf("%c", c.b))
	// }
	// b.WriteRune(' ')
	b.WriteString(m.cells[0].c.String())
	b.WriteRune('-')
	b.WriteString(m.cells[len(m.cells)-1].c.String())
	return b.String()
}

func (c coord) inbounds(g grid) bool {
	return c.y >= 0 && c.y < len(g) && c.x >= 0 && c.x < len(g[c.y])
}

type finder struct {
	name  string
	start func(g grid) coord
	next  func(g grid, i coord) (ii coord, reset bool)
}

func (f finder) search(g grid, word string) []*match {
	var matches []*match
	var reset bool
	got := &match{finder: f.name}

	i := f.start(g)
	for i.inbounds(g) {
		want := word[len(got.cells)]
		next := g.at(i)
		if next == want {
			got.Add(cell{i, want})
		}
		if len(got.cells) == len(word) {
			if *debug {
				g.print(got)
			}
			matches = append(matches, got)
			got = &match{finder: f.name}
		}
		if next != want && len(got.cells) > 0 {
			got.Reset()
			// We can restart the search from the current coordinate because our
			// target word is prefix-free. Otherwise we'd need to unwind through
			// got to start at its 1 index.
			continue
		}
		i, reset = f.next(g, i)
		if reset {
			got.Reset()
		}
	}
	return matches
}

func (g grid) print(got *match) {
	fmt.Println(got)

	highlight := color.New(color.FgRed, color.Bold)
	background := color.New(color.FgWhite, color.Bold)
	for y := range g {
		for x := range g[y] {
			c := coord{x, y}
			if got != nil && got.Contains(c) {
				highlight.Printf("%c ", g.at(c))
			} else {
				background.Printf("%c ", g.at(c))
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

var ltr = finder{
	name: "ltr",
	start: func(g grid) coord {
		return coord{0, 0}
	},
	next: func(g grid, i coord) (ii coord, reset bool) {
		ii = coord{i.x + 1, i.y}
		if ii.x >= len(g) {
			ii = coord{0, i.y + 1}
			reset = true
		}
		if ii.y >= len(g) {
			return coord{-1, -1}, false
		}
		return ii, reset
	},
}

var rtl = finder{
	name: "rtl",
	start: func(g grid) coord {
		return coord{len(g) - 1, len(g) - 1}
	},
	next: func(g grid, i coord) (ii coord, reset bool) {
		ii = coord{i.x - 1, i.y}
		if ii.x < 0 {
			ii = coord{len(g) - 1, i.y - 1}
			reset = true
		}
		if ii.y >= len(g) {
			return coord{-1, -1}, false
		}
		return ii, reset
	},
}

var down = finder{
	name: "down",
	start: func(g grid) coord {
		return coord{0, 0}
	},
	next: func(g grid, i coord) (ii coord, reset bool) {
		ii = coord{i.x, i.y + 1}
		if ii.y >= len(g) {
			ii = coord{i.x + 1, 0}
			reset = true
		}
		return ii, reset
	},
}

var up = finder{
	name: "up",
	start: func(g grid) coord {
		return coord{len(g) - 1, len(g) - 1}
	},
	next: func(g grid, i coord) (ii coord, reset bool) {
		ii = coord{i.x, i.y - 1}
		if ii.y < 0 {
			ii = coord{i.x - 1, len(g) - 1}
			reset = true
		}
		return ii, reset
	},
}

var dr = finder{
	name: "dr",
	start: func(g grid) coord {
		return coord{0, len(g) - 1}
	},
	next: func(g grid, i coord) (ii coord, reset bool) {
		ii = coord{i.x + 1, i.y + 1}
		if ii.x >= len(g) || ii.y >= len(g) {
			ii.x++
			for ii.x > 0 && ii.y > 0 {
				ii.x--
				ii.y--
			}
			reset = true
		}
		return ii, reset
	},
}

var ul = finder{
	name: "ul",
	start: func(g grid) coord {
		return coord{len(g) - 1, 0}
	},
	next: func(g grid, i coord) (ii coord, reset bool) {
		ii = coord{i.x - 1, i.y - 1}
		if ii.x < 0 || ii.y < 0 {
			ii.x--
			for ii.x < len(g)-1 && ii.y < len(g)-1 {
				ii.x++
				ii.y++
			}
			reset = true
		}
		return ii, reset
	},
}

var ur = finder{
	name: "ur",
	start: func(g grid) coord {
		return coord{len(g) - 1, len(g) - 1}
	},
	next: func(g grid, i coord) (ii coord, reset bool) {
		ii = coord{i.x + 1, i.y - 1}
		if ii.x > len(g)-1 || ii.y < 0 {
			ii.x--
			for ii.x > 0 && ii.y < len(g)-1 {
				ii.x--
				ii.y++
			}
			reset = true
		}
		return ii, reset
	},
}

var dl = finder{
	name: "dl",
	start: func(g grid) coord {
		return coord{0, 0}
	},
	next: func(g grid, i coord) (ii coord, reset bool) {
		ii = coord{i.x - 1, i.y + 1}
		if ii.x < 0 || ii.y > len(g)-1 {
			ii.x++
			for ii.x < len(g)-1 && ii.y > 0 {
				ii.x++
				ii.y--
			}
			reset = true
		}
		return ii, reset
	},
}

func xmatcher(g grid) []*match {
	var matches []*match

	// Loop offset by 1 because we can't match on the edges
	for y := 1; y < len(g)-1; y++ {
		for x := 1; x < len(g[y])-1; x++ {
			o := coord{x, y}
			if g.at(o) != 'A' {
				continue
			}
			ne := coord{x + 1, y - 1}
			se := coord{x + 1, y + 1}
			sw := coord{x - 1, y + 1}
			nw := coord{x - 1, y - 1}

			nwse := g.at(nw) == 'M' && g.at(se) == 'S' || g.at(nw) == 'S' && g.at(se) == 'M'
			nesw := g.at(ne) == 'M' && g.at(sw) == 'S' || g.at(ne) == 'S' && g.at(sw) == 'M'

			if nwse && nesw {
				got := &match{
					finder: "xmatch",
					cells: []cell{
						{o, g.at(o)},
						{ne, g.at(ne)},
						{se, g.at(se)},
						{sw, g.at(sw)},
						{nw, g.at(nw)},
					},
				}
				if *debug {
					g.print(got)
				}
				matches = append(matches, got)
			}
		}
	}
	return matches
}

func (g grid) at(c coord) byte {
	return g[c.y][c.x]
}

type grid [][]byte

func main() {
	flag.Parse()

	if *forceColor {
		color.NoColor = false
	}

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: d4 [filepath]")
		flag.Usage()
	}

	file := get(os.Open(args[0]))
	s := bufio.NewScanner(file)
	var g grid
	for s.Scan() {
		line := s.Bytes()
		row := make([]byte, len(line))
		copy(row, line)
		g = append(g, row)
	}

	var matches []*match

	want := "XMAS"
	for _, f := range []finder{down, ltr, rtl, up, dr, ul, ur, dl} {
		matches = append(matches, f.search(g, want)...)
	}

	fmt.Printf("total: %d\n", len(matches))

	xmatches := xmatcher(g)

	fmt.Printf("x-mas: %d\n", len(xmatches))
}
