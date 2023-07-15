package main

import (
	"math/rand"
)

// getRandom returns a positive random number for a given range.
func getRandom(rangeInt int) int {
	randomNumber := rand.Intn(rangeInt)
	return randomNumber
}
