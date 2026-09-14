export type LogLevelFilter = "all" | "info" | "warn" | "error";

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
