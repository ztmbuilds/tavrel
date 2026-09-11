const LogLevelFilter = {
  All: "all",
  Info: "info",
  Warn: "warn",
  Error: "error",
} as const;
export type LogLevelFilter =
  (typeof LogLevelFilter)[keyof typeof LogLevelFilter];

export interface LogsFilters {
  search: string;
  level: LogLevelFilter;
  service: string;
  traceId: string;
  spanId: string;
  start?: string;
  end?: string;
}

export interface LogsListParams extends LogsFilters {
  cursor?: string;
}

export const DEFAULT_LOGS_FILTERS: LogsFilters = {
  search: "",
  level: "all",
  service: "all",
  traceId: "",
  spanId: "",
};
