<script setup lang="ts">
import EmptyState from "@/shared/components/EmptyState.vue";
import PageHead from "@/shared/components/PageHead.vue";
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <PageHead title="Service map" />
    <EmptyState
      title="No services found"
      message="Service graph data lands with the services slice."
    />
  </div>
</template>
