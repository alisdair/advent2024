package main

import (
	"bufio"
	"flag"
	"fmt"
	"iter"
	"os"
	"time"

	tm "github.com/buger/goterm"
)

var (
	debug = flag.Bool("debug", false, "extra logs please")
	fps   = flag.Int("fps", 60, "frames per second when rendering")
)

type Set[K comparable] map[K]struct{}

func NewSet[K comparable]() Set[K] {
	return make(Set[K])
}

func (s Set[K]) Add(k K) {
	s[k] = struct{}{}
}

func (s Set[K]) Has(k K) bool {
	_, ok := s[k]
	return ok
}

func (s Set[K]) Remove(k K) {
	delete(s, k)
}

func (s Set[K]) Replace(k0, k1 K) {
	delete(s, k0)
	s[k1] = struct{}{}
}

type Warehouse struct {
	floor         Set[Pos]
	walls         Set[Pos]
	boxes         Set[Pos]
	robot         Pos
	width, height int
}

func NewWarehouse() *Warehouse {
	return &Warehouse{
		floor:  NewSet[Pos](),
		walls:  NewSet[Pos](),
		boxes:  NewSet[Pos](),
		width:  -1,
		height: -1,
	}
}

func (wh *Warehouse) Each() iter.Seq[Pos] {
	return func(yield func(Pos) bool) {
		for y := 0; y < wh.height; y++ {
			for x := 0; x < wh.width; x++ {
				if !yield(Pos{x, y}) {
					return
				}
			}
		}
	}
}

func (wh *Warehouse) Draw() {
	for p := range wh.Each() {
		tm.MoveCursor(p.x*2+1, p.y+1)
		switch {
		case wh.robot == p:
			tm.Print(tm.Bold(tm.Color("@", tm.RED)))
		case wh.walls.Has(p):
			tm.Print(tm.Color("#", tm.BLACK))
		case wh.boxes.Has(p):
			tm.Print(tm.Bold(tm.Color("O", tm.YELLOW)))
		case wh.floor.Has(p):
			tm.Print(tm.Bold(tm.Color(".", tm.WHITE)))
		default:
			panic(fmt.Sprintf("no object at %s", p))
		}
	}
	tm.Println()
	tm.Flush()
}

func (wh *Warehouse) MoveRobot(m Move) {
	wh.robot = wh.moveObject(wh.robot, m)
}

func (wh *Warehouse) moveObject(cur Pos, m Move) Pos {
	p := cur.Move(m)
	switch {
	case wh.walls.Has(p):
		return cur
	case wh.boxes.Has(p):
		if pp := wh.moveObject(p, m); pp != p {
			wh.boxes.Replace(p, pp)
			return p
		}
	case wh.floor.Has(p):
		return p
	default:
		panic(fmt.Sprintf("no object at %s", p))
	}
	return cur
}

func (wh *Warehouse) SumBoxes() int {
	var sum int
	for p := range wh.boxes {
		sum += p.x + 100*p.y
	}
	return sum
}

type Pos struct {
	x, y int
}

func (p Pos) String() string {
	return fmt.Sprintf("(%d, %d)", p.x, p.y)
}

func (p Pos) Move(d Move) Pos {
	switch d {
	case Up:
		return Pos{p.x, p.y - 1}
	case Right:
		return Pos{p.x + 1, p.y}
	case Down:
		return Pos{p.x, p.y + 1}
	case Left:
		return Pos{p.x - 1, p.y}
	default:
		panic("invalid direction")
	}
}

type Move rune

func NewMove(r rune) Move {
	switch r {
	case rune(Up):
		return Up
	case rune(Right):
		return Right
	case rune(Down):
		return Down
	case rune(Left):
		return Left
	default:
		panic(fmt.Sprintf("invalid move %q", r))
	}
}

func (d Move) String() string {
	return fmt.Sprintf("%c", d)
}

const (
	Up    Move = '^'
	Right Move = '>'
	Down  Move = 'v'
	Left  Move = '<'
)

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

	wh := NewWarehouse()
	var moves []Move

	s := bufio.NewScanner(f)
	y := 0
	readingLayout := true
	for s.Scan() {
		line := s.Text()
		if readingLayout {
			if len(line) == 0 {
				readingLayout = false
				continue
			} else if wh.width == -1 {
				wh.width = len(line)
			}
			for x, c := range line {
				p := Pos{x, y}
				switch c {
				case '.':
					wh.floor.Add(p)
				case '#':
					wh.walls.Add(p)
				case 'O':
					wh.floor.Add(p)
					wh.boxes.Add(p)
				case '@':
					wh.floor.Add(p)
					wh.robot = p
				default:
					panic(fmt.Sprintf("%s: %c", p, c))
				}
			}
			y++
		} else {
			for _, m := range line {
				moves = append(moves, NewMove(m))
			}
		}
	}
	wh.height = y

	if *debug {
		tm.Clear()
		wh.Draw()
	}

	frameDelay := time.Second / time.Duration(*fps)
	for _, m := range moves {
		t := time.Now()
		wh.MoveRobot(m)
		if *debug {
			wh.Draw()
			time.Sleep(frameDelay - time.Since(t))
		}
	}

	wh.Draw()
	fmt.Printf("\n\nSum of boxes' GPS coordinates: %d\n", wh.SumBoxes())
}
