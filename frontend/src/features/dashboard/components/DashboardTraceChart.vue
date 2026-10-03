<script setup lang="ts">
import { computed, onUnmounted, ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { use } from "echarts/core";
import type { ComposeOption } from "echarts/core";
import { BarChart } from "echarts/charts";
import type { BarSeriesOption } from "echarts/charts";
import { GridComponent, TooltipComponent } from "echarts/components";
import type {
  GridComponentOption,
  TooltipComponentOption,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import VChart from "vue-echarts";

import type {
  DashboardOverview,
  DashboardRange,
  TraceTimeBucket,
} from "@/features/dashboard/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import { formatCount } from "@/shared/utils/format";
import { CUSTOM_RANGE } from "@/shared/utils/date-range";

const props = defineProps<{
  buckets: TraceTimeBucket[];
  overview: DashboardOverview;
  range: DashboardRange;
}>();

use([CanvasRenderer, BarChart, GridComponent, TooltipComponent]);

type ChartOption = ComposeOption<
  BarSeriesOption | GridComponentOption | TooltipComponentOption
>;

const router = useRouter();

function cssVar(name: string, fallback: string): string {
  const value = getComputedStyle(document.documentElement)
    .getPropertyValue(name)
    .trim();
  return value === "" ? fallback : value;
}

const palette = {
  ok: cssVar("--color-chart-ok", "#22c55e"),
  error: cssVar("--color-chart-error", "#ef4444"),
  tertiary: cssVar("--color-tertiary", "#7c8a96"),
  subtle: cssVar("--color-subtle", "#1c242d"),
  border: cssVar("--color-default", "#2b3540"),
  raised: cssVar("--color-raised", "#1a222b"),
  heading: cssVar("--color-heading", "#f5f7f8"),
};

const reducedMotionQuery = window.matchMedia(
  "(prefers-reduced-motion: reduce)",
);
const reducedMotion = ref(reducedMotionQuery.matches);
function onMotionChange(event: MediaQueryListEvent): void {
  reducedMotion.value = event.matches;
}
reducedMotionQuery.addEventListener("change", onMotionChange);
onUnmounted(() =>
  reducedMotionQuery.removeEventListener("change", onMotionChange),
);

const totalCount = computed(() => props.overview.traceCount);
const errorCount = computed(() => props.overview.errorTraceCount);
const okCount = computed(() => totalCount.value - errorCount.value);
const PERCENT_SCALE = 100;
const errorRate = computed(() =>
  totalCount.value === 0
    ? "0.0%"
    : `${((errorCount.value / totalCount.value) * PERCENT_SCALE).toFixed(1)}%`,
);
const isEmpty = computed(() => totalCount.value === 0);
const compactCountFormatter = new Intl.NumberFormat(undefined, {
  notation: "compact",
  maximumFractionDigits: 1,
});

function formatBucketTime(value: string): string {
  const options: Intl.DateTimeFormatOptions =
    props.range === "7d"
      ? {
          month: "short",
          day: "numeric",
          hour: "2-digit",
          hourCycle: "h23",
        }
      : { hour: "2-digit", minute: "2-digit", hourCycle: "h23" };
  return new Intl.DateTimeFormat(undefined, options).format(new Date(value));
}

const labels = computed(() =>
  props.buckets.map((bucket) => formatBucketTime(bucket.startTime)),
);
const okCounts = computed(() => props.buckets.map((bucket) => bucket.okCount));
const errorCounts = computed(() =>
  props.buckets.map((bucket) => bucket.errorCount),
);

const option = computed<ChartOption>(() => ({
  animation: !reducedMotion.value,
  animationDuration: 200,
  animationDurationUpdate: 200,
  grid: {
    left: 12,
    right: 12,
    top: 12,
    bottom: 8,
    containLabel: true,
  },
  tooltip: {
    trigger: "axis",
    axisPointer: { type: "shadow" },
    backgroundColor: palette.raised,
    borderColor: palette.border,
    borderWidth: 1,
    textStyle: { color: palette.heading, fontSize: 11 },
  },
  xAxis: {
    type: "category",
    data: labels.value,
    axisTick: { show: false },
    axisLine: { lineStyle: { color: palette.border } },
    axisLabel: {
      color: palette.tertiary,
      fontSize: 11,
      hideOverlap: true,
    },
  },
  yAxis: {
    type: "value",
    minInterval: 1,
    axisLine: { show: false },
    axisLabel: {
      color: palette.tertiary,
      fontSize: 11,
      formatter: (value: number) => compactCountFormatter.format(value),
    },
    splitLine: { lineStyle: { color: palette.subtle } },
  },
  series: [
    {
      name: "OK",
      type: "bar",
      stack: "traces",
      data: okCounts.value,
      barMaxWidth: 10,
      itemStyle: { color: palette.ok },
    },
    {
      name: "Error",
      type: "bar",
      stack: "traces",
      data: errorCounts.value,
      barMaxWidth: 10,
      itemStyle: { color: palette.error },
    },
  ],
}));

const summaryLabel = computed(
  () =>
    `Traces over time. ${formatCount(totalCount.value)} total traces, ${errorRate.value} error rate, ${formatCount(okCount.value)} OK, ${formatCount(errorCount.value)} Error.`,
);

const allTracesTo = computed(() => ({
  name: "traces",
  query: { range: props.range },
}));
const errorTracesTo = computed(() => ({
  name: "traces",
  query: { status: "error", range: props.range },
}));
const servicesTo = computed(() => ({
  name: "services",
  query: { range: props.range },
}));

function tracesTo(bucket: TraceTimeBucket): {
  name: string;
  query: Record<string, string>;
} {
  return {
    name: "traces",
    query: {
      range: CUSTOM_RANGE,
      start: bucket.startTime,
      end: bucket.endTime,
    },
  };
}

function onChartClick(event: unknown): void {
  if (typeof event !== "object" || event === null) return;
  if (!("dataIndex" in event) || typeof event.dataIndex !== "number") return;
  const bucket = props.buckets[event.dataIndex];
  if (bucket === undefined) return;
  void router.push(tracesTo(bucket));
}
</script>

<template>
  <section
    class="border-subtle bg-panel overflow-hidden rounded-xl border"
    aria-labelledby="dashboard-traces-title"
  >
    <header
      class="flex flex-wrap items-center justify-between gap-x-6 gap-y-4 px-5 pt-4"
    >
      <h2
        id="dashboard-traces-title"
        class="text-heading text-sm font-semibold"
      >
        Traces
      </h2>
      <ul
        class="m-0 flex list-none items-center gap-5 p-0"
        aria-label="Trace outcomes"
      >
        <li class="text-secondary flex items-center gap-2 text-xs">
          <span class="bg-chart-ok size-2.5 rounded-full" aria-hidden="true" />
          OK
        </li>
        <li class="text-secondary flex items-center gap-2 text-xs">
          <span
            class="bg-chart-error size-2.5 rounded-full"
            aria-hidden="true"
          />
          Error
        </li>
      </ul>
    </header>

    <div class="flex flex-wrap items-end gap-x-8 gap-y-3 px-5 pt-5">
      <RouterLink
        :to="allTracesTo"
        :aria-label="`View ${formatCount(overview.traceCount)} traces in this window`"
        class="group flex flex-col gap-1"
      >
        <span class="text-tertiary text-xs">Total traces</span>
        <span
          class="text-heading group-hover:text-primary text-2xl leading-none font-semibold tabular-nums"
        >
          {{ compactCountFormatter.format(overview.traceCount) }}
        </span>
      </RouterLink>
      <RouterLink
        :to="errorTracesTo"
        :aria-label="`View failed traces, ${errorRate} error rate`"
        class="group flex flex-col gap-1"
      >
        <span class="text-tertiary text-xs">Error rate</span>
        <span
          class="text-chart-error text-2xl leading-none font-semibold tabular-nums group-hover:brightness-125"
        >
          {{ errorRate }}
        </span>
      </RouterLink>
      <RouterLink
        :to="servicesTo"
        :aria-label="`View ${formatCount(overview.activeServiceCount)} active services`"
        class="group flex flex-col gap-1"
      >
        <span class="text-tertiary text-xs">Active services</span>
        <span
          class="text-heading group-hover:text-primary text-2xl leading-none font-semibold tabular-nums"
        >
          {{ compactCountFormatter.format(overview.activeServiceCount) }}
        </span>
      </RouterLink>
    </div>

    <EmptyState
      v-if="isEmpty"
      title="No traces in this window"
      message="Trace activity will appear here when traces arrive."
    />
    <div v-else class="px-3 pt-4 pb-3">
      <div class="h-72">
        <VChart
          :option="option"
          autoresize
          role="img"
          :aria-label="summaryLabel"
          class="cursor-pointer"
          @click="onChartClick"
        />
      </div>
    </div>
  </section>
</template>
