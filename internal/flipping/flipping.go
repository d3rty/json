package flipping

import (
	"math/rand"
	"time"
)

type Coin struct {
	rng *rand.Rand
}

func MaybeNewCoin(coinArg ...*Coin) *Coin {
	if len(coinArg) > 0 {
		return coinArg[0]
	}

	return NewCoin()
}

func NewCoin(rngArg ...*rand.Rand) *Coin {
	// Use provided RNG or fallback to default.
	var rng *rand.Rand
	if len(rngArg) > 0 && rngArg[0] != nil {
		rng = rngArg[0]
	} else {
		//nolint: gosec // wer're find with simple `math/rand` here
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return &Coin{rng}
}

func (c *Coin) Flip() bool {
	return c.rng.Intn(2) == 1
}

func (c *Coin) Chance(threshold float64) bool {
	return c.rng.Float64() >= threshold
}

func (c *Coin) Rng() *rand.Rand { return c.rng }

// FeelingLucky returns random item from the given slice.
func FeelingLucky[T comparable](s []T, coinArg ...*Coin) T {
	if len(s) == 0 {
		var zero T
		return zero
	}

	var coin *Coin
	if len(coinArg) > 0 {
		coin = coinArg[0]
	} else {
		coin = NewCoin()
	}

	idx := coin.rng.Intn(len(s))
	return s[idx]
}
