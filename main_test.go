package main

import (
	"math/rand/v2"
	"sync/atomic"
	"testing"
)

var counter atomic.Int64

func randomInt(min, max int) int {
	return rand.IntN(max-min) + min
}

func TestApp(t *testing.T) {
}

