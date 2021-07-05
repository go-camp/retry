package retry

import (
	"errors"
	"testing"
)

func TestBreak(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := Break(nil)
		if err != nil {
			t.Fatalf("want nil but get %v", err)
		}
	})

	t.Run("err", func(t *testing.T) {
		ierr := errors.New("ierr")
		err := Break(ierr)
		var berr *BreakError
		if !errors.Is(err, berr) {
			t.Fatal("want err is *BreakError")
		}
		if !errors.Is(err, ierr) {
			t.Fatalf("want error is %v", ierr)
		}
		if !errors.As(err, &berr) {
			t.Fatalf("want err as **BreakError")
		}
	})
}
