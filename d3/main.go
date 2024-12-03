package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

func get[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: d3 [filepath]")
	}

	f := get(os.ReadFile(os.Args[1]))

	var total int
	var results []int

	mul := regexp.MustCompile(`^mul\((\d+),(\d+)\)`)
	start := regexp.MustCompile((`^do()`))
	stop := regexp.MustCompile((`^don't()`))

	enabled := true
	for len(f) > 0 {
		if mul.Match(f) {
			if enabled {
				ms := mul.FindSubmatch(f)
				x := get(strconv.Atoi(string(ms[1])))
				y := get(strconv.Atoi(string(ms[2])))
				z := x * y
				total += z
				results = append(results, z)
			}
		}
		if start.Match(f) {
			enabled = true
		}
		if stop.Match(f) {
			enabled = false
		}
		f = f[1:]
	}

	fmt.Printf("total: %d\n", total)
}
