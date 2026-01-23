package lib

import (
	"math/rand"
)

/* ---------------- Motifs ---------------- */

func RandRange(rng *rand.Rand, a, b float64) float64 {
	return a + rng.Float64()*(b-a)
}

func ChooseRandom[T any](r *rand.Rand, items []T) T {
	return items[r.Intn(len(items))]
}
