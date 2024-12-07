package main

import "iter"

func combinations[T ~[]E, E any](options T, length int) iter.Seq[T] {
	return func(yield func(T) bool) {
		// state is a slice of digits of base `length`, starting at 0
		state := make([]int, length)

		// total = pow(len(options), length)
		total := len(options)
		for i := 0; i < length-1; i++ {
			total *= len(options)
		}

		// v is the slice yielded on each iteration
		v := make([]E, length)
		for i := 0; i < total; i++ {
			// convert the slice of indices to a slice of values
			for j := range state {
				v[j] = options[state[j]]
			}

			if !yield(v) {
				return
			}

			// increment state, with carry between digits
			carry := false
			for j := len(state) - 1; j >= 0; j-- {
				if state[j] < len(options)-1 {
					// we can increment this digit, so we're done
					state[j]++
					carry = false
					break
				} else {
					// this digit has overflowed, so reset it and mark carry
					state[j] = 0
					carry = true
				}
			}
			// if the final digit has carry, we've completed the iterations
			if carry {
				return
			}
		}
	}
}
