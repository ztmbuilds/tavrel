CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trace_id TEXT,
    span_id TEXT,
    timestamp_ns INTEGER NOT NULL,
    severity_text TEXT ,
    severity_number INTEGER NOT NULL,
    body TEXT NOT NULL,
    service TEXT NOT NULL,
    environment TEXT,
    attributes TEXT,
    resource_attributes TEXT
);

CREATE INDEX IF NOT EXISTS idx_logs_timestamp_id ON logs (timestamp_ns,id);

