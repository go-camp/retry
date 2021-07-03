package retry

import (
	"math"
	"math/rand"
	"time"
)

// Delayer calculates the delay before the next attempt after this attempt.
//
// The delay before the first attempt(<=0) is always 0.
type Delayer interface {
	Delay(attempt int) time.Duration
}

// NopDelayer provides zero delay between attempts.
type NopDelayer struct{}

// Delay always returns 0.
func (NopDelayer) Delay(int) time.Duration {
	return 0
}

// ConstantDelayer provides constant delay between attempts.
type ConstantDelayer struct {
	// Duration is the constant delay.
	Duration time.Duration
}

// Delay returns a constant delay.
func (d ConstantDelayer) Delay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	if d.Duration < 0 {
		return 0
	}
	return d.Duration
}

// ExpDelayer provides exponential and random growth delay between attempts.
type ExpDelayer struct {
	// Initial is the delay after first attempt.
	// If Initial less than or equals to 0, the ExpInitial will be used.
	Initial time.Duration
	// Multiplier is a multiplicator used to calculate the delay after the last attempt.
	// If Multiplier less than 1, the ExpMultiplier will be used.
	Multiplier float64
	// Max is max day between attempts.
	// If Max less than or equals to 0, the math.MaxInt64 will be used.
	Max time.Duration
	// Rand gives delay a variation of ±(Rand%).
	// The range of Rand is [0,100].
	// If Rand greater than 100, 100 will be used.
	// If Rand greater than 0, the delay may greater than Max.
	Rand uint8
}

// Default values for ExpDelayer.
const (
	ExpInitial    = 500 * time.Millisecond
	ExpMultiplier = 1.5
)

// DefaultDelayer is an ExpDelayer with Rand 50.
var DefaultDelayer = ExpDelayer{Rand: 50}

func (d ExpDelayer) initial() time.Duration {
	if d.Initial <= 0 {
		return ExpInitial
	}
	return d.Initial
}

func (d ExpDelayer) multiplier() float64 {
	if d.Multiplier < 1 || math.IsNaN(d.Multiplier) || math.IsInf(d.Multiplier, 0) {
		return ExpMultiplier
	}
	return d.Multiplier
}

func (d ExpDelayer) max() time.Duration {
	if d.Max > 0 {
		return d.Max
	}
	return math.MaxInt64
}

func (d ExpDelayer) percent() uint8 {
	if d.Rand <= 100 {
		return d.Rand
	}
	return 100
}

func (d ExpDelayer) delay(attempt int) time.Duration {
	init := d.initial()
	mul := d.multiplier()
	max := d.max()
	n := float64(init) * math.Pow(mul, float64(attempt-1))
	if n > float64(max) {
		return max
	}
	return time.Duration(n)
}

func (d ExpDelayer) rand(b time.Duration) time.Duration {
	per := d.percent()
	if per == 0 {
		return b
	}

	bf := float64(b)
	delta := float64(per) / 100 * bf
	min := bf - delta
	max := bf + delta
	delay := time.Duration(min + (rand.Float64() * (max - min + 1)))
	if delay < 0 {
		return math.MaxInt64
	}
	return delay
}

func (d ExpDelayer) Delay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	return d.rand(d.delay(attempt))
}
