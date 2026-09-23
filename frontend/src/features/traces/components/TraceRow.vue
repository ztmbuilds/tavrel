<script setup lang="ts">
import { computed } from "vue";

import type { Trace } from "@/shared/types/entities/trace";
import { cn } from "@/shared/utils/cn";
import { truncateTraceId } from "@/shared/utils/format";
import {
  diffMsPrecise,
  formatDurationMs,
  formatRelativeTime,
  formatShortDateTimeWithSeconds,
} from "@/shared/utils/time";

const props = defineProps<{
  trace: Trace;
  maxDurationMs: number;
}>();

defineEmits<{
  openTrace: [traceId: string];
}>();

const FULL_METER_PERCENT = 100;

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

const shortTraceId = computed(() => truncateTraceId(props.trace.traceId));

const meterWidth = computed(() => {
  if (props.maxDurationMs <= 0) return 0;
  return Math.min(
    FULL_METER_PERCENT,
    (durationMs.value / props.maxDurationMs) * FULL_METER_PERCENT,
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
        'border-subtle hover:bg-hover flex cursor-pointer items-center gap-3 border-b border-l-2 border-l-transparent px-4 py-2 transition-colors duration-120 ease-out',
        trace.status === 'error' && 'border-l-error',
      )
    "
    role="listitem"
    tabindex="0"
    :aria-label="rowLabel"
    @click="$emit('openTrace', trace.traceId)"
    @keydown.enter.prevent="$emit('openTrace', trace.traceId)"
    @keydown.space.prevent="$emit('openTrace', trace.traceId)"
  >
    <div
      class="text-tertiary w-[3.75rem] shrink-0 truncate font-mono text-xs tabular-nums"
      :title="trace.traceId"
    >
      {{ shortTraceId }}
    </div>
    <div class="min-w-0 flex-1">
      <div
        :class="
          cn(
            'truncate text-[0.78125rem] leading-snug font-medium',
            trace.status === 'error' ? 'text-error' : 'text-heading',
          )
        "
      >
        <span :title="trace.rootName">{{ trace.rootName }}</span>
      </div>
    </div>
    <div
      class="text-tertiary w-[5.5rem] shrink-0 truncate text-xs max-[53.75rem]:hidden"
      :title="trace.environment"
    >
      {{ trace.environment }}
    </div>
    <div
      class="bg-surface relative h-3.5 w-[9.375rem] shrink-0 overflow-hidden rounded-sm max-[53.75rem]:hidden"
      aria-hidden="true"
    >
      <div
        :class="
          cn(
            'absolute top-0 bottom-0 left-0 rounded-sm opacity-90',
            trace.status === 'error' ? 'bg-error' : 'bg-ok',
          )
        "
        :style="{ width: `${meterWidth}%` }"
      />
    </div>
    <div
      class="text-secondary w-[3.625rem] shrink-0 text-right font-mono text-xs tabular-nums"
    >
      {{ durationLabel }}
    </div>
    <div
      class="text-tertiary w-[3.875rem] shrink-0 text-right font-mono text-xs tabular-nums max-[53.75rem]:hidden"
    >
      {{ trace.spanCount }}
    </div>
    <div
      class="text-tertiary w-[10.75rem] shrink-0 truncate text-right font-mono text-xs tabular-nums max-[40rem]:hidden"
      :title="relativeLabel"
    >
      {{ dateTimeLabel }}
    </div>
  </div>
</template>
