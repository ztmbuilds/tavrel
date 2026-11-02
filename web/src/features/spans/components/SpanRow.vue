<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";

import type { Span } from "@/shared/types/entities/span";
import Badge from "@/shared/components/ui/Badge.vue";
import { cn } from "@/shared/utils/cn";
import { SHORT_LIST_ID_LENGTH } from "@/shared/utils/format";
import { serviceColor } from "@/shared/utils/service-color";
import {
  diffMsPrecise,
  formatDurationMs,
  formatRelativeTime,
  formatShortDateTimeWithSeconds,
} from "@/shared/utils/time";
import CopyButton from "@/shared/components/CopyButton.vue";

const props = defineProps<{
  span: Span;
  maxDurationMs: number;
}>();

const durationMs = computed(() => {
  const value = diffMsPrecise(props.span.startTime, props.span.endTime);
  if (Number.isNaN(value)) return 0;
  return Math.max(value, 0);
});

const FULL_METER_PERCENT = 100;
const MIN_METER_PERCENT = 6;

const durationLabel = computed(() => formatDurationMs(durationMs.value));

const relativeLabel = computed(() => formatRelativeTime(props.span.startTime));

const dateTimeLabel = computed(() =>
  formatShortDateTimeWithSeconds(props.span.startTime),
);

const shortSpanId = computed(() =>
  props.span.spanId.slice(0, SHORT_LIST_ID_LENGTH),
);

const shortTraceId = computed(() =>
  props.span.traceId.slice(0, SHORT_LIST_ID_LENGTH),
);

const meterWidth = computed(() => {
  if (props.maxDurationMs <= 0) return 0;
  return Math.max(
    MIN_METER_PERCENT,
    Math.min(
      FULL_METER_PERCENT,
      (durationMs.value / props.maxDurationMs) * FULL_METER_PERCENT,
    ),
  );
});

const rowLabel = computed(
  () =>
    `${props.span.name}, ${props.span.service}, ${props.span.status}, ${durationLabel.value}, ${dateTimeLabel.value}`,
);
</script>

<template>
  <div
    :class="
      cn(
        'group border-subtle relative flex min-h-[3.375rem] items-center gap-3 border-b px-5 py-2 transition-colors duration-120 ease-out',
        span.status === 'error'
          ? 'bg-error/[0.05] hover:bg-error/[0.08] shadow-[inset_2px_0_0_0_var(--color-error)]'
          : 'hover:bg-white/[0.035]',
      )
    "
    role="listitem"
  >
    <RouterLink
      :to="{ name: 'trace-detail', params: { id: span.traceId } }"
      class="absolute inset-0 z-10"
      :aria-label="rowLabel"
    />
    <div class="flex w-4 shrink-0 justify-center" aria-hidden="true">
      <span
        class="size-1.75 rounded-full"
        :style="{ background: serviceColor(span.service) }"
      />
    </div>
    <div class="min-w-0 flex-1">
      <div class="flex min-w-0 items-center gap-2">
        <span
          :class="
            cn(
              'truncate text-[0.78125rem] leading-snug font-medium',
              span.status === 'error' ? 'text-error-text' : 'text-primary',
            )
          "
          :title="span.name"
        >
          {{ span.name }}
        </span>
        <Badge
          v-if="span.status === 'error'"
          variant="error"
          class="h-5 shrink-0 px-2 text-[0.6875rem]"
        >
          Error
        </Badge>
      </div>
      <div
        class="text-tertiary mt-0.5 flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs"
      >
        <span class="inline-flex min-w-0 shrink items-center gap-2">
          <span class="text-secondary min-w-0 truncate">{{
            span.service
          }}</span>
          <span aria-hidden="true">·</span>
        </span>
        <span class="inline-flex shrink-0 items-center gap-2">
          <span class="text-tertiary">span</span>
          <span class="inline-flex items-center" :title="span.spanId">
            <span class="text-secondary font-mono tabular-nums">{{
              shortSpanId
            }}</span>
            <span class="relative z-20 inline-flex">
              <CopyButton
                compact
                :text="span.spanId"
                label="Copy span ID"
                title="Copy span ID"
              />
            </span>
          </span>
          <span aria-hidden="true">·</span>
        </span>
        <span class="inline-flex shrink-0 items-center gap-2">
          <span class="text-tertiary">trace</span>
          <span class="inline-flex items-center">
            <span class="text-secondary font-mono tabular-nums">{{
              shortTraceId
            }}</span>
            <span class="relative z-20 inline-flex">
              <CopyButton
                compact
                :text="span.traceId"
                label="Copy trace ID"
                title="Copy trace ID"
              />
            </span>
          </span>
        </span>
      </div>
    </div>
    <div
      class="text-tertiary w-28 shrink-0 truncate text-xs max-[53.75rem]:hidden"
      :title="span.environment"
      aria-hidden="true"
    >
      {{ span.environment }}
    </div>
    <div
      class="relative h-1.5 w-[8.125rem] shrink-0 overflow-hidden rounded-[3px] bg-white/[0.06] max-[53.75rem]:hidden"
      aria-hidden="true"
    >
      <div
        :class="
          cn(
            'absolute top-0 bottom-0 left-0 rounded-[3px]',
            span.status === 'error' ? 'bg-chart-error' : 'bg-chart-ok',
          )
        "
        :style="{ width: `${meterWidth}%` }"
      />
    </div>
    <div
      class="text-secondary w-[3.625rem] shrink-0 text-right text-xs tabular-nums"
    >
      {{ durationLabel }}
    </div>
    <div
      class="text-tertiary w-[10.75rem] shrink-0 truncate text-right text-xs tabular-nums max-[40rem]:hidden"
      :title="dateTimeLabel"
    >
      {{ relativeLabel }}
    </div>
  </div>
</template>
