import type { SpanKind } from "@/shared/types/entities/span";

const SpanStatusFilter = {
  All: "all",
  Error: "error",
  Ok: "ok",
} as const;
export type SpanStatusFilter =
  (typeof SpanStatusFilter)[keyof typeof SpanStatusFilter];

export interface SpansFilters {
  search: string;
  status: SpanStatusFilter;
  kind: "all" | SpanKind;
  service: string;
  start?: string;
  end?: string;
}

export interface SpansListParams extends SpansFilters {
  cursor?: string;
}

export interface TimePreset {
  value: string;
  label: string;
  minutes: number;
}

export const TIME_PRESETS: readonly TimePreset[] = [
  { value: "15m", label: "15m", minutes: 15 },
  { value: "1h", label: "1h", minutes: 60 },
  { value: "6h", label: "6h", minutes: 360 },
  { value: "24h", label: "24h", minutes: 1440 },
  { value: "7d", label: "7d", minutes: 10080 },
];

export const CUSTOM_RANGE = "custom";

export const DEFAULT_SPANS_FILTERS: SpansFilters = {
  search: "",
  status: "all",
  kind: "all",
  service: "all",
};
