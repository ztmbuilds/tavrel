<script setup lang="ts">
import EmptyState from "@/shared/components/EmptyState.vue";
import PageHead from "@/shared/components/PageHead.vue";
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <PageHead title="Dashboard" />
    <EmptyState
      title="No dashboard data"
      message="Statistics endpoints land with the dashboard slice."
    />
  </div>
</template>
