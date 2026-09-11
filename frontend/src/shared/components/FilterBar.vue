<script setup lang="ts">
defineProps<{
  label: string;
}>();
</script>

<template>
  <div
    class="border-subtle bg-panel mx-4 mt-3 mb-3 flex shrink-0 flex-col gap-2.5 rounded-xl border px-3.5 pt-3.5 pb-3"
    role="search"
    :aria-label="label"
  >
    <slot />
  </div>
</template>
