<script setup lang="ts">
import { computed, onUnmounted, ref } from "vue";
import { useRouter } from "vue-router";
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
import VChart, { type Exposed } from "vue-echarts";

import type {
  DashboardOverview,
  DashboardRange,
  TraceTimeBucket,
} from "@/features/dashboard/types";
import Card from "@/shared/components/ui/Card.vue";
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
const chartRef = ref<Exposed | null>(null);

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
  font: cssVar("--font-sans", "sans-serif"),
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

const hoveredIndex = ref<number | null>(null);
const BAR_RADIUS_PX = 3;

const readout = computed(() => {
  const index = hoveredIndex.value;
  if (index !== null) {
    const bucket = props.buckets[index];
    const label = labels.value[index];
    if (bucket !== undefined && label !== undefined) {
      return `${label} · ${formatCount(bucket.okCount)} OK · ${formatCount(bucket.errorCount)} Error`;
    }
  }
  return `${formatCount(totalCount.value)} traces · ${errorRate.value} error rate`;
});

const option = computed<ChartOption>(() => ({
  animation: !reducedMotion.value,
  animationDuration: 200,
  animationDurationUpdate: 200,
  textStyle: { fontFamily: palette.font },
  grid: {
    left: 8,
    right: 12,
    top: 12,
    bottom: 8,
    containLabel: true,
  },
  tooltip: {
    trigger: "axis",
    showContent: false,
    axisPointer: {
      type: "shadow",
      shadowStyle: { color: "rgba(255, 255, 255, 0.06)" },
    },
  },
  xAxis: {
    type: "category",
    data: labels.value,
    axisTick: { show: false },
    axisLine: { lineStyle: { color: palette.border } },
    axisLabel: {
      color: palette.tertiary,
      fontSize: 12,
      hideOverlap: true,
    },
  },
  yAxis: {
    type: "value",
    minInterval: 1,
    axisLine: { show: false },
    axisLabel: {
      color: palette.tertiary,
      fontSize: 12,
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
      itemStyle: {
        color: palette.ok,
        borderRadius: [BAR_RADIUS_PX, BAR_RADIUS_PX, 0, 0],
      },
    },
    {
      name: "Error",
      type: "bar",
      stack: "traces",
      data: errorCounts.value,
      barMaxWidth: 10,
      itemStyle: {
        color: palette.error,
        borderRadius: [BAR_RADIUS_PX, BAR_RADIUS_PX, 0, 0],
      },
    },
  ],
}));

const summaryLabel = computed(
  () =>
    `Traces over time. ${formatCount(totalCount.value)} total traces, ${errorRate.value} error rate, ${formatCount(okCount.value)} OK, ${formatCount(errorCount.value)} Error.`,
);

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

function indexFromPixel(x: number, y: number): number | null {
  const chart = chartRef.value?.chart;
  if (chart === undefined) return null;
  const value = chart.convertFromPixel({ xAxisIndex: 0, yAxisIndex: 0 }, [
    x,
    y,
  ]);
  if (!Array.isArray(value)) return null;
  const xValue = value[0];
  if (typeof xValue === "number") {
    if (
      Number.isInteger(xValue) &&
      xValue >= 0 &&
      xValue < props.buckets.length
    )
      return xValue;
    return null;
  }
  if (typeof xValue === "string") {
    const index = labels.value.indexOf(xValue);
    return index >= 0 ? index : null;
  }
  return null;
}

function onChartMouseMove(event: MouseEvent): void {
  if (!(event.currentTarget instanceof HTMLElement)) return;
  const rect = event.currentTarget.getBoundingClientRect();
  const index = indexFromPixel(
    event.clientX - rect.left,
    event.clientY - rect.top,
  );
  hoveredIndex.value = index;
}

function clearHover(): void {
  hoveredIndex.value = null;
}
</script>

<template>
  <Card title="Traces over time">
    <template #action>
      <ul
        class="m-0 flex list-none items-center gap-4 p-0"
        aria-label="Trace outcomes"
      >
        <li class="text-secondary flex items-center gap-2 text-xs">
          <span class="bg-chart-ok size-2 rounded-full" aria-hidden="true" />
          OK
        </li>
        <li class="text-secondary flex items-center gap-2 text-xs">
          <span class="bg-chart-error size-2 rounded-full" aria-hidden="true" />
          Error
        </li>
      </ul>
    </template>

    <p
      class="text-tertiary m-0 text-xs leading-relaxed tabular-nums"
      aria-live="polite"
    >
      {{ readout }}
    </p>

    <p v-if="isEmpty" class="text-tertiary py-14 text-center text-[0.8125rem]">
      No traces in this window.
    </p>
    <div
      v-else
      class="mt-3 h-72"
      @mousemove="onChartMouseMove"
      @mouseleave="clearHover"
    >
      <VChart
        ref="chartRef"
        :option="option"
        autoresize
        role="img"
        :aria-label="summaryLabel"
        class="cursor-pointer"
        @click="onChartClick"
      />
    </div>
  </Card>
</template>
