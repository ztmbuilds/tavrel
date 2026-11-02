import type { Attributes } from "@/shared/api/otel";

import type { SpanStatus } from "./span";

export interface SpanLinkTarget {
  name: string;
  service: string;
  status: SpanStatus;
}

export interface SpanLink {
  traceId: string;
  spanId: string;
  attributes: Attributes;
  target: SpanLinkTarget | null;
}
