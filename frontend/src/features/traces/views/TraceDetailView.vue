<script setup lang="ts">
import { computed, nextTick, onUnmounted, provide, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import SpanPanel from "@/features/traces/components/SpanPanel.vue";
import TraceHeader from "@/features/traces/components/TraceHeader.vue";
import WaterfallRow from "@/features/traces/components/WaterfallRow.vue";
import { useTraceGraph } from "@/features/traces/composables/useTraceGraph";
import {
  DEFAULT_LABEL_WIDTH_REM,
  LABEL_RESIZE_STEP_REM,
  MAX_LABEL_WIDTH_REM,
  MIN_LABEL_WIDTH_REM,
  NARROW_LABEL_WIDTH_REM,
  waterfallKey,
} from "@/features/traces/composables/waterfall-context";
import { isAsyncSpan } from "@/features/traces/api";
import {
  useTraceDetailQuery,
  useTraceSpanLogsQuery,
} from "@/features/traces/queries";
import type { SpanLogWindow, TraceDetailTab } from "@/features/traces/types";
import { ApiError } from "@/shared/api/client";
import { toSpanEventViews } from "@/shared/api/otel";
import Button from "@/shared/components/ui/Button.vue";
import EmptyState from "@/shared/components/EmptyState.vue";
import type { Span } from "@/shared/types/entities/span";
import { truncateTraceId } from "@/shared/utils/format";
import {
  formatDurationMs,
  formatRelativeTime,
} from "@/shared/utils/time";

const route = useRoute();
const router = useRouter();

const FULL_PCT = 100;
const MIN_POSITION_WIDTH_PCT = 0.5;
const GRID_LINE_COUNT = 6;
const COPY_FLASH_MS = 1600;
const NARROW_VIEWPORT_PX = 760;
const PX_PER_REM = 16;

const idRef = computed((): string => {
  const raw = route.params.id;
  if (Array.isArray(raw)) return raw[0] ?? "";
  return raw ?? "";
});

function initialSpanId(): string | null {
  const raw = route.query.spanId;
  return typeof raw === "string" && raw !== "" ? raw : null;
}

const selectedId = ref<string | null>(initialSpanId());
const activeTab = ref<TraceDetailTab>("overview");
const copied = ref(false);
const collapsed = ref<Set<string>>(new Set<string>());
const waterfallEl = ref<HTMLDivElement | null>(null);
const waterfallWrapEl = ref<HTMLDivElement | null>(null);
const labelWidth = ref(
  window.innerWidth <= NARROW_VIEWPORT_PX
    ? NARROW_LABEL_WIDTH_REM
    : DEFAULT_LABEL_WIDTH_REM,
);

let copyTimer: ReturnType<typeof setTimeout> | undefined;

const detailQuery = useTraceDetailQuery(idRef);
const detail = computed(() => detailQuery.data.value);

const graph = useTraceGraph(detail);

const selectedSpan = computed((): Span | null => {
  if (selectedId.value === null) return null;
  return graph.findSpan(selectedId.value);
});

const selectedEvents = computed(() =>
  selectedSpan.value === null
    ? []
    : toSpanEventViews(selectedSpan.value.events, selectedSpan.value.startTime),
);

const spanLogsWindow = computed((): SpanLogWindow | null => {
  const span = selectedSpan.value;
  if (span === null) return null;
  return { spanId: span.spanId, start: span.startTime, end: span.endTime };
});

const spanLogsQuery = useTraceSpanLogsQuery(idRef, spanLogsWindow);
const spanLogs = computed(() => spanLogsQuery.data.value ?? []);

const isNotFound = computed(
  () =>
    detailQuery.error.value instanceof ApiError &&
    detailQuery.error.value.isNotFound(),
);

const shortTraceId = computed(() => truncateTraceId(idRef.value));
const durationLabel = computed(() => formatDurationMs(graph.totalDurationMs.value));
const relativeLabel = computed(() =>
  detail.value === undefined
    ? ""
    : `Started ${formatRelativeTime(detail.value.trace.startTime)}`,
);

const selectedDurationMs = computed(() =>
  selectedSpan.value === null ? 0 : graph.spanDurationMs(selectedSpan.value),
);
const selectedStartMs = computed(() =>
  selectedSpan.value === null ? 0 : graph.spanStartMs(selectedSpan.value),
);
const selectedSelfMs = computed(() =>
  selectedId.value === null ? 0 : graph.selfTimeMs(selectedId.value),
);
const selectedChildCount = computed(() =>
  selectedId.value === null ? 0 : graph.childCount(selectedId.value),
);
const selectedAncestors = computed(() =>
  selectedId.value === null ? [] : graph.ancestors(selectedId.value),
);
const selectedPercent = computed(() => {
  if (graph.totalDurationMs.value <= 0) return 0;
  return Math.round((selectedDurationMs.value / graph.totalDurationMs.value) * FULL_PCT);
});
const selectedPosition = computed(() => {
  if (graph.totalDurationMs.value <= 0) return { start: 0, width: 0 };
  return {
    start: (selectedStartMs.value / graph.totalDurationMs.value) * FULL_PCT,
    width: Math.max(
      (selectedDurationMs.value / graph.totalDurationMs.value) * FULL_PCT,
      MIN_POSITION_WIDTH_PCT,
    ),
  };
});
const selectedStatusText = computed(() =>
  selectedSpan.value === null
    ? "No span selected"
    : `${selectedSpan.value.name}, ${formatDurationMs(selectedDurationMs.value)}, ${selectedSpan.value.status}`,
);

const gridStyles = computed(() =>
  Array.from({ length: GRID_LINE_COUNT }, (_, index) => {
    const fraction = index / (GRID_LINE_COUNT - 1);
    const trackStart = labelWidth.value + 1;
    return {
      left: `calc(${trackStart}rem + (100% - ${trackStart + 1}rem) * ${fraction})`,
    };
  }),
);

watch(idRef, () => {
  selectedId.value = initialSpanId();
  activeTab.value = "overview";
  collapsed.value = new Set<string>();
  labelWidth.value =
    window.innerWidth <= NARROW_VIEWPORT_PX
      ? NARROW_LABEL_WIDTH_REM
      : DEFAULT_LABEL_WIDTH_REM;
});

onUnmounted(() => {
  if (copyTimer !== undefined) clearTimeout(copyTimer);
});

function select(spanId: string): void {
  selectedId.value = spanId;
  activeTab.value = "overview";
  const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  void nextTick(() => {
    const el = waterfallEl.value?.querySelector(`[data-span-id="${CSS.escape(spanId)}"]`);
    el?.scrollIntoView({ block: "nearest", behavior: reduceMotion ? "auto" : "smooth" });
  });
}

function toggleCollapse(spanId: string): void {
  const next = new Set(collapsed.value);
  if (next.has(spanId)) next.delete(spanId);
  else next.add(spanId);
  collapsed.value = next;
}

provide(waterfallKey, {
  selectedId,
  collapsed,
  totalMs: graph.totalDurationMs,
  labelWidth,
  select,
  toggleCollapse,
  spanStartMs: graph.spanStartMs,
  spanDurationMs: graph.spanDurationMs,
  selfTimeMs: graph.selfTimeMs,
});

function flatVisible(): Span[] {
  return graph.visibleSpans(collapsed);
}

function onKeydown(event: KeyboardEvent): void {
  if (event.key === "Escape") {
    selectedId.value = null;
    return;
  }
  const flat = flatVisible();
  if (flat.length === 0) return;
  const currentIndex = flat.findIndex((span) => span.spanId === selectedId.value);
  if (event.key === "j" || event.key === "ArrowDown") {
    event.preventDefault();
    const nextIndex =
      currentIndex < 0 ? 0 : Math.min(currentIndex + 1, flat.length - 1);
    const next = flat[nextIndex];
    if (next !== undefined) select(next.spanId);
  } else if (event.key === "k" || event.key === "ArrowUp") {
    event.preventDefault();
    const prevIndex = currentIndex < 0 ? 0 : Math.max(currentIndex - 1, 0);
    const prev = flat[prevIndex];
    if (prev !== undefined) select(prev.spanId);
  } else if (event.key === "Enter" && selectedId.value !== null) {
    if (graph.childCount(selectedId.value) > 0) toggleCollapse(selectedId.value);
  }
}

function clampLabelWidth(widthRem: number): number {
  return Math.min(Math.max(widthRem, MIN_LABEL_WIDTH_REM), MAX_LABEL_WIDTH_REM);
}

function resetLabelWidth(): void {
  labelWidth.value =
    window.innerWidth <= NARROW_VIEWPORT_PX
      ? NARROW_LABEL_WIDTH_REM
      : DEFAULT_LABEL_WIDTH_REM;
}

function onLabelResize(event: PointerEvent): void {
  const handle = event.currentTarget;
  if (!(handle instanceof HTMLElement)) return;
  const wrap = waterfallWrapEl.value;
  if (wrap === null) return;
  const wrapRect = wrap.getBoundingClientRect();
  const move = (moveEvent: PointerEvent): void => {
    const px = moveEvent.clientX - wrapRect.left;
    const maxRem = wrapRect.width / PX_PER_REM / 2;
    labelWidth.value = Math.min(
      clampLabelWidth(px / PX_PER_REM),
      maxRem,
    );
  };
  const stop = (): void => {
    window.removeEventListener("pointermove", move);
    window.removeEventListener("pointerup", stop);
    window.removeEventListener("pointercancel", stop);
  };
  window.addEventListener("pointermove", move);
  window.addEventListener("pointerup", stop);
  window.addEventListener("pointercancel", stop);
}

function onLabelResizeKey(event: KeyboardEvent): void {
  if (event.key === "ArrowLeft") {
    event.preventDefault();
    labelWidth.value = clampLabelWidth(labelWidth.value - LABEL_RESIZE_STEP_REM);
  } else if (event.key === "ArrowRight") {
    event.preventDefault();
    labelWidth.value = clampLabelWidth(labelWidth.value + LABEL_RESIZE_STEP_REM);
  }
}

function back(): void {
  void router.push({ name: "traces" });
}

function copyTraceId(): void {
  void navigator.clipboard?.writeText(idRef.value).catch(() => {});
  copied.value = true;
  if (copyTimer !== undefined) clearTimeout(copyTimer);
  copyTimer = setTimeout(() => {
    copied.value = false;
  }, COPY_FLASH_MS);
}
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <a class="skip-link" href="#waterfall">Skip to spans</a>
    <div
      v-if="detailQuery.isPending.value"
      class="text-tertiary flex flex-col items-center gap-1.5 px-5 py-12 text-center text-[0.8125rem]"
      role="status"
    >
      <span class="text-heading text-sm font-semibold">Loading trace…</span>
    </div>
    <div
      v-else-if="detailQuery.isError.value"
      class="flex min-h-0 flex-1 flex-col items-center justify-center px-5"
      role="alert"
    >
      <EmptyState
        :title="isNotFound ? 'Trace not found' : 'Couldn’t load trace'"
        :message="
          isNotFound
            ? 'This trace id doesn’t match anything in retention. It may have expired or never existed.'
            : 'Something went wrong fetching this trace. Try again.'
        "
        :action-label="isNotFound ? 'Back to traces' : 'Retry'"
        @clear="isNotFound ? back() : void detailQuery.refetch()"
      />
      <Button v-if="!isNotFound" variant="secondary" class="mt-1" @click="back()">
        Back to traces
      </Button>
    </div>
    <template v-else-if="detail">
      <TraceHeader
        :short-trace-id="shortTraceId"
        :root-name="detail.trace.rootName"
        :root-service="detail.trace.rootService"
        :root-kind="graph.rootSpan.value?.kind ?? 'unspecified'"
        :environment="detail.trace.environment"
        :status="detail.trace.status"
        :duration-label="durationLabel"
        :relative-label="relativeLabel"
        :span-count="graph.spanCount.value"
        :copied="copied"
        :legend="graph.serviceLegend.value"
        :issues="graph.issues.value"
        :diagnosis="graph.diagnosis.value"
        @back="back"
        @copy="copyTraceId"
        @select="select"
      />
      <EmptyState
        v-if="graph.spanCount.value === 0"
        title="No spans in this trace"
        message="The trace was recorded without any spans."
      />
      <div v-else class="relative flex min-h-0 flex-1">
        <div ref="waterfallWrapEl" class="bg-panel relative flex min-w-0 flex-1 overflow-hidden">
          <div class="pointer-events-none absolute inset-0 z-0 max-[47.5rem]:hidden" aria-hidden="true">
            <i
              v-for="(style, index) in gridStyles"
              :key="index"
              class="bg-subtle absolute top-0 bottom-0 w-px opacity-70"
              :style="style"
            />
          </div>
          <div
            id="waterfall"
            ref="waterfallEl"
            tabindex="0"
            role="tree"
            aria-label="Spans in this trace. Use J and K to move, Enter to expand or collapse."
            class="relative z-1 min-w-0 flex-1 overflow-x-hidden overflow-y-auto pt-1.5 pb-6"
            @keydown="onKeydown"
          >
            <WaterfallRow
              v-for="(node, index) in graph.rootNodes.value"
              :key="node.span.spanId"
              :node="node"
              :depth="1"
              :level="1"
              :guides="[]"
              :is-last="index === graph.rootNodes.value.length - 1"
            />
          </div>
          <div
            role="separator"
            aria-orientation="vertical"
            aria-label="Resize span names column. Arrow keys adjust, double-click resets."
            tabindex="0"
            class="hover:bg-strong absolute top-0 bottom-0 z-10 w-2 cursor-col-resize touch-none rounded bg-transparent transition-colors duration-120 ease-out"
            :style="{ left: `calc(${labelWidth}rem - 0.25rem)` }"
            @pointerdown="onLabelResize"
            @keydown="onLabelResizeKey"
            @dblclick="resetLabelWidth"
          />
        </div>

        <aside
          v-if="selectedSpan"
          class="border-subtle bg-surface absolute top-0 right-0 bottom-0 z-20 border-l shadow-xl"
          aria-label="Span details"
        >
          <SpanPanel
            v-model:active-tab="activeTab"
            :span="selectedSpan"
            :ancestors="selectedAncestors"
            :child-count="selectedChildCount"
            :self-time-ms="selectedSelfMs"
            :percent-of-trace="selectedPercent"
            :position-start-pct="selectedPosition.start"
            :position-width-pct="selectedPosition.width"
            :start-offset-ms="selectedStartMs"
            :duration-ms="selectedDurationMs"
            :is-async="isAsyncSpan(selectedSpan.kind)"
            :events="selectedEvents"
            :logs="spanLogs"
            :logs-pending="spanLogsQuery.isPending.value"
            :logs-error="spanLogsQuery.isError.value"
            @close="selectedId = null"
            @select-ancestor="select"
            @retry-logs="void spanLogsQuery.refetch()"
          />
        </aside>
      </div>
    </template>
    <div class="sr-only" role="status">{{ selectedStatusText }}</div>
  </div>
</template>
