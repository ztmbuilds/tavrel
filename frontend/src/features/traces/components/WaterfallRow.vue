<script setup lang="ts">
import { computed } from "vue";

import chevronDownIcon from "@/assets/icons/chevron-down.svg";
import { cn } from "@/shared/utils/cn";
import { serviceColor } from "@/shared/utils/service-color";
import { formatDurationMs } from "@/shared/utils/time";

import { isAsyncSpan } from "../api";
import { useWaterfallContext } from "../composables/waterfall-context";
import type { SpanNode } from "../types";
import TreeGuides from "./TreeGuides.vue";
import WaterfallRowSelf from "./WaterfallRow.vue";

const props = defineProps<{
  node: SpanNode;
  depth: number;
  isLast: boolean;
  guides: boolean[];
  level: number;
}>()

const FULL_PCT = 100;
const BAR_CLAMP_MAX_PCT = 105;
const BAR_CLAMP_MIN_PCT = -5;
const MIN_BAR_WIDTH_PCT = 0.3;
const LABEL_FIT_EDGE_PCT = 86;

const context = useWaterfallContext();

const childGuides = computed(() =>
  props.depth <= 1 ? [] : [...props.guides, !props.isLast],
);

const isCollapsed = computed(() => context.collapsed.value.has(props.node.span.spanId));
const isSelected = computed(() => context.selectedId.value === props.node.span.spanId);
const showAsync = computed(() => isAsyncSpan(props.node.span.kind));

const startMs = computed(() => context.spanStartMs(props.node.span));
const durationMs = computed(() => context.spanDurationMs(props.node.span));

const barGeom = computed(() => {
  const windowMs = context.totalMs.value;
  if (windowMs <= 0) return { left: 0, width: MIN_BAR_WIDTH_PCT };
  const left = Math.max(
    Math.min((startMs.value / windowMs) * FULL_PCT, BAR_CLAMP_MAX_PCT),
    BAR_CLAMP_MIN_PCT,
  );
  return { left, width: Math.max((durationMs.value / windowMs) * FULL_PCT, MIN_BAR_WIDTH_PCT) };
});

const barBackground = computed(() =>
  props.node.span.status === "error" ? "var(--color-error)" : serviceColor(props.node.span.service),
);

const labelFitsAfterBar = computed(
  () => barGeom.value.left + barGeom.value.width <= LABEL_FIT_EDGE_PCT,
);

const descendants = computed(() => {
  let count = 0;
  let totalMs = 0;
  const walk = (list: SpanNode[]): void => {
    for (const item of list) {
      count += 1;
      totalMs += context.spanDurationMs(item.span);
      walk(item.children);
    }
  };
  walk(props.node.children);
  return { count, totalMs };
});

const statusText = computed(() => {
  if (props.node.span.status === "error") return "Error: ";
  if (props.node.span.status === "unset") return "Unset: ";
  return "";
});

const rowLabel = computed(
  () =>
    `${statusText.value}${props.node.span.name}, ${props.node.span.service}, ${formatDurationMs(durationMs.value)}, ${props.node.span.status}`,
);

function onChevron(event: MouseEvent): void {
  event.stopPropagation();
  if (props.node.children.length > 0) context.toggleCollapse(props.node.span.spanId);
}
</script>

<template>
  <div>
    <div
      :class="
        cn(
          'hover:bg-hover flex h-[1.875rem] cursor-pointer items-center border-l-[3px] border-l-transparent transition-colors duration-120 ease-out',
          node.span.status === 'error' && 'border-l-error bg-error/10',
          isSelected && 'bg-raised shadow-[inset_3px_0_0_0_var(--color-signal)]',
        )
      "
      role="treeitem"
      :aria-level="level"
      :aria-selected="isSelected"
      :aria-expanded="node.children.length > 0 ? !isCollapsed : undefined"
      :aria-label="rowLabel"
      :data-span-id="node.span.spanId"
      tabindex="-1"
      @click="context.select(node.span.spanId)"
    >
      <div
        class="border-subtle bg-panel flex h-full shrink-0 items-center gap-1 border-r pr-2.5"
        :style="{ width: context.labelWidth.value + 'rem' }"
      >
        <TreeGuides :guides="guides" :depth="depth" :is-last="isLast" />
        <button
          :class="
            cn(
              'text-tertiary hover:bg-hover hover:text-secondary flex size-6 shrink-0 cursor-pointer items-center justify-center rounded-[5px] border-none bg-transparent p-0 transition-colors duration-120 ease-out',
              isCollapsed && '-rotate-90',
              node.children.length === 0 && 'invisible',
            )
          "
          type="button"
          :tabindex="node.children.length === 0 ? -1 : 0"
          :aria-label="`${isCollapsed ? 'Expand' : 'Collapse'} ${node.span.name}`"
          :aria-expanded="node.children.length > 0 ? !isCollapsed : undefined"
          @click="onChevron"
        >
          <img :src="chevronDownIcon" alt="" aria-hidden="true" class="size-2.5" />
        </button>
        <span
          v-if="node.span.status === 'error'"
          class="bg-error mx-[1px] size-2 shrink-0 rotate-45 scale-90 rounded-[2px]"
          aria-hidden="true"
        />
        <span
          v-else
          class="bg-strong mx-[1px] size-1.5 shrink-0 rounded-full"
          aria-hidden="true"
        />
        <span class="sr-only">{{ statusText }}</span>
        <span
          class="size-[0.4375rem] shrink-0 rounded-full"
          :style="{ background: serviceColor(node.span.service) }"
          aria-hidden="true"
        />
        <span
          :class="
            cn(
              'min-w-0 flex-1 truncate text-[0.78125rem] leading-snug',
              node.span.status === 'error' ? 'text-error font-medium' : 'text-primary',
            )
          "
          :title="node.span.name"
        >
          {{ node.span.name }}<span
            v-if="node.span.kind !== 'internal'"
            class="text-tertiary ml-1.5 text-[0.625rem]"
            >{{ node.span.kind }}</span
          ><span
            v-if="showAsync"
            class="text-tertiary border-default ml-1.5 rounded-[3px] border px-1 text-[0.625rem] font-semibold tracking-wide"
            >async</span
          >
        </span>
      </div>
      <div class="relative h-full flex-1 overflow-hidden px-4" aria-hidden="true">
        <div
          class="absolute top-1/2 h-4 -translate-y-1/2 rounded-[3px]"
          :style="{
            left: `${barGeom.left}%`,
            width: `${barGeom.width}%`,
            background: barBackground,
          }"
        >
          <span
            v-if="!labelFitsAfterBar"
            class="text-heading absolute top-1/2 right-1.5 -translate-y-1/2 text-[0.6875rem] font-medium whitespace-nowrap tabular-nums"
          >
            {{ formatDurationMs(durationMs) }}
          </span>
        </div>
        <div
          v-if="node.span.status === 'error'"
          class="bg-error absolute top-1/2 size-1.5 -translate-y-1/2 rounded-full"
          :style="{ left: `calc(${barGeom.left}% - 2px)` }"
        />
        <div
          v-if="labelFitsAfterBar"
          class="text-secondary absolute top-1/2 -translate-y-1/2 text-[0.6875rem] whitespace-nowrap tabular-nums"
          :style="{ left: `calc(${barGeom.left}% + ${barGeom.width}% + 0.375rem)` }"
        >
          {{ formatDurationMs(durationMs) }}
        </div>
      </div>
    </div>

    <div v-if="node.children.length > 0 && !isCollapsed" role="group">
      <WaterfallRowSelf
        v-for="(child, index) in node.children"
        :key="child.span.spanId"
        :node="child"
        :depth="depth + 1"
        :level="level + 1"
        :guides="childGuides"
        :is-last="index === node.children.length - 1"
      />
    </div>
    <div v-else-if="node.children.length > 0 && isCollapsed" role="group" class="flex h-[1.875rem] items-center border-l-[3px] border-l-transparent">
      <div
        class="border-subtle bg-panel flex h-full shrink-0 items-center border-r"
        :style="{ width: context.labelWidth.value + 'rem' }"
      >
        <TreeGuides :guides="childGuides" :depth="depth + 1" :is-last="true" />
        <button
          type="button"
          class="text-tertiary hover:text-secondary flex h-full w-full cursor-pointer items-center border-none bg-transparent p-0 text-[0.71875rem] tabular-nums"
          :aria-label="`Expand ${node.span.name}, ${descendants.count} spans folded`"
          @click="onChevron"
        >
          <span class="pl-1.5"
            >({{ descendants.count }} spans,
            {{ formatDurationMs(descendants.totalMs) }} folded)</span
          >
        </button>
      </div>
      <div class="flex-1" />
    </div>
  </div>
</template>
