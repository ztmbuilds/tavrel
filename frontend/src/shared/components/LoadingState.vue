<script setup lang="ts">
import LoadingSpinner from "@/shared/components/LoadingSpinner.vue";

defineProps<{
  label: string;
}>();
</script>

<template>
  <div class="flex flex-col items-center px-5 py-12">
    <LoadingSpinner size="md" :label="label" />
  </div>
</template>
