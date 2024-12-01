package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
)

func get[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func distance(as, bs []int) int {
	if len(as) != len(bs) {
		log.Fatalf("inequal lengths %d and %d", len(as), len(bs))
	}

	ret := 0
	for i := range as {
		d := as[i] - bs[i]
		if d < 0 {
			d = d * -1
		}
		ret += d
	}

	return ret
}

func similarity(as, bs []int) int {
	if len(as) != len(bs) {
		log.Fatalf("inequal lengths %d and %d", len(as), len(bs))
	}

	nbs := make(map[int]int)

	for _, b := range bs {
		if _, ok := nbs[b]; !ok {
			nbs[b] = 0
		}
		nbs[b]++
	}

	ret := 0
	for _, a := range as {
		if nb, ok := nbs[a]; ok {
			ret += a * nb
		}
	}

	return ret
}

func main() {
	if len(os.Args) != 2 {
		os.Args = []string{"d1", "example.txt"}
		// log.Fatal("Usage: d1 [filepath]")
	}

	f := get(os.Open(os.Args[1]))
	defer f.Close()

	var as, bs []int

	re := regexp.MustCompile("[ \t]+")
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		columns := re.Split(s.Text(), -1)
		if len(columns) != 2 {
			log.Fatalf("invalid line (expected 2 columns, got %d)\n%q", len(columns), line)
		}
		as = append(as, get(strconv.Atoi(columns[0])))
		bs = append(bs, get(strconv.Atoi(columns[1])))
	}

	sort.Ints(as)
	sort.Ints(bs)

	fmt.Printf("distance: %d\n", distance(as, bs))
	fmt.Printf("similarity: %d\n", similarity(as, bs))
}
