<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import { ToggleGroupItem, ToggleGroupRoot } from "reka-ui";

import { cn } from "@/shared/utils/cn";

export interface SegmentOption {
  value: string;
  label: string;
  dot?: string;
}

const props = defineProps<{
  options: readonly SegmentOption[];
  ariaLabel?: string;
  class?: HTMLAttributes["class"];
}>();
const model = defineModel<string>({ required: true });

function onUpdate(value: unknown): void {
  if (typeof value === "string" && value) model.value = value;
}
</script>

<template>
  <ToggleGroupRoot
    type="single"
    :model-value="model"
    :aria-label="props.ariaLabel"
    :class="
      cn(
        'border-subtle bg-canvas inline-flex gap-0.5 rounded-[10px] border p-0.5',
        props.class,
      )
    "
    @update:model-value="onUpdate"
  >
    <ToggleGroupItem
      v-for="option in props.options"
      :key="option.value"
      :value="option.value"
      class="text-secondary hover:text-heading data-[state=on]:raised data-[state=on]:text-heading inline-flex h-7 cursor-pointer items-center gap-1.5 rounded-lg border border-transparent bg-transparent px-3 font-medium transition-[color]"
    >
      <span
        v-if="option.dot"
        aria-hidden="true"
        :class="cn('size-2 rounded-full', option.dot)"
      />
      {{ option.label }}
    </ToggleGroupItem>
  </ToggleGroupRoot>
</template>
