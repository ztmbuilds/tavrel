<script setup lang="ts">
import { computed } from "vue";
import type { RouteLocationRaw } from "vue-router";

import checkOkIcon from "@/assets/icons/check-ok.svg?component";
import chevronLeftIcon from "@/assets/icons/chevron-left.svg?component";
import copyIcon from "@/assets/icons/copy.svg?component";
import ArrowLink from "@/shared/components/ArrowLink.vue";
import Badge from "@/shared/components/ui/Badge.vue";
import Button from "@/shared/components/ui/Button.vue";
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
  environment: string;
  status: TraceStatus;
  durationLabel: string;
  relativeLabel: string;
  spanCount: number;
  logsTo: RouteLocationRaw;
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

const statusLabel = computed(() => (props.status === "error" ? "Error" : "OK"));
</script>

<template>
  <header
    class="border-subtle bg-panel flex h-9.5 shrink-0 items-center gap-2.5 border-b px-10 max-[40rem]:px-4"
  >
    <Button
      variant="ghost"
      size="sm"
      class="-ml-2 h-7 gap-1.5 px-2 text-[0.78125rem] [&_svg]:size-3.5"
      @click="$emit('back')"
    >
      <component :is="chevronLeftIcon" aria-hidden="true" />
      Traces
    </Button>
    <div class="bg-subtle h-3.5 w-px shrink-0" aria-hidden="true" />
    <Button
      variant="ghost"
      size="sm"
      class="group decoration-strong -ml-1.5 h-6 gap-1 px-1.5 font-mono text-[0.71875rem] underline decoration-dotted underline-offset-4 [&_svg]:size-3"
      title="Copy full trace ID"
      @click="$emit('copy')"
    >
      <span>trace&nbsp;{{ shortTraceId }}</span>
      <component
        :is="copied ? checkOkIcon : copyIcon"
        aria-hidden="true"
        :class="
          cn(
            'transition-opacity duration-120 ease-out',
            copied
              ? 'text-ok opacity-100'
              : 'opacity-0 group-hover:opacity-100 group-focus-visible:opacity-100',
          )
        "
      />
    </Button>
    <span v-if="copied" class="sr-only" role="status">Copied</span>
    <div class="flex-1" />
    <span class="text-tertiary text-xs tabular-nums max-[47.5rem]:hidden">{{
      relativeLabel
    }}</span>
  </header>

  <section
    class="border-subtle bg-panel shrink-0 border-b px-10 pt-7 pb-5 max-[40rem]:px-4 max-[40rem]:pt-5 max-[40rem]:pb-4"
    aria-labelledby="trace-title"
  >
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
            class="text-heading m-0 text-[1.875rem] leading-[1.2] font-semibold tracking-[-0.02em]"
          >
            {{ rootName }}
          </h1>
          <Badge
            :variant="status === 'error' ? 'error' : 'ok'"
            dot
            class="ml-1"
          >
            {{ statusLabel }}
          </Badge>
          <Badge
            v-if="environment !== ''"
            variant="outline"
            :title="`Environment: ${environment}`"
          >
            {{ environment }}
          </Badge>
        </div>
      </div>
      <div class="shrink-0 text-right">
        <span
          class="text-heading text-[2.75rem] leading-none font-semibold tracking-[-0.03em] tabular-nums"
          >{{ durationLabel }}</span
        >
        <span class="text-tertiary mt-1 block text-[0.8125rem]"
          >total duration</span
        >
      </div>
    </div>

    <p
      v-if="diagnosis.kind === 'error' && diagnosis.spanId !== null"
      class="text-secondary mt-3 flex max-w-[46.25rem] items-start gap-2 text-[0.78125rem] leading-relaxed text-pretty"
    >
      <span
        class="bg-error mt-1.5 size-1.5 shrink-0 rounded-full"
        aria-hidden="true"
      />
      <span>
        {{ diagnosis.headline }} —
        <button
          type="button"
          class="text-heading decoration-strong hover:text-signal-hover cursor-pointer border-none bg-transparent p-0 font-semibold underline decoration-dotted underline-offset-4"
          @click="$emit('select', diagnosis.spanId)"
        >
          {{ diagnosis.service }} · {{ diagnosis.spanName }}</button
        >{{ diagnosis.tail }}
      </span>
    </p>
    <p
      v-else
      class="text-secondary mt-3 flex max-w-[46.25rem] items-start gap-2 text-[0.78125rem] leading-relaxed text-pretty"
    >
      <span
        class="bg-ok mt-1.5 size-1.5 shrink-0 rounded-full"
        aria-hidden="true"
      />
      <span>Every span in this trace completed without errors.</span>
    </p>

    <div v-if="issues.length > 0" class="mt-2.5 ml-4.25 flex flex-wrap gap-1.5">
      <Button
        v-for="issue in issues"
        :key="issue.spanId"
        variant="ghost"
        size="sm"
        class="border-subtle hover:border-default h-6 rounded-full border px-2 text-[0.71875rem]"
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
        <span class="text-tertiary tabular-nums">{{
          formatDurationMs(issue.durationMs)
        }}</span>
      </Button>
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
      <span class="text-strong" aria-hidden="true">·</span>
      <ArrowLink
        :to="logsTo"
        route-name="logs"
        class="text-tertiary hover:text-heading"
      >
        View logs in this time range
      </ArrowLink>
    </div>
  </section>
</template>
