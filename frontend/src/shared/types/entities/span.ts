import type { Attributes, SpanEventDto } from "@/shared/api/otel";

export type SpanKind =
  | "unspecified"
  | "internal"
  | "server"
  | "client"
  | "producer"
  | "consumer";

export type SpanStatus = "unset" | "ok" | "error";

export interface Span {
  spanId: string;
  traceId: string;
  parentSpanId: string;
  service: string;
  environment: string;
  name: string;
  kind: SpanKind;
  status: SpanStatus;
  statusMessage: string;
  startTime: string;
  endTime: string;
  attributes: Attributes;
  events: SpanEventDto[];
}
