<script setup lang="ts">
import { computed, useSlots } from "vue";

import { cn } from "@/shared/utils/cn";

defineProps<{
  title: string;
}>();

const slots = useSlots();
const hasActions = computed(() => slots.actions !== undefined);
</script>

<template>
  <div
    :class="
      cn(
        'bg-panel shrink-0 px-4 pt-3 max-[40rem]:pt-3.5',
        hasActions ? 'flex flex-wrap items-end gap-x-4 gap-y-2 pb-0' : 'pb-0.5',
      )
    "
  >
    <div :class="hasActions && 'min-w-55 flex-1'">
      <h1
        class="text-heading text-[0.9375rem] leading-tight font-semibold tracking-[-0.01em] text-balance"
      >
        {{ title }}
      </h1>
    </div>
    <div v-if="hasActions" class="flex flex-wrap items-center gap-2">
      <slot name="actions" />
    </div>
  </div>
</template>
