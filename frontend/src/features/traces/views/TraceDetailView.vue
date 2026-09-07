<script setup lang="ts">
import EmptyState from "@/shared/components/EmptyState.vue";
import PageHead from "@/shared/components/PageHead.vue";
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <PageHead title="Trace" />
    <EmptyState
      title="No trace loaded"
      message="Trace detail lands with the trace-detail slice."
    />
  </div>
</template>
