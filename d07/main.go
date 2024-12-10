package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var debug = flag.Bool("debug", false, "show equations")

func get[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

type Operator rune

const (
	unknown Operator = '?'
	plus    Operator = '+'
	times   Operator = '*'
	concat  Operator = '|'
)

type Equation struct {
	total     int
	operands  []int
	operators []Operator
}

func NewEquation(total int, operands []int) *Equation {
	if len(operands) == 0 {
		log.Fatalf("invalid equation: empty operands")
	}
	operators := make([]Operator, len(operands)-1)
	for i := range operators {
		operators[i] = unknown
	}
	return &Equation{
		total:     total,
		operands:  operands,
		operators: operators,
	}
}

func (e *Equation) String() string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(e.total))
	b.WriteString(" = ")
	for i, operand := range e.operands {
		b.WriteString(strconv.Itoa(operand))
		if i < len(e.operators) {
			b.WriteString(fmt.Sprintf(" %c ", e.operators[i]))
		}
	}
	return b.String()
}

func (e *Equation) Evaluate() (int, bool) {
	total := e.operands[0]
	for i, op := range e.operators {
		switch op {
		case plus:
			total = total + e.operands[i+1]
		case times:
			total = total * e.operands[i+1]
		case concat:
			total = concatenate(total, e.operands[i+1])
		default:
			return 0, false
		}
	}
	return total, true
}

func concatenate(x, y int) int {
	z := 10
	for y >= z {
		z *= 10
	}
	return x*z + y
}

func (e *Equation) Solve(operators []Operator) {
	for combination := range combinations(operators, len(e.operators)) {
		e.operators = combination
		if total, ok := e.Evaluate(); ok && total == e.total {
			return
		}
	}
}

func (e *Equation) Valid() bool {
	result, ok := e.Evaluate()
	return ok && result == e.total
}

func main() {
	flag.Parse()

	filename := "example.txt"
	if args := flag.Args(); len(args) > 0 {
		filename = args[0]
	}

	f := get(os.Open(filename))
	defer f.Close()

	var equations []*Equation
	sep := regexp.MustCompile("[: ]+")
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		columns := sep.Split(line, -1)
		total := get(strconv.Atoi(columns[0]))
		var operands []int
		for _, operand := range columns[1:] {
			operands = append(operands, get(strconv.Atoi(operand)))
		}
		equations = append(equations, NewEquation(total, operands))
	}

	total := 0
	for _, equation := range equations {
		equation.Solve([]Operator{'+', '*'})
		if equation.Valid() {
			if *debug {
				fmt.Printf("%s: solved\n", equation)
			}
			total += equation.total
		} else {
			if *debug {
				fmt.Printf("%s\n", equation)
			}
		}
	}
	fmt.Printf("two-operator total: %d\n\n", total)

	for _, equation := range equations {
		if equation.Valid() {
			continue
		}
		equation.Solve([]Operator{'+', '*', '|'})
		if equation.Valid() {
			if *debug {
				fmt.Printf("%s: solved\n", equation)
			}
			total += equation.total
		}
	}
	fmt.Printf("three-operator total: %d\n", total)
}
