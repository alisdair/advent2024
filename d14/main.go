package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	tm "github.com/buger/goterm"
)

var (
	gridsize   = flag.String("grid", "101x103", "width and height of the grid")
	iterations = flag.Int("iterations", 100, "iterations to simulate")
	search     = flag.Bool("search", false, "look for christmas tree")
)

func get[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

type Robot struct {
	pos Pos
	vel Vel
}

func (r *Robot) Step(n, xmax, ymax int) {
	r.pos.x = (r.pos.x + r.vel.x*n) % xmax
	for r.pos.x < 0 {
		r.pos.x += xmax
	}
	r.pos.y = (r.pos.y + r.vel.y*n) % ymax
	for r.pos.y < 0 {
		r.pos.y += ymax
	}
}

type Pos struct {
	x, y int
}

func ParsePos(s string) Pos {
	xy := strings.Split(s, ",")
	return Pos{
		x: get(strconv.Atoi(xy[0])),
		y: get(strconv.Atoi(xy[1])),
	}
}

type Vel struct {
	x, y int
}

func ParseVel(s string) Vel {
	xy := strings.Split(s, ",")
	return Vel{
		x: get(strconv.Atoi(xy[0])),
		y: get(strconv.Atoi(xy[1])),
	}
}

type Grid struct {
	width, height int
	robots        []*Robot
}

func NewGrid(size string, robots []*Robot) *Grid {
	xy := strings.Split(size, "x")
	return &Grid{
		width:  get(strconv.Atoi(xy[0])),
		height: get(strconv.Atoi(xy[1])),
		robots: robots,
	}
}

func (g *Grid) Draw() {
	g.draw(false)
}

func (g *Grid) DrawQuadrants() {
	g.draw(true)
}

func (g *Grid) draw(skipmid bool) {
	midx := g.width / 2
	midy := g.height / 2
	tm.MoveCursor(1, 2)
	tm.Clear()
	for y := 0; y < g.height; y++ {
		if skipmid && y == midy {
			tm.Println()
			continue
		}
		for x := 0; x < g.width; x++ {
			if skipmid && x == midx {
				tm.Print("  ")
				continue
			}
			robots := 0
			p := Pos{x, y}
			for _, robot := range g.robots {
				if robot.pos == p {
					robots++
				}
			}
			switch robots {
			case 0:
				tm.Print(" .")
			default:
				tm.Print(tm.Color(fmt.Sprintf("%2d", robots), tm.RED))
			}
		}
		tm.Println()
	}
	tm.Println(tm.Color("", tm.WHITE))
	tm.Flush()
}

func (g *Grid) Step(n int) {
	for _, robot := range g.robots {
		robot.Step(n, g.width, g.height)
	}
}

func (g *Grid) IsTree() int {
	// What is a tree? How about something where the top left and top right
	// quadrants have fewer than half the robots in the bottom left and bottom
	// right. Would that work?
	//
	// qs, _ := g.quadrants()
	// return qs[0] < qs[2]/2 && qs[1] < qs[3]/2
	//
	// Answer: lol no

	// Another attempt: most of the robots are in this part of the grid:
	//
	// .....*.....
	// ....***....
	// ...*****...
	// ..*******..
	// .*********.
	// ***********
	// .....*.....
	//
	// midx := g.width / 2
	// tree := make(map[Pos]bool, g.height*g.width)
	// for y := 0; y < g.height; y++ {
	// 	for x := 0; x < g.width; x++ {
	// 		l, r := midx-y, midx+y
	// 		if l < 0 {
	// 			l = midx
	// 		}
	// 		if r > g.width-1 {
	// 			r = midx
	// 		}
	// 		tree[Pos{x, y}] = x >= l && x <= r
	// 	}
	// }
	// match := 0
	// for _, r := range g.robots {
	// 	if tree[r.pos] {
	// 		match++
	// 	} else {
	// 		match--
	// 	}
	// }
	// if *debug {
	// 	tm.MoveCursor(30, 1)
	// 	tm.Printf("match: %5d   ", match)
	// 	tm.Flush()
	// }
	// return match > 0
	//
	// Welp nope

	// Okay, let's assume the tree is a small image somewhere in the grid. Then
	// all the robots should be close to each other.
	var dx, dy int
	for _, r0 := range g.robots {
		var r0dx, r0dy int
		for _, r1 := range g.robots {
			r1dx := r0.pos.x - r1.pos.x
			r0dx += r1dx * r1dx
			r1dy := r0.pos.y - r1.pos.y
			r0dy += r1dy * r1dy
		}
		dx += r0dx
		dy += r0dy

	}
	return dx + dy
}

func (g *Grid) SafetyFactor() {
	quadrants, middle := g.quadrants()
	fmt.Printf("quadrants: %v, middle: %d\n", quadrants, middle)
	product := 1
	for _, q := range quadrants {
		product *= q
	}
	fmt.Printf("safety factor: %d\n", product)
}

func (g *Grid) quadrants() ([4]int, int) {
	midx := g.width / 2
	midy := g.height / 2
	quadrants := [4]int{}
	middle := 0
	for _, r := range g.robots {
		switch {
		case r.pos.x < midx && r.pos.y < midy:
			quadrants[0]++
		case r.pos.x > midx && r.pos.y < midy:
			quadrants[1]++
		case r.pos.x < midx && r.pos.y > midy:
			quadrants[2]++
		case r.pos.x > midx && r.pos.y > midy:
			quadrants[3]++
		default:
			middle++
		}
	}
	return quadrants, middle
}

func main() {
	flag.Parse()
	filename := "example.txt"
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
	}

	f := get(os.Open(filename))

	var robots []*Robot

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		robot := &Robot{}
		for i, column := range strings.Split(line, " ") {
			kv := strings.Split(column, "=")
			k, v := kv[0], kv[1]
			switch k {
			case "p":
				robot.pos = ParsePos(v)
			case "v":
				robot.vel = ParseVel(v)
			default:
				panic(fmt.Sprintf("line %d: %s", i, k))
			}
		}
		robots = append(robots, robot)
	}

	grid := NewGrid(*gridsize, robots)

	var i int
	ii := 0
	minDelta := math.MaxInt
	if *search {
		for i = 0; i < *iterations; i++ {
			grid.Step(1)
			delta := grid.IsTree()
			if delta < minDelta {
				minDelta = delta
				ii = i
			}
		}
		grid.Draw()
		fmt.Printf("delta %d, ii %d\n", minDelta, ii)
	} else {
		grid.Step(*iterations)
		grid.Draw()
		grid.SafetyFactor()
	}
}
