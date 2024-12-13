package main

import "testing"

func TestRegion_perimeter(t *testing.T) {
	testcases := []struct {
		name  string
		plots []Pos
		want  int
	}{
		{"empty", []Pos{}, 0},
		{"unit", []Pos{{0, 0}}, 4},
		{"two", []Pos{{0, 0}, {1, 0}}, 6},
		{"line", []Pos{{0, 0}, {1, 0}, {2, 0}, {3, 0}}, 10},
		{"vert", []Pos{{0, 0}, {0, 1}, {0, 2}}, 8},
		{"L", []Pos{{0, 0}, {0, 1}, {0, 2}, {1, 2}}, 10},
		{"O", []Pos{{0, 0}, {0, 1}, {0, 2}, {1, 2}, {2, 2}, {2, 1}, {2, 0}, {1, 0}}, 16},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			plots := make(map[Pos]bool, len(tc.plots))
			for _, p := range tc.plots {
				plots[p] = true
			}

			r := Region{plots: plots}

			if got, want := r.perimeter(), tc.want; got != want {
				t.Errorf("wrong result. got = %d, want = %d", got, want)
			}
		})
	}
}

func TestRegion_sides(t *testing.T) {
	testcases := []struct {
		name  string
		plots []Pos
		want  int
	}{
		{"empty", []Pos{}, 0},
		{"unit", []Pos{{0, 0}}, 4},
		{"two", []Pos{{0, 0}, {1, 0}}, 4},
		{"line", []Pos{{0, 0}, {1, 0}, {2, 0}, {3, 0}}, 4},
		{"vert", []Pos{{0, 0}, {0, 1}, {0, 2}}, 4},
		{"L", []Pos{{0, 0}, {0, 1}, {0, 2}, {1, 2}}, 6},
		{"S", []Pos{{0, 0}, {1, 0}, {1, 1}, {2, 1}}, 8},
		{"O", []Pos{{0, 0}, {0, 1}, {0, 2}, {1, 2}, {2, 2}, {2, 1}, {2, 0}, {1, 0}}, 8},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			*debug = true
			t.Logf("\n\n%s\n", tc.name)
			plots := make(map[Pos]bool, len(tc.plots))
			for _, p := range tc.plots {
				plots[p] = true
			}

			r := Region{plots: plots}

			if got, want := r.sides(), tc.want; got != want {
				t.Errorf("wrong result. got = %d, want = %d", got, want)
			}
		})
	}
}
