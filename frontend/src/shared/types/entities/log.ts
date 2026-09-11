const LogLevel = {
  Info: "info",
  Warn: "warn",
  Error: "error",
} as const;
export type LogLevel = (typeof LogLevel)[keyof typeof LogLevel];

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
  attributes: unknown;
  resourceAttributes: unknown;
}
