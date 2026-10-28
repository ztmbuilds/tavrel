<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import xIcon from "@/assets/icons/x.svg?component";
import { cn } from "@/shared/utils/cn";

const props = defineProps<{
  label: string;
  value: string;
  class?: HTMLAttributes["class"];
}>();
defineEmits<{ remove: [] }>();
</script>

<template>
  <span
    :class="
      cn(
        'border-subtle text-secondary inline-flex h-7 items-center gap-1.5 rounded-full border pr-1 pl-3',
        props.class,
      )
    "
  >
    {{ props.label }}: <span class="text-heading">{{ props.value }}</span>
    <button
      type="button"
      :aria-label="`Clear ${props.label.toLowerCase()} filter`"
      class="text-secondary hover:bg-hover hover:text-heading inline-grid size-5 cursor-pointer place-items-center rounded-full bg-transparent p-0"
      @click="$emit('remove')"
    >
      <component :is="xIcon" aria-hidden="true" class="size-3.5" />
    </button>
  </span>
</template>
