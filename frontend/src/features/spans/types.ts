import type { SharedTime } from "@/shared/composables/useGlobalTimeRange";

export type SpanStatusFilter = "all" | "error" | "ok";

export interface SpansFilters extends SharedTime {
  search: string;
  status: SpanStatusFilter;
  service: string[];
  environment: string[];
}

export interface SpansListParams extends SpansFilters {
  cursor?: string;
}
