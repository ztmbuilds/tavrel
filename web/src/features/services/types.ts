import type { SharedTime } from "@/shared/composables/useGlobalTimeRange";

export interface ServiceGraphFilters extends SharedTime {
  selected?: string;
}

export type ServiceMapNodeData = {
  dimmed: boolean;
  focused: boolean;
  onSelect: () => void;
  onMove: (delta: { x: number; y: number }) => void;
};
