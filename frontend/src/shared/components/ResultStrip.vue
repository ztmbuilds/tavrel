<script setup lang="ts">
defineProps<{
  count: number;
  noun: string;
}>();
</script>

<template>
  <div
    class="border-subtle bg-panel text-tertiary shrink-0 border-b px-4 py-1.5 text-xs"
    role="status"
    aria-live="polite"
  >
    {{ count }} {{ count === 1 ? noun : `${noun}s` }} matched
  </div>
</template>
