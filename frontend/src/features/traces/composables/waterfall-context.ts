import { ref, type InjectionKey, type Ref } from "vue";

import type { Span } from "@/shared/types/entities/span";

export interface WaterfallContext {
  selectedId: Ref<string | null>;
  collapsed: Ref<Set<string>>;
  totalMs: Ref<number>;
  labelWidth: Ref<number>;
  select: (spanId: string) => void;
  toggleCollapse: (spanId: string) => void;
  spanStartMs: (span: Span) => number;
  spanDurationMs: (span: Span) => number;
  selfTimeMs: (spanId: string) => number;
}

export const waterfallKey: InjectionKey<WaterfallContext> =
  Symbol("trace-waterfall");

export const DEFAULT_LABEL_WIDTH_REM = 24;
export const NARROW_LABEL_WIDTH_REM = 9.375;
export const MIN_LABEL_WIDTH_REM = 8;
export const MAX_LABEL_WIDTH_REM = 40;
export const LABEL_RESIZE_STEP_REM = 0.5;

export function fallbackWaterfallContext(): WaterfallContext {
  return {
    selectedId: ref(null),
    collapsed: ref(new Set<string>()),
    totalMs: ref(0),
    labelWidth: ref(DEFAULT_LABEL_WIDTH_REM),
    select: () => {},
    toggleCollapse: () => {},
    spanStartMs: () => 0,
    spanDurationMs: () => 0,
    selfTimeMs: () => 0,
  };
}
