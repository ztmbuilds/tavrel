<script setup lang="ts">
import { computed } from "vue";

import xIcon from "@/assets/icons/x.svg";
import type { Log } from "@/shared/types/entities/log";
import type { Span } from "@/shared/types/entities/span";
import { cn } from "@/shared/utils/cn";
import { serviceColor } from "@/shared/utils/service-color";
import { formatDurationMs } from "@/shared/utils/time";

import SpanAttributes from "./SpanAttributes.vue";

import type { AncestorEntry } from "../composables/useTraceGraph";
import type { SpanEventView, TraceDetailTab } from "../types";

const props = defineProps<{
  span: Span;
  ancestors: AncestorEntry[];
  childCount: number;
  selfTimeMs: number;
  percentOfTrace: number;
  positionStartPct: number;
  positionWidthPct: number;
  startOffsetMs: number;
  durationMs: number;
  isAsync: boolean;
  events: SpanEventView[];
  logs: Log[];
  logsPending: boolean;
  logsError: boolean;
}>();

defineEmits<{
  close: [];
  selectAncestor: [spanId: string];
  retryLogs: [];
}>();

const activeTab = defineModel<TraceDetailTab>("activeTab", { default: "overview" });

const TABS: readonly { value: TraceDetailTab; label: string }[] = [
  { value: "overview", label: "Overview" },
  { value: "attributes", label: "Attributes" },
  { value: "events", label: "Events" },
  { value: "logs", label: "Logs" },
];

const statusLabel = computed(() =>
  props.span.status === "error"
    ? "error"
    : props.span.status === "ok"
      ? "ok"
      : "unset",
);

function logOffsetMs(timestamp: string): number {
  return Math.max(Date.parse(timestamp) - Date.parse(props.span.startTime), 0);
}
</script>

<template>
  <div class="flex h-full w-[22.5rem] max-w-[90vw] shrink-0 flex-col max-[68.75rem]:w-[min(22.5rem,90vw)]">
    <div class="border-subtle shrink-0 border-b px-4 pt-3 pb-2.5">
      <div class="flex items-start justify-between gap-2.5">
        <p class="text-heading m-0 text-[0.90625rem] leading-snug font-semibold text-pretty">
          {{ span.name }}
        </p>
        <button
          type="button"
          class="text-tertiary hover:bg-hover hover:text-secondary flex size-7 shrink-0 cursor-pointer items-center justify-center rounded-md border-none bg-transparent transition-colors duration-120 ease-out"
          aria-label="Close span details"
          @click="$emit('close')"
        >
          <img :src="xIcon" alt="" aria-hidden="true" class="size-3" />
        </button>
      </div>
      <div class="text-secondary mt-1.5 flex items-center gap-1.5 text-xs leading-relaxed">
        <span
          class="size-[0.4375rem] rounded-full"
          :style="{ background: serviceColor(span.service) }"
          aria-hidden="true"
        />
        {{ span.service }} · {{ span.kind }}
        <span class="sr-only">, status {{ statusLabel }}</span>
      </div>
      <nav class="text-tertiary mt-2.5 flex flex-wrap items-center gap-0.5 text-[0.71875rem] leading-relaxed" aria-label="Span path">
        <template v-for="(ancestor, index) in ancestors" :key="ancestor.spanId">
          <button
            type="button"
            class="hover:bg-hover hover:text-secondary cursor-pointer rounded border-none bg-transparent px-1.5 py-0.5 transition-colors duration-120 ease-out"
            @click="$emit('selectAncestor', ancestor.spanId)"
          >
            {{ ancestor.name }}
          </button>
          <span v-if="index < ancestors.length - 1" class="text-strong" aria-hidden="true">›</span>
        </template>
      </nav>
    </div>

    <div class="border-subtle flex shrink-0 gap-0.5 border-b px-5" role="tablist" aria-label="Span detail views">
      <button
        v-for="tab in TABS"
        :key="tab.value"
        type="button"
        role="tab"
        :aria-selected="activeTab === tab.value"
        :class="
          cn(
            'text-tertiary hover:text-secondary mr-4 cursor-pointer border-x-0 border-t-0 border-b-2 border-b-transparent bg-transparent px-1 py-2.5 text-[0.78125rem] leading-snug transition-colors duration-120 ease-out',
            activeTab === tab.value && 'text-heading border-b-signal font-medium',
          )
        "
        @click="activeTab = tab.value"
      >
        {{ tab.label }}
      </button>
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto px-4 py-3" role="tabpanel">
      <template v-if="activeTab === 'overview'">
        <div class="mb-3.5 grid grid-cols-2 gap-2">
          <div :class="cn('border-subtle bg-raised rounded border px-2.5 py-2')">
            <div class="text-tertiary mb-1 text-[0.65625rem] font-medium tracking-wider uppercase">Duration</div>
            <div class="text-heading text-[0.84375rem] leading-snug font-semibold tabular-nums">
              {{ formatDurationMs(durationMs) }}
            </div>
          </div>
          <div class="border-subtle bg-raised rounded border px-2.5 py-2">
            <div class="text-tertiary mb-1 text-[0.65625rem] font-medium tracking-wider uppercase">Self time</div>
            <div class="text-heading text-[0.84375rem] leading-snug font-semibold tabular-nums">
              {{ formatDurationMs(selfTimeMs) }}
            </div>
          </div>
          <div class="border-subtle bg-raised rounded border px-2.5 py-2">
            <div class="text-tertiary mb-1 text-[0.65625rem] font-medium tracking-wider uppercase">% of trace</div>
            <div class="text-heading text-[0.84375rem] leading-snug font-semibold tabular-nums">
              {{ percentOfTrace }}%
            </div>
          </div>
          <div class="border-subtle bg-raised rounded border px-2.5 py-2">
            <div class="text-tertiary mb-1 text-[0.65625rem] font-medium tracking-wider uppercase">Status</div>
            <div
              :class="
                cn(
                  'text-[0.84375rem] leading-snug font-semibold',
                  span.status === 'error' ? 'text-error' : 'text-heading',
                )
              "
            >
              {{ statusLabel }}
            </div>
          </div>
        </div>

        <p class="text-tertiary mt-0 mb-2 text-[0.65625rem] font-medium tracking-wider uppercase">
          Position in trace
        </p>
        <div class="bg-canvas relative mb-4 h-1.5 overflow-hidden rounded-full">
          <div
            :class="cn('absolute top-0 bottom-0 rounded-full', span.status === 'error' ? 'bg-error' : 'bg-ok')"
            :style="{ left: `${positionStartPct}%`, width: `${positionWidthPct}%` }"
          />
        </div>

        <p class="text-tertiary mt-0 mb-2 text-[0.65625rem] font-medium tracking-wider uppercase">Identity</p>
        <table class="w-full border-collapse">
          <tbody>
            <tr class="border-subtle border-b last:border-b-0">
              <td class="text-tertiary w-32 py-2 pr-3 font-mono text-xs break-words">span.id</td>
              <td class="text-primary hover:text-heading cursor-pointer rounded-sm py-2 font-mono text-xs break-all tabular-nums" title="Select to copy">
                {{ span.spanId }}
              </td>
            </tr>
            <tr class="border-subtle border-b last:border-b-0">
              <td class="text-tertiary w-32 py-2 pr-3 font-mono text-xs break-words">start offset</td>
              <td class="text-primary py-2 font-mono text-xs break-all tabular-nums">
                {{ formatDurationMs(startOffsetMs) }}
              </td>
            </tr>
            <tr class="border-subtle border-b last:border-b-0">
              <td class="text-tertiary w-32 py-2 pr-3 font-mono text-xs break-words">async</td>
              <td class="text-primary py-2 font-mono text-xs break-all tabular-nums">
                {{ isAsync ? "true" : "false" }}
              </td>
            </tr>
            <tr class="border-subtle border-b last:border-b-0">
              <td class="text-tertiary w-32 py-2 pr-3 font-mono text-xs break-words">children</td>
              <td class="text-primary py-2 font-mono text-xs break-all tabular-nums">{{ childCount }}</td>
            </tr>
          </tbody>
        </table>
      </template>

      <template v-else-if="activeTab === 'attributes'">
        <SpanAttributes :attributes="span.attributes" />
      </template>

      <template v-else-if="activeTab === 'events'">
        <div v-if="events.length === 0" class="px-2 py-6 text-center">
          <div class="text-primary mb-1 text-[0.78125rem] font-semibold">No events on this span</div>
          <div class="text-tertiary text-xs leading-relaxed text-pretty">
            Events mark exceptions and milestones during the span.
          </div>
        </div>
        <div v-for="(event, index) in events" :key="index" class="mb-3.5 flex gap-2.5">
          <div
            :class="cn('mt-1.5 size-2 shrink-0 rounded-full', event.type === 'error' ? 'bg-error' : 'bg-info')"
            aria-hidden="true"
          />
          <div class="min-w-0">
            <div class="text-tertiary text-[0.6875rem] tabular-nums">+{{ formatDurationMs(event.offsetMs) }}</div>
            <div class="text-primary text-[0.78125rem] leading-snug font-medium">{{ event.name }}</div>
            <div
              v-if="event.detail"
              class="text-secondary border-subtle bg-raised mt-1.5 rounded border px-2.5 py-2 font-mono text-[0.71875rem] leading-relaxed break-words whitespace-pre-wrap"
            >
              {{ event.detail }}
            </div>
          </div>
        </div>
      </template>

      <template v-else>
        <div v-if="logsPending" class="px-2 py-6 text-center" role="status">
          <div class="text-heading text-[0.78125rem] font-semibold">Loading logs…</div>
        </div>
        <div v-else-if="logsError" class="px-2 py-6 text-center" role="alert">
          <div class="text-heading mb-1 text-[0.78125rem] font-semibold">Couldn’t load logs</div>
          <button
            type="button"
            class="text-signal cursor-pointer border-none bg-transparent p-0 text-xs hover:underline hover:underline-offset-2"
            @click="$emit('retryLogs')"
          >
            Retry
          </button>
        </div>
        <div v-else-if="logs.length === 0" class="px-2 py-6 text-center">
          <div class="text-primary mb-1 text-[0.78125rem] font-semibold">No logs in this span’s window</div>
          <div class="text-tertiary text-xs leading-relaxed text-pretty">
            Logs from this service and time range will appear here.
          </div>
        </div>
        <div
          v-for="log in logs"
          :key="log.id"
          class="border-subtle flex gap-2 border-b py-2 text-[0.71875rem] leading-relaxed last:border-b-0"
        >
          <div
            :class="
              cn(
                'mt-1 size-[0.4375rem] shrink-0 rounded-full',
                log.level === 'error' && 'bg-error',
                log.level === 'warn' && 'bg-warn',
                log.level === 'info' && 'bg-info',
              )
            "
            aria-hidden="true"
          />
          <div class="text-tertiary shrink-0 tabular-nums">+{{ formatDurationMs(logOffsetMs(log.timestamp)) }}</div>
          <div class="text-secondary min-w-0 break-words">{{ log.message }}</div>
        </div>
      </template>
    </div>

    <div class="border-subtle text-tertiary flex shrink-0 flex-wrap gap-4 border-t px-4 py-2 text-[0.6875rem] leading-relaxed max-[47.5rem]:hidden" aria-hidden="true">
      <span><kbd class="bg-raised border-default rounded border px-1.5 py-px font-sans">j</kbd><kbd class="bg-raised border-default rounded border px-1.5 py-px font-sans">k</kbd> move</span>
      <span><kbd class="bg-raised border-default rounded border px-1.5 py-px font-sans">Enter</kbd> expand</span>
      <span><kbd class="bg-raised border-default rounded border px-1.5 py-px font-sans">Esc</kbd> close</span>
    </div>
  </div>
</template>
