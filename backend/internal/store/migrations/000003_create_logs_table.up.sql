CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trace_id TEXT,
    span_id TEXT,
    timestamp_ns INTEGER NOT NULL,
    severity TEXT NOT NULL,
    body TEXT NOT NULL,
    service TEXT NOT NULL,
    attributes TEXT
);

CREATE INDEX IF NOT EXISTS idx_logs_trace_id ON logs(trace_id);
CREATE INDEX IF NOT EXISTS idx_logs_filter ON logs(timestamp_ns, service, severity);