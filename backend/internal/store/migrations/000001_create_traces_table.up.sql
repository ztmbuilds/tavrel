CREATE TABLE IF NOT EXISTS traces (
    trace_id TEXT PRIMARY KEY,
    root_service TEXT NOT NULL,
    environment TEXT,
    root_name TEXT NOT NULL,
    start_time_ns INTEGER NOT NULL,
    end_time_ns INTEGER NOT NULL,
    span_count INTEGER NOT NULL,
    error INTEGER NOT NULL
);

