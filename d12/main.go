package main

import (
	"bufio"
	"flag"
	"fmt"
	"maps"
	"os"
)

var (
	debug = flag.Bool("debug", false, "extra logs please")
	sides = flag.Bool("sides", true, "price based on sides")
)

func pop[K comparable, V any](m map[K]V) K {
	for ret := range m {
		delete(m, ret)
		return ret
	}
	panic("popped empty set")
}

type Farm struct {
	plots         map[Pos]rune // coordinates => plant
	width, height int
}

func (f *Farm) regions() []*Region {
	var ret []*Region

	plots := make(map[Pos]rune, len(f.plots))
	maps.Copy(plots, f.plots)

	for p, plant := range plots {
		delete(plots, p)
		r := &Region{
			plant: plant,
			plots: make(map[Pos]bool),
		}

		next := map[Pos]bool{p: true}
		for len(next) > 0 {
			p := pop(next)
			delete(plots, p)

			r.plots[p] = true

			for _, n := range p.neighbours() {
				if plant, exists := plots[n]; exists && plant == r.plant {
					if _, ok := r.plots[n]; !ok {
						next[n] = true
					}
				}
			}
		}

		ret = append(ret, r)
	}

	return ret
}

type Pos struct {
	x, y int
}

func (p Pos) String() string {
	return fmt.Sprintf("(%d, %d)", p.x, p.y)
}

func (p Pos) neighbours() []Pos {
	return []Pos{
		{p.x - 1, p.y},
		{p.x + 1, p.y},
		{p.x, p.y - 1},
		{p.x, p.y + 1},
	}
}

func (p Pos) direction(o Pos) Direction {
	switch o {
	case Pos{p.x, p.y - 1}:
		return North
	case Pos{p.x + 1, p.y}:
		return East
	case Pos{p.x, p.y + 1}:
		return South
	case Pos{p.x - 1, p.y}:
		return West
	default:
		panic("invalid direction")
	}
}

func (p Pos) move(d Direction) Pos {
	switch d {
	case North:
		return Pos{p.x, p.y - 1}
	case East:
		return Pos{p.x + 1, p.y}
	case South:
		return Pos{p.x, p.y + 1}
	case West:
		return Pos{p.x - 1, p.y}
	default:
		panic("invalid direction")
	}
}

type Edge struct {
	p Pos
	d Direction
}

func (e Edge) String() string {
	return fmt.Sprintf("%s-%c", e.p, e.d)
}

type Direction rune

func (d Direction) String() string {
	return fmt.Sprintf("%c", d)
}

const (
	North Direction = 'N'
	East  Direction = 'E'
	South Direction = 'S'
	West  Direction = 'W'
)

type Region struct {
	plant rune
	plots map[Pos]bool
}

func (r *Region) perimeter() int {
	var perimeter int

	if len(r.plots) == 0 {
		return 0
	}

	var root Pos
	for k := range r.plots {
		root = k
		break
	}

	next := map[Pos]bool{root: true}
	seen := map[Pos]bool{}

	for len(next) > 0 {
		p := pop(next)
		seen[p] = true
		perimeter += 4

		for _, n := range p.neighbours() {
			if _, exists := r.plots[n]; exists {
				// Neighbours don't have a fence
				perimeter--

				// If we haven't seen it, check its neighbours out
				if _, ok := seen[n]; !ok {
					next[n] = true
				}
			}
		}
	}

	return perimeter
}

func (r *Region) sides() int {
	edges := make(map[Edge]bool)

	if len(r.plots) == 0 {
		return 0
	}

	var root Pos
	for k := range r.plots {
		root = k
		break
	}

	next := map[Pos]bool{root: true}
	seen := map[Pos]bool{}

	for len(next) > 0 {
		p := pop(next)
		seen[p] = true

		for _, n := range p.neighbours() {
			if _, exists := r.plots[n]; exists {
				// If we haven't seen it, check its neighbours out
				if _, ok := seen[n]; !ok {
					next[n] = true
				}
			} else {
				edges[Edge{p, p.direction(n)}] = true
			}
		}
	}

	var sides [][]Edge
	if *debug {
		fmt.Printf("edges: %v\n", edges)
	}
	for edge := range edges {
		if *debug {
			fmt.Printf("edge: %v\n", edge)
		}
		delete(edges, edge)

		directions := [2]Direction{}
		switch edge.d {
		case North, South:
			directions[0] = West
			directions[1] = East
		case East, West:
			directions[0] = North
			directions[1] = South
		}
		if *debug {
			fmt.Printf("directions: %v\n", directions)
		}
		side := []Edge{edge}
		for _, d := range directions {
			if *debug {
				fmt.Printf("trying direction %v, side %v\n", d, side)
				fmt.Printf("edges: %v, edge: %v\n", edges, edge)
			}
			e := edge
			for {
				e = Edge{p: e.p.move(d), d: e.d}
				if _, ok := edges[e]; !ok {
					if *debug {
						fmt.Printf("no edge at %v\n", e)
					}
					break
				}
				if *debug {
					fmt.Println("deleting edge")
				}
				delete(edges, e)
				side = append(side, e)
				if *debug {
					fmt.Printf("side is %v\n", side)
				}
			}
			if *debug {
				fmt.Printf("side is %v\n", side)
			}
		}
		if *debug {
			fmt.Println("appending side")
		}
		sides = append(sides, side)
		if *debug {
			fmt.Printf("sides: %v\n", sides)
		}
	}
	if *debug {
		fmt.Printf("sides: %v\n", sides)
	}
	return len(sides)
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

	farm := &Farm{
		plots: make(map[Pos]rune),
	}

	s := bufio.NewScanner(f)
	y := 0
	for s.Scan() {
		line := s.Text()
		farm.width = len(line)
		for x, p := range line {
			farm.plots[Pos{x, y}] = p
		}
		y++
	}
	farm.height = y

	if *debug {
		fmt.Printf("plots: %d\n", len(farm.plots))
		fmt.Printf("sides: %v\n", *sides)
	}
	regions := farm.regions()

	total := 0
	for _, region := range regions {
		area := len(region.plots)
		var price int
		fmt.Printf("region %c: ", region.plant)

		if *sides {
			sides := region.sides()
			price = area * sides
			fmt.Printf("price = %d * %d = %d\n", area, sides, price)
		} else {
			perimeter := region.perimeter()
			price = area * perimeter
			fmt.Printf("price = %d * %d = %d\n", area, perimeter, price)
		}
		total += price
	}
	fmt.Printf("total price: %d\n", total)
}
