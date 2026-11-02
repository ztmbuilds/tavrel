CREATE TABLE IF NOT EXISTS spans (
    span_id TEXT PRIMARY KEY,
    trace_id TEXT NOT NULL,
    parent_span_id TEXT,
    service TEXT NOT NULL,
    environment TEXT,
    name TEXT NOT NULL,
    kind INTEGER NOT NULL DEFAULT 0,
    start_time_ns INTEGER NOT NULL,
    end_time_ns INTEGER NOT NULL,
    status_code INTEGER NOT NULL DEFAULT 0,
    status_message TEXT,
    attributes TEXT,
    events TEXT,
    resource_id INTEGER NOT NULL REFERENCES resources(id),
    FOREIGN KEY (trace_id) REFERENCES traces(trace_id)
);

CREATE INDEX IF NOT EXISTS idx_spans_start_time_span_id ON spans (start_time_ns,span_id);

CREATE INDEX IF NOT EXISTS idx_spans_trace_id_start_time_span_id ON spans (trace_id, start_time_ns, span_id); 

CREATE INDEX IF NOT EXISTS idx_spans_service_start_time ON spans (service, start_time_ns, span_id);

CREATE INDEX IF NOT EXISTS idx_spans_service_trace ON spans (service, trace_id);

CREATE INDEX IF NOT EXISTS idx_spans_env_start ON spans (environment, start_time_ns, span_id);

CREATE INDEX IF NOT EXISTS idx_spans_status_start ON spans (status_code, start_time_ns, span_id);

