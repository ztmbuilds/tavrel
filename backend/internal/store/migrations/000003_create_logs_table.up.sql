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

CREATE INDEX IF NOT EXISTS idx_logs_service_timestamp ON logs (service, timestamp_ns);

CREATE INDEX IF NOT EXISTS idx_logs_trace_id_timestamp ON logs (trace_id, timestamp_ns);

CREATE INDEX IF NOT EXISTS idx_logs_span_id ON logs (span_id);

CREATE INDEX IF NOT EXISTS idx_logs_severity_number ON logs (severity_number);

CREATE INDEX IF NOT EXISTS idx_logs_env_ts ON logs (environment, timestamp_ns, id);

CREATE INDEX IF NOT EXISTS idx_logs_severity_ts ON logs (severity_number, timestamp_ns, id);

CREATE INDEX IF NOT EXISTS idx_logs_span_ts ON logs (span_id, timestamp_ns, id);


