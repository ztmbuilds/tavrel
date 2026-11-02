<script setup lang="ts">
defineProps<{
  guides: boolean[];
  depth: number;
  isLast: boolean;
}>();
</script>

<template>
  <div class="flex h-full shrink-0 items-stretch" aria-hidden="true">
    <div
      v-for="(guide, index) in guides"
      :key="index"
      class="relative h-full w-4 shrink-0"
    >
      <div
        v-if="guide"
        class="bg-default absolute top-0 bottom-0 left-2 w-px"
      />
    </div>
    <div v-if="depth > 1" class="relative h-full w-4 shrink-0">
      <div class="bg-default absolute top-0 left-2 h-1/2 w-px" />
      <div
        v-if="!isLast"
        class="bg-default absolute top-1/2 bottom-0 left-2 w-px"
      />
      <div class="bg-default absolute top-1/2 right-0 left-2 h-px" />
    </div>
  </div>
</template>
