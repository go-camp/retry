package retry

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestDelayerAttemptLTEZero(t *testing.T) {
	delayers := []Delayer{
		NopDelayer{},
		ConstantDelayer{Duration: time.Second},
		ExpDelayer{},
	}
	for i := 0; i < 10; i++ {
		attempt := -rand.Int()
		for _, delayer := range delayers {
			d := delayer.Delay(attempt)
			if d != 0 {
				t.Errorf("want (%T).Delay(%d) to return 0 but get %d", delayer, attempt, d)
			}
		}
	}
}

func TestExpDelayer(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		delayer := ExpDelayer{}
		delay := delayer.Delay(1)
		if delay <= 0 {
			t.Fatalf("want delay greater than 0 but get %d", delay)
		}
	})

	t.Run("delay overflow", func(t *testing.T) {
		delayer := ExpDelayer{
			Initial:    math.MaxInt64,
			Multiplier: 2,
			Rand:       1,
		}
		delay := delayer.Delay(math.MaxInt32)
		if delay <= 0 {
			t.Fatalf("want delay greater than 0 but get %d", delay)
		}
	})
}
