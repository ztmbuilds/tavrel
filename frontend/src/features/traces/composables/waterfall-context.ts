import { inject, type InjectionKey, type Ref } from "vue";

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
export const DEFAULT_PANEL_WIDTH_REM = 22.5;
export const MIN_PANEL_WIDTH_REM = 16;
export const PANEL_RESIZE_STEP_REM = 0.5;

export function useWaterfallContext(): WaterfallContext {
  const context = inject(waterfallKey);
  if (context === undefined)
    throw new Error("Missing trace waterfall provider");
  return context;
}
