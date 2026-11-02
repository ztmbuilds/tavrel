CREATE TABLE IF NOT EXISTS traces (
    trace_id TEXT PRIMARY KEY NOT NULL,
    root_service TEXT NOT NULL,
    environment TEXT,
    root_name TEXT NOT NULL,
    start_time_ns INTEGER NOT NULL,
    end_time_ns INTEGER NOT NULL,
    span_count INTEGER NOT NULL,
    error INTEGER NOT NULL,
    updated_at_ns INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_traces_start_time_trace_id
ON traces (start_time_ns, trace_id );

CREATE INDEX IF NOT EXISTS idx_traces_end_time
ON traces (end_time_ns);

CREATE INDEX IF NOT EXISTS idx_traces_env_start
ON traces (environment, start_time_ns, trace_id);