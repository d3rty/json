package flipping_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/d3rty/json/internal/flipping"
	"github.com/stretchr/testify/assert"
)

func TestNewCoin(t *testing.T) {
	coin := flipping.NewCoin()
	assert.NotNil(t, coin, "NewCoin should not return nil")
}

func TestMaybeNewCoin(t *testing.T) {
	// When a coin is provided, MaybeNewCoin should return it.
	coin1 := flipping.NewCoin()
	coin2 := flipping.MaybeNewCoin(coin1)
	assert.Equal(t, coin1, coin2, "MaybeNewCoin should return the provided coin")

	// When no coin is provided, it should create a new one.
	coin3 := flipping.MaybeNewCoin()
	assert.NotNil(t, coin3, "MaybeNewCoin should create a new coin when no argument is provided")
}

func TestFlip(t *testing.T) {
	// Use a fixed seed for reproducibility.
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	coin := flipping.NewCoin(rng)

	// Call Flip multiple times and check that we eventually see both true and false.
	foundTrue := false
	foundFalse := false
	for range 100 {
		if coin.Flip() {
			foundTrue = true
		} else {
			foundFalse = true
		}
		if foundTrue && foundFalse {
			break
		}
	}
	assert.True(t, foundTrue, "Flip should eventually return true")
	assert.True(t, foundFalse, "Flip should eventually return false")
}

func TestChance(t *testing.T) {
	// Use a fixed seed for reproducibility.
	seed := int64(42)
	rng := rand.New(rand.NewSource(seed))
	coin := flipping.NewCoin(rng)

	// With threshold 0, Chance should always return true.
	for range 100 {
		assert.True(t, coin.Chance(0), "Chance with threshold 0 should always return true")
	}
	// With threshold 1, Chance should always return false because rng.Float64() is in [0,1).
	for range 100 {
		assert.False(t, coin.Chance(1), "Chance with threshold 1 should always return false")
	}
}

func TestFeelingLucky(t *testing.T) {
	// Test with a non-empty slice.
	nums := []int{1, 2, 3, 4, 5}
	seed := int64(42)
	rng := rand.New(rand.NewSource(seed))
	coin := flipping.NewCoin(rng)

	result := flipping.FeelingLucky(nums, coin)
	found := false
	for _, v := range nums {
		if result == v {
			found = true
			break
		}
	}
	assert.True(t, found, "FeelingLucky should return one of the items in the slice")

	// Test with an empty slice, expecting the zero value.
	var empty []int
	resultEmpty := flipping.FeelingLucky(empty, coin)
	assert.Equal(t, 0, resultEmpty, "FeelingLucky with an empty slice should return the zero value")
}
