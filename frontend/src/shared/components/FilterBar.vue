<script setup lang="ts">
defineProps<{
  label: string;
  showClear: boolean;
}>();

defineEmits<{
  clear: [];
}>();
</script>

<template>
  <div
    class="border-subtle bg-panel flex shrink-0 flex-wrap items-center gap-2 border-b px-4 py-2"
    role="search"
    :aria-label="label"
  >
    <slot />
    <button
      v-if="showClear"
      type="button"
      class="text-tertiary hover:bg-surface hover:text-signal ml-auto shrink-0 cursor-pointer rounded-md border-none bg-transparent px-2 py-1.5 text-xs font-medium transition-colors duration-120 ease-out"
      @click="$emit('clear')"
    >
      Clear filters
    </button>
  </div>
</template>
