package api

import (
	"net/http"
	"time"
)

func parseTimeRange(r *http.Request) (start, end *time.Time, err error) {
	query := r.URL.Query()

	if value := query.Get("start"); value != "" {
		parsed, parseErr := time.Parse(time.RFC3339, value)
		if parseErr != nil {
			return nil, nil, BadRequest("invalid start: must be ISO 8601", parseErr)
		}
		start = &parsed
	}
	if value := query.Get("end"); value != "" {
		parsed, parseErr := time.Parse(time.RFC3339, value)
		if parseErr != nil {
			return nil, nil, BadRequest("invalid end: must be ISO 8601", parseErr)
		}
		end = &parsed
	}
	if start != nil && end != nil && start.After(*end) {
		return nil, nil, BadRequest("invalid time range: start must not be after end", nil)
	}
	return start, end, nil
}
