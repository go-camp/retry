package retry_test

import (
	"fmt"
	"time"

	"github.com/go-camp/retry"
)

func ExampleConstantDelayer() {
	delayer := retry.ConstantDelayer{Duration: time.Second}
	for attempt := 1; attempt <= 3; attempt++ {
		delay := delayer.Delay(attempt)
		fmt.Println(delay)
	}
	// Output:
	// 1s
	// 1s
	// 1s
}

func ExampleExpDelayer() {
	delayer := retry.ExpDelayer{
		Initial:    1 * time.Second,
		Multiplier: 2,
		Max:        20 * time.Second,
	}
	for attempt := 1; attempt <= 10; attempt++ {
		delay := delayer.Delay(attempt)
		fmt.Println(delay)
	}
	// Output:
	// 1s
	// 2s
	// 4s
	// 8s
	// 16s
	// 20s
	// 20s
	// 20s
	// 20s
	// 20s
}

func ExampleExpDelayer_Rand() {
	delayer := retry.ExpDelayer{
		Initial:    time.Second,
		Multiplier: 2,
		Max:        20 * time.Second,
		Rand:       50,
	}
	delayRanges := [][2]time.Duration{
		{500 * time.Millisecond, 1500 * time.Millisecond},
		{1 * time.Second, 3 * time.Second},
		{2 * time.Second, 6 * time.Second},
		{4 * time.Second, 12 * time.Second},
		{8 * time.Second, 24 * time.Second},
		{10 * time.Second, 30 * time.Second},
		{10 * time.Second, 30 * time.Second},
		{10 * time.Second, 30 * time.Second},
		{10 * time.Second, 30 * time.Second},
		{10 * time.Second, 30 * time.Second},
	}
	for i, delayRange := range delayRanges {
		attempt := i + 1
		delay := delayer.Delay(attempt)
		if delay >= delayRange[0] && delay <= delayRange[1] {
			fmt.Println(delayRange)
		} else {
			fmt.Println(delay, "out range of", delayRange)
		}
	}
	// Output:
	// [500ms 1.5s]
	// [1s 3s]
	// [2s 6s]
	// [4s 12s]
	// [8s 24s]
	// [10s 30s]
	// [10s 30s]
	// [10s 30s]
	// [10s 30s]
	// [10s 30s]
}
