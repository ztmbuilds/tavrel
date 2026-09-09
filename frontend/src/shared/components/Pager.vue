<script setup lang="ts">
import chevronLeftIcon from "@/assets/icons/chevron-left.svg";
import chevronRightIcon from "@/assets/icons/chevron-right.svg";

defineProps<{
  from: number;
  to: number;
  hasPrev: boolean;
  hasNext: boolean;
  itemName: string;
}>();

defineEmits<{
  prev: [];
  next: [];
}>();
</script>

<template>
  <nav
    class="border-subtle bg-panel flex shrink-0 items-center justify-end gap-2.5 border-t px-4 py-2"
    :aria-label="`${itemName} pages`"
  >
    <span class="text-tertiary text-xs tabular-nums"
      >Showing {{ from }}–{{ to }}</span
    >
    <button
      type="button"
      :disabled="!hasPrev"
      class="border-default bg-surface text-secondary hover:border-strong hover:bg-hover hover:text-heading flex size-7 cursor-pointer items-center justify-center rounded-md border transition-colors duration-120 ease-out disabled:cursor-default disabled:opacity-35"
      :aria-label="`Show newer ${itemName}`"
      @click="$emit('prev')"
    >
      <img :src="chevronLeftIcon" alt="" class="size-3.5" />
    </button>
    <button
      type="button"
      :disabled="!hasNext"
      class="border-default bg-surface text-secondary hover:border-strong hover:bg-hover hover:text-heading flex size-7 cursor-pointer items-center justify-center rounded-md border transition-colors duration-120 ease-out disabled:cursor-default disabled:opacity-35"
      :aria-label="`Show older ${itemName}`"
      @click="$emit('next')"
    >
      <img :src="chevronRightIcon" alt="" class="size-3.5" />
    </button>
  </nav>
</template>
