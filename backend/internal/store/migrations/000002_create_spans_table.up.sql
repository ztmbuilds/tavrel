CREATE TABLE IF NOT EXISTS spans (
    span_id TEXT PRIMARY KEY,
    trace_id TEXT NOT NULL,
    parent_span_id TEXT,
    service TEXT NOT NULL,
    name TEXT NOT NULL,
    kind INTEGER NOT NULL DEFAULT 0,
    start_time_ns INTEGER NOT NULL,
    end_time_ns INTEGER NOT NULL,
    status_code INTEGER NOT NULL DEFAULT 0,
    status_message TEXT,
    attributes TEXT,
    events TEXT,
    FOREIGN KEY (trace_id) REFERENCES traces(trace_id)
);

CREATE INDEX IF NOT EXISTS idx_spans_start_time_span_id ON spans (start_time_ns,span_id)

CREATE INDEX IF NOT EXISTS idx_spans_trace_id_start_time_span_id ON spans (trace_id, start_time_ns, span_id)

