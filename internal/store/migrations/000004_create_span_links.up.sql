CREATE TABLE IF NOT EXISTS span_links (
    trace_id TEXT NOT NULL,
    span_id TEXT NOT NULL,
    linked_trace_id TEXT NOT NULL,
    linked_span_id TEXT NOT NULL,
    attributes TEXT NOT NULL DEFAULT '{}',
    PRIMARY KEY (trace_id, span_id, linked_trace_id, linked_span_id)
);
