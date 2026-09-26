<script setup lang="ts">
import { computed } from "vue";

import type { Span } from "@/shared/types/entities/span";
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

defineEmits<{
  openTrace: [traceId: string];
}>();

const durationMs = computed(() => {
  const value = diffMsPrecise(props.span.startTime, props.span.endTime);
  if (Number.isNaN(value)) return 0;
  return Math.max(value, 0);
});

const FULL_METER_PERCENT = 100;

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
  return Math.min(
    FULL_METER_PERCENT,
    (durationMs.value / props.maxDurationMs) * FULL_METER_PERCENT,
  );
});

const rowLabel = computed(
  () =>
    `${props.span.name}, ${props.span.service}, ${durationLabel.value}, ${dateTimeLabel.value}`,
);
</script>

<template>
  <div
    :class="
      cn(
        'border-subtle hover:bg-hover flex cursor-pointer items-center gap-3 border-b border-l-2 border-l-transparent px-4 py-2 transition-colors duration-120 ease-out',
        span.status === 'error' && 'border-l-error',
      )
    "
    role="listitem"
    tabindex="0"
    :aria-label="rowLabel"
    @click="$emit('openTrace', span.traceId)"
    @keydown.enter.prevent="$emit('openTrace', span.traceId)"
    @keydown.space.prevent="$emit('openTrace', span.traceId)"
  >
    <div class="flex w-4 shrink-0 justify-center" aria-hidden="true">
      <span
        class="size-1.75 rounded-full"
        :style="{ background: serviceColor(span.service) }"
      />
    </div>
    <div class="min-w-0 flex-1">
      <div
        :class="
          cn(
            'text-primary truncate text-sm leading-snug font-medium',
            span.status === 'error' && 'text-error',
          )
        "
      >
        <span :title="span.name">{{ span.name }}</span>
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
          <span class="group inline-flex items-center" :title="span.spanId">
            <span class="text-secondary font-mono tabular-nums">{{
              shortSpanId
            }}</span>
            <CopyButton
              compact
              :text="span.spanId"
              label="Copy span ID"
              title="Copy span ID"
            />
          </span>
          <span aria-hidden="true">·</span>
        </span>
        <span class="inline-flex shrink-0 items-center gap-2">
          <span class="text-tertiary">trace</span>
          <span class="group inline-flex items-center">
            <button
              type="button"
              class="text-secondary hover:text-signal cursor-pointer border-none bg-transparent p-0 font-mono text-xs tabular-nums hover:underline hover:underline-offset-2"
              :title="`Open trace ${span.traceId}`"
              @click.stop="$emit('openTrace', span.traceId)"
            >
              {{ shortTraceId }}
            </button>
            <CopyButton
              compact
              :text="span.traceId"
              label="Copy trace ID"
              title="Copy trace ID"
            />
          </span>
        </span>
      </div>
    </div>
    <div
      class="text-tertiary w-[5.5rem] shrink-0 truncate text-xs max-[53.75rem]:hidden"
      :title="span.environment"
      aria-hidden="true"
    >
      {{ span.environment }}
    </div>
    <div
      class="bg-surface relative h-3.5 w-[8.125rem] shrink-0 overflow-hidden rounded-sm max-[53.75rem]:hidden"
      aria-hidden="true"
    >
      <div
        :class="
          cn(
            'absolute top-0 bottom-0 left-0 rounded-sm opacity-90',
            span.status === 'error' ? 'bg-error' : 'bg-ok',
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
      class="text-tertiary w-[10.75rem] shrink-0 truncate text-right font-mono text-xs tabular-nums"
      :title="relativeLabel"
    >
      {{ dateTimeLabel }}
    </div>
  </div>
</template>
