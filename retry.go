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

type RetryResult struct {
	MaxAttempts int
	Attempts    []Attempt
}

type Attempt struct {
	Delay          time.Duration
	ContextError   error
	OperationError error
}

func (a Attempt) Err() error {
	if a.ContextError != nil {
		return a.ContextError
	}
	return a.OperationError
}

func (rr RetryResult) FinalOperationError() error {
	if len(rr.Attempts) == 0 {
		return nil
	}
	i := len(rr.Attempts) - 1
	attempt := rr.Attempts[i]
	if attempt.ContextError == nil && attempt.OperationError == nil {
		return nil
	}
	for {
		if attempt.OperationError != nil {
			return attempt.OperationError
		}
		if i--; i < 0 {
			break
		}
		attempt = rr.Attempts[i]
	}
	return nil
}

func (rr RetryResult) FinalAttemptError() error {
	if len(rr.Attempts) == 0 {
		return nil
	}
	return rr.Attempts[len(rr.Attempts)-1].Err()
}

// Retry retries to call function operation at least once.
func (r Retryer) Retry(ctx context.Context, operation func(context.Context) error) RetryResult {
	var err error
	maxAttempts := r.MaxAttempts
	delayer := r.delayer()

	result := RetryResult{MaxAttempts: maxAttempts}
	var attempt Attempt
	appendAttempt := func() {
		result.Attempts = append(result.Attempts, attempt)
	}

	for {
		err = operation(ctx)
		if err == nil {
			appendAttempt()
			break
		}

		var berr *BreakError
		if errors.As(err, &berr) {
			attempt.OperationError = berr.Err
			appendAttempt()
			break
		}

		attempt.OperationError = err
		appendAttempt()
		if maxAttempts > 0 && len(result.Attempts) >= maxAttempts {
			break
		}

		d := delayer.Delay(len(result.Attempts))
		attempt = Attempt{Delay: d}
		if err = sleep(ctx, d); err != nil {
			attempt.ContextError = err
			appendAttempt()
			break
		}
	}

	return result
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
