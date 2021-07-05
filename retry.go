package retry

import (
	"context"
	"errors"
	"time"
)

// Retryer contains basic retry logic using Delayer.
type Retryer struct {
	// If Delayer is nil, the DefaultDelayer will be used.
	Delayer Delayer
	// MaxAttempts is the maximum number of calls to op.
	// MaxAttempts is 0 means there is no constraint on the number of attempts.
	MaxAttempts int
}

// BreakError indicates that the operation should not be retried.
type BreakError struct {
	Err error
}

func (e *BreakError) Error() string {
	return e.Err.Error()
}
func (e *BreakError) Unwrap() error {
	return e.Err
}
func (e *BreakError) Is(target error) bool {
	_, ok := target.(*BreakError)
	return ok
}

// Break wraps a non-nil error into BreakError.
func Break(err error) error {
	if err == nil {
		return nil
	}
	return &BreakError{
		Err: err,
	}
}

func (r Retryer) delayer() Delayer {
	if r.Delayer == nil {
		return DefaultDelayer
	}
	return r.Delayer
}

// Retry retries to call function op at least once.
//
// If the op function returns nil, the retry will be terminated,
// and the Retry function returns nil.
//
// If the op function returns BreakError, the retry will be terminated,
// and the Retry function returns an error unwrap from BreakError.
//
// If the calls of op function exceed maximum, the retry will be terminated,
// and the Retry function returns the error returned from the op function.
//
// If the context is canceled, the retry will be terminated,
// and the Retry function returns the error returned from the op function.
func (r Retryer) Retry(ctx context.Context, op func(context.Context) error) error {
	var err error
	maxAttempts := r.MaxAttempts
	attempt := 0
	delayer := r.delayer()
	for {
		attempt++

		err = op(ctx)
		if err == nil {
			return nil
		}

		var berr *BreakError
		if errors.As(err, &berr) {
			return berr.Err
		}

		if maxAttempts > 0 && attempt >= maxAttempts {
			return err
		}

		select {
		case <-ctx.Done():
			return err
		default:
		}
		d := delayer.Delay(attempt)
		deadline, ok := ctx.Deadline()
		if ok {
			if time.Until(deadline) < d {
				return err
			}
		}
		if cerr := sleep(ctx, d); cerr != nil {
			return err
		}
	}
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
