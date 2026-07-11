package main

import (
	"fmt"
	"math/rand/v2"
)

func choose[T any](items []T) (T, error) {
	if len(items) == 0 {
		var zeroValue T
		return zeroValue, fmt.Errorf("Cannot choose from empty slice")
	}

	randomIndex := rand.IntN(len(items))
	return items[randomIndex], nil
}
