<script setup lang="ts">
import { RouterView } from "vue-router";

import AppSidebar from "@/shared/components/AppSidebar.vue";
</script>

<template>
  <div class="flex min-h-screen">
    <a class="skip-link" href="#main">Skip to content</a>
    <AppSidebar />
    <main id="main" tabindex="-1" class="flex min-w-0 flex-1 flex-col">
      <RouterView />
    </main>
  </div>
</template>
