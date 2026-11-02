<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";

import type { Trace } from "@/shared/types/entities/trace";
import Badge from "@/shared/components/ui/Badge.vue";
import { cn } from "@/shared/utils/cn";
import { SHORT_LIST_ID_LENGTH, truncateTraceId } from "@/shared/utils/format";
import {
  diffMsPrecise,
  formatDurationMs,
  formatRelativeTime,
  formatShortDateTimeWithSeconds,
} from "@/shared/utils/time";
import CopyButton from "@/shared/components/CopyButton.vue";

const props = defineProps<{
  trace: Trace;
  maxDurationMs: number;
}>();

const FULL_METER_PERCENT = 100;
const MIN_METER_PERCENT = 6;

const durationMs = computed(() => {
  const value = diffMsPrecise(props.trace.startTime, props.trace.endTime);
  if (Number.isNaN(value)) return 0;
  return Math.max(value, 0);
});

const durationLabel = computed(() => formatDurationMs(durationMs.value));

const relativeLabel = computed(() => formatRelativeTime(props.trace.startTime));

const dateTimeLabel = computed(() =>
  formatShortDateTimeWithSeconds(props.trace.startTime),
);

const shortTraceId = computed(() =>
  truncateTraceId(props.trace.traceId, SHORT_LIST_ID_LENGTH),
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
    `${props.trace.rootName}, ${props.trace.status}, ${durationLabel.value}, ${props.trace.spanCount} spans, ${dateTimeLabel.value}`,
);
</script>

<template>
  <div
    :class="
      cn(
        'group border-subtle relative flex min-h-[3.375rem] items-center gap-3 border-b px-5 py-2 transition-colors duration-120 ease-out',
        trace.status === 'error'
          ? 'bg-error/[0.05] hover:bg-error/[0.08] shadow-[inset_2px_0_0_0_var(--color-error)]'
          : 'hover:bg-white/[0.035]',
      )
    "
    role="listitem"
  >
    <RouterLink
      :to="{ name: 'trace-detail', params: { id: trace.traceId } }"
      class="absolute inset-0 z-10"
      :aria-label="rowLabel"
    />
    <div
      class="text-tertiary w-[5.25rem] shrink-0 font-mono text-xs tabular-nums"
    >
      <span class="inline-flex items-center gap-1">
        <span class="truncate" :title="trace.traceId">{{ shortTraceId }}</span>
        <span class="relative z-20 inline-flex">
          <CopyButton
            :text="trace.traceId"
            label="Copy trace ID"
            title="Copy trace ID"
          />
        </span>
      </span>
    </div>
    <div class="min-w-0 flex-1">
      <div class="flex min-w-0 items-center gap-2">
        <span
          :class="
            cn(
              'truncate text-[0.78125rem] leading-snug font-medium',
              trace.status === 'error' ? 'text-error-text' : 'text-heading',
            )
          "
          :title="trace.rootName"
        >
          {{ trace.rootName }}
        </span>
        <Badge
          v-if="trace.status === 'error'"
          variant="error"
          class="h-5 shrink-0 px-2 text-[0.6875rem]"
        >
          Error
        </Badge>
      </div>
      <div class="text-tertiary mt-0.5 truncate text-[0.6875rem]">
        {{ trace.rootService || "Unknown service" }}
      </div>
    </div>
    <div
      class="text-tertiary w-28 shrink-0 truncate text-xs max-[53.75rem]:hidden"
      :title="trace.environment"
    >
      {{ trace.environment }}
    </div>
    <div
      class="relative h-1.5 w-[9.375rem] shrink-0 overflow-hidden rounded-[3px] bg-white/[0.06] max-[53.75rem]:hidden"
      aria-hidden="true"
    >
      <div
        :class="
          cn(
            'absolute top-0 bottom-0 left-0 rounded-[3px]',
            trace.status === 'error' ? 'bg-chart-error' : 'bg-chart-ok',
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
      class="text-tertiary w-[3.875rem] shrink-0 text-right text-xs tabular-nums max-[53.75rem]:hidden"
    >
      {{ trace.spanCount }}
    </div>
    <div
      class="text-tertiary w-[10.75rem] shrink-0 truncate text-right text-xs tabular-nums max-[40rem]:hidden"
      :title="dateTimeLabel"
    >
      {{ relativeLabel }}
    </div>
  </div>
</template>
