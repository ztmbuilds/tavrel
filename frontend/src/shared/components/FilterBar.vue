<script setup lang="ts">
defineProps<{
  label: string;
}>();
</script>

<template>
  <div
    class="border-subtle bg-panel mx-10 mb-4 flex shrink-0 flex-col gap-2.5 rounded-xl border p-4 max-[40rem]:mx-4"
    role="search"
    :aria-label="label"
  >
    <slot />
  </div>
</template>
