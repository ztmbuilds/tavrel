const SpanKind = {
  Unspecified: "unspecified",
  Internal: "internal",
  Server: "server",
  Client: "client",
  Producer: "producer",
  Consumer: "consumer",
} as const;
export type SpanKind = (typeof SpanKind)[keyof typeof SpanKind];

const SpanStatus = {
  Unset: "unset",
  Ok: "ok",
  Error: "error",
} as const;
export type SpanStatus = (typeof SpanStatus)[keyof typeof SpanStatus];

export interface Span {
  spanId: string;
  traceId: string;
  parentSpanId: string;
  service: string;
  name: string;
  kind: SpanKind;
  status: SpanStatus;
  statusMessage: string;
  startTime: string;
  endTime: string;
  attributes: unknown;
  events: unknown;
}
