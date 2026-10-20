package cron

import (
	"context"
	"strconv"
	"testing"
	"time"
)

func TestStartDoesNotRunWithNonPositiveInterval(t *testing.T) {
	for _, days := range []int{0, -1, 106752} {
		t.Run(strconv.Itoa(days), func(t *testing.T) {
			called := make(chan struct{}, 1)
			c := New(days, func(context.Context, time.Time) error {
				called <- struct{}{}
				return nil
			})

			c.Start(context.Background())
			defer c.Stop()

			select {
			case <-called:
				t.Fatal("retention ran with a non-positive interval")
			case <-time.After(50 * time.Millisecond):
			}
		})
	}
}

func TestStartRunsImmediatelyWithPositiveInterval(t *testing.T) {
	called := make(chan time.Time, 1)
	c := New(1, func(_ context.Context, cutoff time.Time) error {
		called <- cutoff
		return nil
	})

	c.Start(context.Background())
	defer c.Stop()

	select {
	case cutoff := <-called:
		if cutoff.After(time.Now()) {
			t.Fatalf("cutoff %v is in the future", cutoff)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("retention did not run")
	}
}
