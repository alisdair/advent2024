package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

func safe(report []int) bool {
	if len(report) < 2 {
		return false
	}

	prev := report[0]
	asc := false
	desc := false

	for _, cur := range report[1:] {
		if cur > prev {
			if desc {
				return false
			}
			if cur-prev > 3 {
				return false
			}
			asc = true
			prev = cur
			continue
		}
		if cur < prev {
			if asc {
				return false
			}
			if prev-cur > 3 {
				return false
			}
			desc = true
			prev = cur
			continue
		}
		return false
	}
	return true
}

func get[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: d2 [filepath]")
	}

	f := get(os.Open(os.Args[1]))
	defer f.Close()

	var reports [][]int

	re := regexp.MustCompile("[ \t]+")
	s := bufio.NewScanner(f)
	for s.Scan() {
		columns := re.Split(s.Text(), -1)
		report := make([]int, len(columns))
		for i := range columns {
			report[i] = get(strconv.Atoi(columns[i]))
		}
		reports = append(reports, report)
	}

	total := 0
	dampened := 0
	for _, report := range reports {
		if safe(report) {
			total++
		} else {
			tolerated := make([]int, len(report)-1)
			for r := range report {
				for i := range report {
					if i == r {
						continue
					} else if i < r {
						tolerated[i] = report[i]
					} else if i > r {
						tolerated[i-1] = report[i]
					}
				}
				if safe(tolerated) {
					dampened++
					break
				}
			}
		}
	}

	fmt.Printf("total: %d\n", total)
	fmt.Printf("total dampened: %d\n", total+dampened)
}
