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
	wide  = flag.Bool("wide", false, "boxes are double wide (part 2)")
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
	lboxes        Set[Pos]
	rboxes        Set[Pos]
	robot         Pos
	width, height int
}

func NewWarehouse() *Warehouse {
	return &Warehouse{
		floor:  NewSet[Pos](),
		walls:  NewSet[Pos](),
		boxes:  NewSet[Pos](),
		lboxes: NewSet[Pos](),
		rboxes: NewSet[Pos](),
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

func (wh *Warehouse) Draw(m Move) {
	for p := range wh.Each() {
		if *wide {
			tm.MoveCursor(p.x+1, p.y+1)
		} else {
			tm.MoveCursor(p.x*2+1, p.y+1)
		}
		switch {
		case wh.robot == p:
			tm.Print(tm.Bold(tm.Color("@", tm.RED)))
		case wh.walls.Has(p):
			tm.Print(tm.Color("#", tm.BLUE))
		case wh.boxes.Has(p):
			tm.Print(tm.Bold(tm.Color("O", tm.YELLOW)))
		case wh.lboxes.Has(p):
			tm.Print(tm.Bold(tm.Color("[", tm.YELLOW)))
		case wh.rboxes.Has(p):
			tm.Print(tm.Bold(tm.Color("]", tm.YELLOW)))
		case wh.floor.Has(p):
			tm.Print(tm.Bold(tm.Color(".", tm.WHITE)))
		default:
			panic(fmt.Sprintf("no object at %s", p))
		}
	}
	tm.Printf("\nMove: %c\n", m)
	tm.Flush()
}

func (wh *Warehouse) MoveRobot(m Move) {
	if plan := wh.planMove(wh.robot, m, nil); plan != nil {
		wh.applyPlan(m, plan)
	}
}

type PlannedMove struct {
	from, to Pos
}

func (pm PlannedMove) String() string {
	return fmt.Sprintf("%s-%s", pm.from, pm.to)
}

func (wh *Warehouse) applyPlan(m Move, pms []PlannedMove) {
	for _, pm := range pms {
		switch {
		case wh.boxes.Has(pm.from):
			wh.boxes.Replace(pm.from, pm.to)
		case wh.lboxes.Has(pm.from):
			wh.lboxes.Replace(pm.from, pm.to)
		case wh.rboxes.Has(pm.from):
			wh.rboxes.Replace(pm.from, pm.to)
		case wh.robot == pm.from:
			wh.robot = pm.to
		default:
			panic(fmt.Sprintf("invalid planned move %v\n%#v", pm, wh))
		}
	}
}

func merge[T comparable](a, b []T) []T {
	ret := make([]T, 0, len(a)+len(b))
	ret = append(ret, a...)
	for _, x := range b {
		// I am tired please do not judge me
		var seen bool
		for _, y := range ret {
			if x == y {
				seen = true
				break
			}
		}
		if !seen {
			ret = append(ret, x)
		}
	}
	return ret
}

func (wh *Warehouse) planMove(from Pos, m Move, pms []PlannedMove) []PlannedMove {
	to := from.Move(m)
	lr := m == Left || m == Right
	switch {
	case wh.walls.Has(to):
		return nil
	case wh.boxes.Has(to), lr && (wh.lboxes.Has(to) || wh.rboxes.Has(to)):
		pms = wh.planMove(to, m, pms)
		if len(pms) > 0 {
			return append(pms, PlannedMove{from, to})
		} else {
			return pms
		}
	case wh.lboxes.Has(to):
		pmsl, pmsr := wh.planMove(to, m, nil), wh.planMove(to.Move(Right), m, nil)
		if len(pmsl) > 0 && len(pmsr) > 0 {
			pms = append(pms, merge(pmsl, pmsr)...)
			pms = append(pms, PlannedMove{from, to})
		}
		return pms
	case wh.rboxes.Has(to):
		pmsl, pmsr := wh.planMove(to.Move(Left), m, nil), wh.planMove(to, m, nil)
		if len(pmsl) > 0 && len(pmsr) > 0 {
			pms = append(pms, merge(pmsl, pmsr)...)
			pms = append(pms, PlannedMove{from, to})
		}
		return pms
	case wh.floor.Has(to):
		return append(pms, PlannedMove{from, to})
	default:
		panic(fmt.Sprintf("no object at %s", to))
	}
}

func (wh *Warehouse) SumBoxes() int {
	var sum int
	for p := range wh.boxes {
		sum += p.x + 100*p.y
	}
	for p := range wh.lboxes {
		lx := p.x
		rx := p.Move(Right).x
		var x int
		if lx < rx {
			x = lx
		} else {
			x = rx
		}
		sum += x + 100*p.y
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
	Stop  Move = ' '
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
				if *wide {
					wh.width = len(line) * 2
				} else {
					wh.width = len(line)
				}
			}
			for x, c := range line {
				p := Pos{x, y}
				if *wide {
					p = Pos{x * 2, y}
				}
				switch c {
				case '.':
					wh.floor.Add(p)
					if *wide {
						wh.floor.Add(p.Move(Right))
					}
				case '#':
					wh.walls.Add(p)
					if *wide {
						wh.walls.Add(p.Move(Right))
					}
				case 'O':
					wh.floor.Add(p)
					if *wide {
						wh.floor.Add(p.Move(Right))
						wh.lboxes.Add(p)
						wh.rboxes.Add(p.Move(Right))
					} else {
						wh.boxes.Add(p)
					}
				case '@':
					wh.floor.Add(p)
					if *wide {
						wh.floor.Add(p.Move(Right))
					}
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
		wh.Draw(Stop)
	}

	frameDelay := time.Second / time.Duration(*fps)
	for _, m := range moves {
		t := time.Now()
		wh.MoveRobot(m)
		if *debug {
			wh.Draw(m)
			time.Sleep(frameDelay - time.Since(t))
		}
	}

	wh.Draw(Stop)
	fmt.Printf("\n\nSum of boxes' GPS coordinates: %d\n", wh.SumBoxes())
}
