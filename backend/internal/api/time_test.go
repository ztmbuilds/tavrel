package api

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseTimeRangeAllowsOmittedBounds(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/v1/traces", nil)
	start, end, err := parseTimeRange(request)
	if err != nil {
		t.Fatal(err)
	}
	if start != nil || end != nil {
		t.Fatalf("got bounds %v and %v, want both omitted", start, end)
	}
}

func TestParseTimeRangeReadsRFC3339Bounds(t *testing.T) {
	startWant := time.Date(2026, time.January, 2, 3, 0, 0, 0, time.UTC)
	endWant := startWant.Add(time.Hour)
	request := httptest.NewRequest("GET", "/api/v1/traces?start=2026-01-02T03:00:00Z&end=2026-01-02T04:00:00Z", nil)

	start, end, err := parseTimeRange(request)
	if err != nil {
		t.Fatal(err)
	}
	if start == nil || !start.Equal(startWant) || end == nil || !end.Equal(endWant) {
		t.Fatalf("got bounds %v and %v, want %s and %s", start, end, startWant, endWant)
	}
}

func TestParseTimeRangeRejectsInvalidBounds(t *testing.T) {
	for _, target := range []string{
		"/api/v1/traces?start=not-a-date",
		"/api/v1/traces?end=not-a-date",
		"/api/v1/traces?start=2026-01-02T04:00:00Z&end=2026-01-02T03:00:00Z",
	} {
		request := httptest.NewRequest("GET", target, nil)
		_, _, err := parseTimeRange(request)
		var appErr *AppError
		if !errors.As(err, &appErr) || appErr.StatusCode != 400 {
			t.Errorf("parseTimeRange(%q) error = %v, want HTTP 400", target, err)
		}
	}
}
