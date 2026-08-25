package cron

import (
	"context"
	"log/slog"
	"time"
)

type Cron struct {
	interval time.Duration
	deleteFn func(ctx context.Context, cutoff time.Time) error
	stop     chan struct{}
}

func New(days int, deleteFn func(ctx context.Context, cutoff time.Time) error) *Cron {
	return &Cron{
		interval: time.Duration(days) * 24 * time.Hour,
		deleteFn: deleteFn,
		stop:     make(chan struct{}),
	}
}

func (c *Cron) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)

	go func() {
		defer ticker.Stop()

		c.run(ctx, time.Now().Add(-c.interval)) // run once immediately

		for {
			select {
			case <-c.stop:
				return
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				c.run(ctx, t.Add(-c.interval))
			}
		}
	}()
}

func (c *Cron) run(ctx context.Context, cutoff time.Time) {
	slog.Info("running retention cron")
	if err := withRetry(func() error {
		return c.deleteFn(ctx, cutoff)
	}); err != nil {
		slog.Error("retention cron failed", "error", err)
	}
	slog.Info("retention cron run complete")
}

func (c *Cron) Stop() {
	close(c.stop)
}

func withRetry(fn func() error) error {
	var err error
	for i := range 5 {
		err = fn()

		if err == nil {
			return nil
		}

		time.Sleep(time.Duration(50*(i+1)) * time.Millisecond)

	}
	return err
}
