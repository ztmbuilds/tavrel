import type { Attributes } from "@/shared/api/otel";

export type LogLevel = "info" | "warn" | "error";

export interface Log {
  id: number;
  traceId: string;
  spanId: string;
  timestamp: string;
  service: string;
  environment: string;
  severityText: string;
  severityNumber: number;
  level: LogLevel;
  message: string;
  host: string;
  attributes: Attributes;
  resourceAttributes: Attributes;
}
