package api

import (
	"context"
	"net/http"
	"noname/internal/store"
	"time"
)

const defaultStatsWindow = 24 * time.Hour

type dashboardStatsStore interface {
	GetDashboardStats(ctx context.Context, start, end time.Time) (store.DashboardStats, error)
}

func newStatsHandler(s dashboardStatsStore) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		start, end, err := parseTimeRange(r)
		if err != nil {
			return err
		}

		now := time.Now()
		switch {
		case start == nil && end == nil:
			startAt := now.Add(-defaultStatsWindow)
			start, end = &startAt, &now
		case start == nil:
			startAt := end.Add(-defaultStatsWindow)
			start = &startAt
		case end == nil:
			end = &now
		}
		if start.After(*end) {
			return BadRequest("invalid time range: start must not be after end", nil)
		}

		stats, err := s.GetDashboardStats(r.Context(), *start, *end)
		if err != nil {
			return Internal(err)
		}

		writeJson(w, http.StatusOK, &Response{
			Message: "Dashboard statistics retrieved successfully",
			Data:    stats,
			Meta:    &Meta{},
		})
		return nil
	}
}
