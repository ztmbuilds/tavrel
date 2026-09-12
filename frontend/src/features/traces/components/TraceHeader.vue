<script setup lang="ts">
import { computed } from "vue";

import chevronLeftIcon from "@/assets/icons/chevron-left.svg";
import type { TraceStatus } from "@/shared/types/entities/trace";
import { cn } from "@/shared/utils/cn";
import { serviceColor } from "@/shared/utils/service-color";
import { formatDurationMs } from "@/shared/utils/time";

import type { TraceDiagnosis, TraceIssue } from "../types";

const props = defineProps<{
  shortTraceId: string;
  rootName: string;
  rootService: string;
  rootKind: string;
  status: TraceStatus;
  durationLabel: string;
  relativeLabel: string;
  spanCount: number;
  copied: boolean;
  legend: { name: string; color: string }[];
  issues: TraceIssue[];
  diagnosis: TraceDiagnosis;
}>();

defineEmits<{
  back: [];
  copy: [];
  select: [spanId: string];
}>();

const statusLabel = computed(() =>
  props.status === "error" ? "Error" : "OK",
);
</script>

<template>
  <header
    class="border-subtle bg-panel flex h-9.5 shrink-0 items-center gap-2.5 border-b px-5"
  >
    <button
      type="button"
      class="text-tertiary hover:bg-hover hover:text-secondary -ml-2 flex cursor-pointer items-center gap-1.5 rounded-md border-none bg-transparent px-2 py-1.5 text-[0.78125rem] transition-colors duration-120 ease-out"
      @click="$emit('back')"
    >
      <img :src="chevronLeftIcon" alt="" aria-hidden="true" class="size-3.5" />
      Traces
    </button>
    <div class="bg-subtle h-3.5 w-px shrink-0" aria-hidden="true" />
    <button
      type="button"
      class="text-tertiary hover:text-secondary hover:bg-hover -ml-1.5 cursor-pointer rounded border-none bg-transparent px-1.5 py-0.5 font-mono text-[0.71875rem] underline decoration-strong decoration-dotted underline-offset-4 transition-colors duration-120 ease-out"
      title="Copy full trace ID"
      @click="$emit('copy')"
    >
      trace&nbsp;{{ shortTraceId }}
    </button>
    <span v-if="copied" class="text-ok text-[0.6875rem] font-medium" role="status"
      >Copied</span
    >
    <div class="flex-1" />
    <span
      class="text-tertiary text-xs tabular-nums max-[47.5rem]:hidden"
      >{{ relativeLabel }}</span
    >
  </header>

  <section class="border-subtle bg-panel shrink-0 border-b px-5 pt-4.5 pb-3.5" aria-labelledby="trace-title">
    <div class="flex flex-wrap items-end justify-between gap-6">
      <div class="min-w-0">
        <p class="text-tertiary mb-0.5 flex items-center gap-1.5 text-xs">
          <span
            class="size-1.5 shrink-0 rounded-full"
            :style="{ background: serviceColor(rootService) }"
            aria-hidden="true"
          />
          {{ rootService }} · {{ rootKind }}
        </p>
        <div class="flex flex-wrap items-center gap-2">
          <h1
            id="trace-title"
            class="text-heading m-0 text-[1.3125rem] leading-snug font-semibold tracking-[-0.01em]"
          >
            {{ rootName }}
          </h1>
          <span
            :class="
              cn(
                'ml-2.5 inline-flex items-center gap-1.5 rounded-full py-0.5 pr-2.5 pl-2 text-[0.71875rem] font-semibold',
                status === 'error'
                  ? 'text-error bg-error/10'
                  : 'text-ok bg-ok/10',
              )
            "
          >
            <span class="size-1.5 rounded-full bg-current" aria-hidden="true" />
            {{ statusLabel }}
          </span>
        </div>
      </div>
      <div class="shrink-0 text-right">
        <span
          class="text-heading text-[2rem] leading-none font-semibold tracking-[-0.02em] tabular-nums"
          >{{ durationLabel }}</span
        >
        <span class="text-tertiary mt-0.5 block text-[0.6875rem]"
          >total duration</span
        >
      </div>
    </div>

    <p
      v-if="diagnosis.kind === 'error' && diagnosis.spanId !== null"
      class="text-secondary mt-3 flex max-w-[46.25rem] items-start gap-2 text-[0.78125rem] leading-relaxed text-pretty"
    >
      <span class="bg-error mt-1.5 size-1.5 shrink-0 rounded-full" aria-hidden="true" />
      <span>
        <button
          type="button"
          class="text-heading cursor-pointer border-none bg-transparent p-0 font-semibold underline decoration-strong decoration-dotted underline-offset-4 hover:text-signal-hover"
          @click="$emit('select', diagnosis.spanId)"
        >
          {{ diagnosis.spanName }}</button
        >{{ diagnosis.suffix }}{{ diagnosis.tail }}
      </span>
    </p>
    <p
      v-else
      class="text-secondary mt-3 flex max-w-[46.25rem] items-start gap-2 text-[0.78125rem] leading-relaxed text-pretty"
    >
      <span class="bg-ok mt-1.5 size-1.5 shrink-0 rounded-full" aria-hidden="true" />
      <span>Every span in this trace completed without errors.</span>
    </p>

    <div v-if="issues.length > 0" class="mt-2.5 ml-4.25 flex flex-wrap gap-1.5">
      <button
        v-for="issue in issues"
        :key="issue.spanId"
        type="button"
        class="text-secondary border-default bg-surface hover:border-strong hover:bg-hover hover:text-heading inline-flex cursor-pointer items-center gap-1.5 rounded-full border py-1 pr-2.5 pl-2 text-[0.71875rem] transition-colors duration-120 ease-out"
        @click="$emit('select', issue.spanId)"
      >
        <span
          :class="
            cn(
              'size-1.5 shrink-0 rounded-full',
              issue.status === 'error' ? 'bg-error' : 'bg-ok',
            )
          "
          aria-hidden="true"
        />
        {{ issue.name }}
        <span class="text-tertiary tabular-nums">{{ formatDurationMs(issue.durationMs) }}</span>
      </button>
    </div>

    <div
      class="text-tertiary mt-3 ml-4.25 flex flex-wrap items-center gap-x-3 gap-y-1.5 text-xs"
    >
      <span
        v-for="entry in legend"
        :key="entry.name"
        class="inline-flex items-center gap-1.5"
      >
        <span
          class="size-1.5 rounded-full"
          :style="{ background: entry.color }"
          aria-hidden="true"
        />{{ entry.name }}
      </span>
      <span class="text-strong" aria-hidden="true">·</span>
      <span>{{ spanCount }} spans</span>
    </div>
  </section>
</template>
