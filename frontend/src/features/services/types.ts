import type { SharedTime } from "@/shared/composables/useGlobalTimeRange";

export interface ServiceGraphFilters extends SharedTime {
  selected?: string;
}
