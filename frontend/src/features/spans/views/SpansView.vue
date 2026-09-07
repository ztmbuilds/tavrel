<script setup lang="ts">
import EmptyState from "@/shared/components/EmptyState.vue";
import PageHead from "@/shared/components/PageHead.vue";
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <PageHead title="Spans" />
    <EmptyState
      title="No spans found"
      message="No spans match these filters. Try widening the time range or clearing a filter."
      action-label="Clear all filters"
    />
  </div>
</template>
