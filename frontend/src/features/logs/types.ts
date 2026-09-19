import type { SharedTime } from "@/shared/composables/useGlobalTimeRange";

export type LogLevelFilter = "all" | "info" | "warn" | "error";

export interface LogsFilters extends SharedTime {
  search: string;
  level: LogLevelFilter;
  service: string[];
  traceId: string;
  spanId: string;
}

export interface LogsListParams extends LogsFilters {
  cursor?: string;
}
