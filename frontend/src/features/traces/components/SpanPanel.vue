<script setup lang="ts">
import xIcon from "@/assets/icons/x.svg?component";
import Badge from "@/shared/components/ui/Badge.vue";
import Button from "@/shared/components/ui/Button.vue";
import Tabs from "@/shared/components/ui/Tabs.vue";
import LoadingSpinner from "@/shared/components/LoadingSpinner.vue";
import type { Log } from "@/shared/types/entities/log";
import type { Span } from "@/shared/types/entities/span";
import type { SpanLink } from "@/shared/types/entities/span-link";
import { cn } from "@/shared/utils/cn";
import { serviceColor } from "@/shared/utils/service-color";
import { diffMsPrecise, formatDurationMs } from "@/shared/utils/time";

import SpanAttributes from "./SpanAttributes.vue";
import SpanLinkRow from "./SpanLinkRow.vue";

import type { SpanEventView, TraceDetailTab } from "../types";

const props = defineProps<{
  span: Span;
  childCount: number;
  selfTimeMs: number;
  percentOfTrace: number;
  positionStartPct: number;
  positionWidthPct: number;
  startOffsetMs: number;
  durationMs: number;
  isAsync: boolean;
  events: SpanEventView[];
  links: SpanLink[];
  linksPending: boolean;
  linksError: boolean;
  logs: Log[];
  logsPending: boolean;
  logsError: boolean;
}>();

defineEmits<{
  close: [];
  openLink: [link: SpanLink];
  retryLinks: [];
  retryLogs: [];
}>();

const activeTab = defineModel<TraceDetailTab>("activeTab", {
  default: "overview",
});

const TABS: readonly { value: TraceDetailTab; label: string }[] = [
  { value: "overview", label: "Overview" },
  { value: "attributes", label: "Attributes" },
  { value: "events", label: "Events" },
  { value: "links", label: "Links" },
  { value: "logs", label: "Logs" },
];

function logOffsetMs(timestamp: string): number {
  const offset = diffMsPrecise(props.span.startTime, timestamp);
  if (Number.isNaN(offset)) return 0;
  return Math.max(offset, 0);
}
</script>

<template>
  <div class="flex h-full w-full min-w-0 flex-col overflow-hidden">
    <div class="border-subtle shrink-0 border-b px-4 pt-3 pb-2.5">
      <div class="flex items-start justify-between gap-2.5">
        <p
          class="text-heading m-0 text-[0.90625rem] leading-snug font-semibold text-pretty"
        >
          {{ span.name }}
        </p>
        <Button
          variant="ghost"
          size="sm"
          class="size-7 shrink-0 p-0 [&_svg]:size-3.5"
          aria-label="Close span details"
          @click="$emit('close')"
        >
          <component :is="xIcon" aria-hidden="true" />
        </Button>
      </div>
      <div
        class="text-secondary mt-1.5 flex items-center gap-1.5 text-xs leading-relaxed"
      >
        <span
          class="size-[0.4375rem] rounded-full"
          :style="{ background: serviceColor(span.service) }"
          aria-hidden="true"
        />
        {{ span.service }} · {{ span.kind }}
        <Badge
          v-if="span.environment !== ''"
          variant="outline"
          class="shrink-0 px-2 py-px text-[0.6875rem]"
          :title="`Environment: ${span.environment}`"
        >
          {{ span.environment }}
        </Badge>
        <span class="sr-only">, status {{ span.status }}</span>
      </div>
    </div>

    <Tabs
      v-model="activeTab"
      :items="TABS"
      aria-label="Span detail views"
      class="flex min-h-0 flex-1 flex-col"
      list-class="shrink-0 gap-5 overflow-x-auto px-5 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
      content-class="min-h-0 flex-1 overflow-y-auto px-4 py-3"
    >
      <template #overview>
        <div class="mb-3.5 grid grid-cols-2 gap-2">
          <div
            :class="cn('border-subtle bg-raised rounded border px-2.5 py-2')"
          >
            <div
              class="text-tertiary mb-1 text-[0.65625rem] font-medium tracking-wider uppercase"
            >
              Duration
            </div>
            <div
              class="text-heading text-[0.84375rem] leading-snug font-semibold tabular-nums"
            >
              {{ formatDurationMs(durationMs) }}
            </div>
          </div>
          <div class="border-subtle bg-raised rounded border px-2.5 py-2">
            <div
              class="text-tertiary mb-1 text-[0.65625rem] font-medium tracking-wider uppercase"
            >
              Self time
            </div>
            <div
              class="text-heading text-[0.84375rem] leading-snug font-semibold tabular-nums"
            >
              {{ formatDurationMs(selfTimeMs) }}
            </div>
          </div>
          <div class="border-subtle bg-raised rounded border px-2.5 py-2">
            <div
              class="text-tertiary mb-1 text-[0.65625rem] font-medium tracking-wider uppercase"
            >
              % of trace
            </div>
            <div
              class="text-heading text-[0.84375rem] leading-snug font-semibold tabular-nums"
            >
              {{ percentOfTrace }}%
            </div>
          </div>
          <div class="border-subtle bg-raised rounded border px-2.5 py-2">
            <div
              class="text-tertiary mb-1 text-[0.65625rem] font-medium tracking-wider uppercase"
            >
              Status
            </div>
            <div
              :class="
                cn(
                  'text-[0.84375rem] leading-snug font-semibold',
                  span.status === 'error' ? 'text-error' : 'text-heading',
                )
              "
            >
              {{ span.status }}
            </div>
          </div>
        </div>

        <p
          class="text-tertiary mt-0 mb-2 text-[0.65625rem] font-medium tracking-wider uppercase"
        >
          Position in trace
        </p>
        <div class="bg-canvas relative mb-4 h-1.5 overflow-hidden rounded-full">
          <div
            :class="
              cn(
                'absolute top-0 bottom-0 rounded-full',
                span.status === 'error' ? 'bg-error' : 'bg-ok',
              )
            "
            :style="{
              left: `${positionStartPct}%`,
              width: `${positionWidthPct}%`,
            }"
          />
        </div>

        <p
          class="text-tertiary mt-0 mb-2 text-[0.65625rem] font-medium tracking-wider uppercase"
        >
          Identity
        </p>
        <table class="w-full border-collapse">
          <tbody>
            <tr class="border-subtle border-b last:border-b-0">
              <td
                class="text-tertiary w-32 py-2 pr-3 font-mono text-xs break-words"
              >
                span.id
              </td>
              <td
                class="text-primary hover:text-heading cursor-pointer rounded-sm py-2 font-mono text-xs break-all tabular-nums"
                title="Select to copy"
              >
                {{ span.spanId }}
              </td>
            </tr>
            <tr class="border-subtle border-b last:border-b-0">
              <td
                class="text-tertiary w-32 py-2 pr-3 font-mono text-xs break-words"
              >
                start offset
              </td>
              <td
                class="text-primary py-2 font-mono text-xs break-all tabular-nums"
              >
                {{ formatDurationMs(startOffsetMs) }}
              </td>
            </tr>
            <tr class="border-subtle border-b last:border-b-0">
              <td
                class="text-tertiary w-32 py-2 pr-3 font-mono text-xs break-words"
              >
                async
              </td>
              <td
                class="text-primary py-2 font-mono text-xs break-all tabular-nums"
              >
                {{ isAsync ? "true" : "false" }}
              </td>
            </tr>
            <tr class="border-subtle border-b last:border-b-0">
              <td
                class="text-tertiary w-32 py-2 pr-3 font-mono text-xs break-words"
              >
                children
              </td>
              <td
                class="text-primary py-2 font-mono text-xs break-all tabular-nums"
              >
                {{ childCount }}
              </td>
            </tr>
          </tbody>
        </table>
      </template>

      <template #attributes>
        <SpanAttributes :attributes="span.attributes" />
      </template>

      <template #events>
        <div v-if="events.length === 0" class="px-2 py-6 text-center">
          <div class="text-primary mb-1 text-[0.78125rem] font-semibold">
            No events on this span
          </div>
          <div class="text-tertiary text-xs leading-relaxed text-pretty">
            Events mark exceptions and milestones during the span.
          </div>
        </div>
        <div
          v-for="(event, index) in events"
          :key="index"
          class="mb-3.5 flex gap-2.5"
        >
          <div
            :class="
              cn(
                'mt-1.5 size-2 shrink-0 rounded-full',
                event.type === 'error' ? 'bg-error' : 'bg-info',
              )
            "
            aria-hidden="true"
          />
          <div class="min-w-0">
            <div class="text-tertiary text-[0.6875rem] tabular-nums">
              +{{ formatDurationMs(event.offsetMs) }}
            </div>
            <div
              class="text-primary text-[0.78125rem] leading-snug font-medium"
            >
              {{ event.name }}
            </div>
            <div
              v-if="event.detail"
              class="text-secondary border-subtle bg-raised mt-1.5 rounded border px-2.5 py-2 font-mono text-[0.71875rem] leading-relaxed break-words whitespace-pre-wrap"
            >
              {{ event.detail }}
            </div>
          </div>
        </div>
      </template>

      <template #links>
        <div v-if="linksPending" class="flex justify-center px-2 py-6">
          <LoadingSpinner size="sm" label="Loading links" />
        </div>
        <div v-else-if="linksError" class="px-2 py-6 text-center" role="alert">
          <div class="text-heading mb-1 text-[0.78125rem] font-semibold">
            Couldn’t load links
          </div>
          <Button size="sm" @click="$emit('retryLinks')">Retry</Button>
        </div>
        <div v-else-if="links.length === 0" class="px-2 py-6 text-center">
          <div class="text-primary mb-1 text-[0.78125rem] font-semibold">
            No links on this span
          </div>
          <div class="text-tertiary text-xs leading-relaxed text-pretty">
            Links point at related spans, like a batch producer or a trigger.
          </div>
        </div>
        <SpanLinkRow
          v-for="link in links"
          :key="`${link.traceId}:${link.spanId}`"
          :link="link"
          @open="$emit('openLink', link)"
        />
      </template>

      <template #logs>
        <div v-if="logsPending" class="flex justify-center px-2 py-6">
          <LoadingSpinner size="sm" label="Loading logs" />
        </div>
        <div v-else-if="logsError" class="px-2 py-6 text-center" role="alert">
          <div class="text-heading mb-1 text-[0.78125rem] font-semibold">
            Couldn’t load logs
          </div>
          <Button size="sm" @click="$emit('retryLogs')">Retry</Button>
        </div>
        <div v-else-if="logs.length === 0" class="px-2 py-6 text-center">
          <div class="text-primary mb-1 text-[0.78125rem] font-semibold">
            No logs in this span’s window
          </div>
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
          <div class="text-tertiary shrink-0 tabular-nums">
            +{{ formatDurationMs(logOffsetMs(log.timestamp)) }}
          </div>
          <div class="text-secondary min-w-0 break-words">
            {{ log.message }}
          </div>
        </div>
      </template>
    </Tabs>

    <div
      class="border-subtle text-tertiary flex shrink-0 flex-wrap gap-4 border-t px-4 py-2 text-[0.6875rem] leading-relaxed max-[47.5rem]:hidden"
      aria-hidden="true"
    >
      <span
        ><kbd
          class="bg-raised border-default rounded border px-1.5 py-px font-sans"
          >j</kbd
        ><kbd
          class="bg-raised border-default rounded border px-1.5 py-px font-sans"
          >k</kbd
        >
        move</span
      >
      <span
        ><kbd
          class="bg-raised border-default rounded border px-1.5 py-px font-sans"
          >Enter</kbd
        >
        expand</span
      >
      <span
        ><kbd
          class="bg-raised border-default rounded border px-1.5 py-px font-sans"
          >Esc</kbd
        >
        close</span
      >
    </div>
  </div>
</template>
