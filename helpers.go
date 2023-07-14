package main

import (
	"log"
	"math/rand"
	"time"
)

// getRandom returns a positive random number for a given range.
func getRandom(rangeInt int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	randomNumber := rand.Intn(rangeInt)
	return randomNumber
}
