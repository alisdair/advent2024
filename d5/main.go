package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
)

var debug = flag.Bool("debug", false, "debug mode")

func get[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

type invalid struct {
	i, j  int // indexes of conflict found
	rule  *rule
	cause *rule
}

func (i *invalid) String() string {
	return fmt.Sprintf("page %d must be before page %d", i.cause.page, i.rule.page)
}

type rule struct {
	page  int
	after map[int]*rule // page -> rule
	// TODO: remove after if not used
	before map[int]*rule // page -> rule
}

func (r *rule) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d->(", r.page))
	var afters []string
	for a := range r.after {
		afters = append(afters, fmt.Sprintf("%d", a))
	}
	b.WriteString(strings.Join(afters, ","))
	b.WriteRune(')')
	return b.String()
}

func (r *rule) addAfter(ra *rule) {
	if _, ok := r.after[ra.page]; ok {
		log.Fatalf("redundant rule %d|%d", r.page, ra.page)
	} else {
		r.after[ra.page] = ra
	}
}

func (r *rule) addBefore(rb *rule) {
	if _, ok := r.before[rb.page]; ok {
		log.Fatalf("redundant rule %d|%d", rb.page, r.page)
	} else {
		r.before[rb.page] = rb
	}
}

type rules struct {
	rs map[int]*rule
}

func (rs *rules) get(page int) *rule {
	r, ok := rs.rs[page]
	if !ok {
		r = &rule{
			page:   page,
			after:  make(map[int]*rule),
			before: make(map[int]*rule),
		}
		rs.rs[page] = r
	}
	return r
}

func (rs *rules) valid(u update) bool {
	if len(u) == 0 {
		return false
	}
	return rs.findInvalid(u) == nil
}

func (rs *rules) findInvalid(u update) *invalid {
	for i, p := range u {
		r, ok := rs.rs[p]
		if !ok {
			continue
		}
		for j, pp := range u[i:] {
			if _, ok := r.before[pp]; ok {
				return &invalid{
					i:     i,
					j:     j + i,
					rule:  r,
					cause: r.before[pp],
				}
			}
		}
	}
	return nil
}

func (rs *rules) autocorrect(u update) update {
	// Sort the update pages into most-constrained first order
	mcf := make(update, len(u))
	copy(mcf, u)
	slices.SortFunc(mcf, func(a, b int) int {
		ra, aok := rs.rs[a]
		rb, bok := rs.rs[b]

		if !aok && !bok {
			return 0
		}
		if !aok {
			return 1
		}
		if !bok {
			return -1
		}

		return len(rb.after) - len(ra.after)
	})

	ret := make(update, len(mcf))
	copy(ret, mcf)

	for !rs.valid(ret) {
		invalid := rs.findInvalid(ret)
		if *debug {
			fmt.Printf("invalid: %s\n", ret)
			fmt.Printf("invalid at %d and %d: %s\n", invalid.i, invalid.j, invalid)
		}
		ret[invalid.i], ret[invalid.j] = ret[invalid.j], ret[invalid.i]
	}

	return ret
}

type update []int

func (u update) String() string {
	var b strings.Builder
	var pages []string
	for _, p := range u {
		pages = append(pages, fmt.Sprintf("%d", p))
	}
	b.WriteString(strings.Join(pages, ","))
	return b.String()
}

func (u update) middle() int {
	return u[len(u)/2]
}

func main() {
	flag.Parse()

	filename := "example.txt"
	if args := flag.Args(); len(args) > 0 {
		filename = args[0]
	}

	f := get(os.Open(filename))
	defer f.Close()

	rules := &rules{
		rs: make(map[int]*rule),
	}
	updates := make([]update, 0)

	section := "rules"
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()

		switch section {
		case "rules":
			if line == "" {
				section = "updates"
				continue
			}

			pages := strings.Split(line, "|")
			if len(pages) != 2 {
				log.Fatalf("invalid rule %q", line)
			}

			before := get(strconv.Atoi(pages[0]))
			after := get(strconv.Atoi(pages[1]))

			rb := rules.get(before)
			ra := rules.get(after)

			rb.addAfter(ra)
			ra.addBefore(rb)

		case "updates":
			pages := strings.Split(line, ",")
			var u update
			for _, page := range pages {
				u = append(u, get(strconv.Atoi(page)))
			}
			updates = append(updates, u)
		}
	}

	if *debug {
		fmt.Printf("rules:\n")
		for _, r := range rules.rs {
			fmt.Println(r)
		}
	}

	if *debug {
		fmt.Printf("\nupdates:\n")
	}

	var invalids []update
	validMiddles := 0
	for _, u := range updates {
		valid := rules.valid(u)
		if *debug {
			fmt.Printf("%s: %v\n", u, valid)
		}
		if valid {
			validMiddles += u.middle()
		} else {
			invalids = append(invalids, u)
		}
	}

	if *debug {
		fmt.Printf("\ncorrected:\n")
	}

	correctedMiddles := 0
	for _, u := range invalids {
		corrected := rules.autocorrect(u)
		if *debug {
			fmt.Printf("%s: %v\n", corrected, rules.valid(corrected))
		}
		if !rules.valid(corrected) {
			invalid := rules.findInvalid(corrected)
			fmt.Printf("invalid after autocorrect because %s\n", invalid)
			os.Exit(1)
		}
		correctedMiddles += corrected.middle()
	}

	fmt.Printf("\nvalid middles: %d\n", validMiddles)
	fmt.Printf("\ncorrected middles: %d\n", correctedMiddles)
}
