package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
)

var (
	blinks = flag.Int("blinks", 75, "blink at the stones this many times")
)

func get[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func width(n int) int {
	if n == 0 {
		return 1
	}
	return int(math.Floor(math.Log10(float64(n)) + 1))
}

func halve(n, width int) (int, int) {
	if width%2 != 0 {
		panic("can't halve")
	}
	split := 10
	for width > 2 {
		split *= 10
		width -= 2
	}
	b := n % split
	a := (n - b) / split
	return a, b
}

func blink(stones map[int]int) map[int]int {
	ret := make(map[int]int, len(stones)*2)

	for stone, count := range stones {
		width := width(stone)
		switch {
		case stone == 0:
			ret[1] += count
		case width%2 == 0:
			a, b := halve(stone, width)
			ret[a] += count
			ret[b] += count
		default:
			ret[stone*2024] += count
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

	// stone => count
	stones := make(map[int]int)

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanWords)
	for s.Scan() {
		stone := get(strconv.Atoi(s.Text()))
		stones[stone] = 1
	}

	for range *blinks {
		stones = blink(stones)
	}

	count := 0
	for _, v := range stones {
		count += v
	}

	fmt.Printf("count: %d\n", count)
}
