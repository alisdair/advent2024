package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

var corrected = flag.Bool("corrected", true, "corrected prize location")

func get[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

type Machine struct {
	a, b  Button
	prize Pos
}

func (m *Machine) String() string {
	return fmt.Sprintf("A=%s, B=%s, P=%s", m.a, m.b, m.prize)
}

func (m *Machine) solution() (solution Solution, ok bool) {
	solution = Solution{}

	// See README for the derivation of these equations
	cn := float64(m.prize.x*m.b.y - m.b.x*m.prize.y)
	cd := float64(m.a.x*m.b.y - m.b.x*m.a.y)
	c := cn / cd
	dn := float64(m.prize.x*m.a.y - m.a.x*m.prize.y)
	dd := float64(m.b.x*m.a.y - m.a.x*m.b.y)
	d := dn / dd

	// Solution is valid if these are integer results
	if c == math.Trunc(c) && d == math.Trunc(d) {
		solution.a = int(c)
		solution.b = int(d)
		solution.cost = solution.a*m.a.cost + solution.b*m.b.cost
		return solution, true
	}
	return solution, false
}

type Solution struct {
	a, b, cost int
}

func (s Solution) String() string {
	return fmt.Sprintf("A: %d, B: %d, Cost: %d", s.a, s.b, s.cost)
}

type Button struct {
	x, y int
	cost int
}

func (b Button) Apply(p *Pos, n int) {
	p.x += b.x * n
	p.y += b.y * n
}

func (b Button) String() string {
	return fmt.Sprintf("(%+d,%+d) $%d", b.x, b.y, b.cost)
}

func ParseButton(s string) Button {
	ret := Button{}
	coords := strings.Split(s, ", ")
	for _, coord := range coords {
		switch {
		case strings.HasPrefix(coord, "X"):
			ret.x = get(strconv.Atoi(coord[1:]))
		case strings.HasPrefix(coord, "Y"):
			ret.y = get(strconv.Atoi(coord[1:]))
		default:
			panic(coord)
		}
	}
	return ret
}

type Pos struct {
	x, y int
}

func (p Pos) String() string {
	return fmt.Sprintf("(%d, %d)", p.x, p.y)
}

func ParsePos(s string) Pos {
	ret := Pos{}
	coords := strings.Split(s, ", ")
	for _, coord := range coords {
		switch {
		case strings.HasPrefix(coord, "X="):
			ret.x = get(strconv.Atoi(coord[2:]))
		case strings.HasPrefix(coord, "Y="):
			ret.y = get(strconv.Atoi(coord[2:]))
		default:
			panic(coord)
		}
	}
	return ret
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

	var machines []*Machine
	machine := &Machine{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if line == "" {
			machines = append(machines, machine)
			machine = &Machine{}
			continue
		}
		fields := strings.Split(line, ": ")
		switch fields[0] {
		case "Button A":
			machine.a = ParseButton(fields[1])
			machine.a.cost = 3
		case "Button B":
			machine.b = ParseButton(fields[1])
			machine.b.cost = 1
		case "Prize":
			machine.prize = ParsePos(fields[1])
			if *corrected {
				machine.prize.x += 10000000000000
				machine.prize.y += 10000000000000
			}
		default:
			panic(line)
		}
	}
	if machine.prize.x != 0 && machine.prize.y != 0 {
		machines = append(machines, machine)
	}

	total := 0
	for _, machine := range machines {
		fmt.Printf("%v\n", machine)
		if solution, ok := machine.solution(); ok {
			fmt.Printf("%v", solution)
			total += solution.cost
		} else {
			fmt.Printf("no solution")
		}
		fmt.Printf("\n\n")
	}
	fmt.Printf("total: %d\n", total)
}
