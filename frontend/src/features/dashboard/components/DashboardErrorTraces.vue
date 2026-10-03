<script setup lang="ts">
import { RouterLink } from "vue-router";

import ArrowLink from "@/shared/components/ArrowLink.vue";
import type { DashboardRange } from "@/features/dashboard/types";
import type { Trace } from "@/shared/types/entities/trace";
import { formatRelativeTime } from "@/shared/utils/time";

defineProps<{
  traces: Trace[];
  range: DashboardRange;
}>();
</script>

<template>
  <section class="border-subtle bg-panel overflow-hidden rounded-xl border">
    <header
      class="border-subtle flex items-center justify-between gap-3 border-b px-4 py-3"
    >
      <h2 class="text-heading text-sm font-semibold">Failed traces</h2>
      <ArrowLink
        :to="{ name: 'traces', query: { status: 'error', range } }"
        route-name="traces"
        class="shrink-0"
      >
        View all
      </ArrowLink>
    </header>
    <ul v-if="traces.length > 0" class="m-0 list-none p-0">
      <li
        v-for="trace in traces"
        :key="trace.traceId"
        class="border-subtle hover:bg-hover border-b last:border-b-0"
      >
        <RouterLink
          :to="{ name: 'trace-detail', params: { id: trace.traceId } }"
          class="flex min-w-0 items-center justify-between gap-3 px-4 py-2.5"
        >
          <span class="min-w-0">
            <span
              class="text-primary block truncate text-xs font-medium"
              :title="trace.rootName"
            >
              {{ trace.rootName || "Unnamed operation" }}
            </span>
            <span
              class="text-tertiary mt-1 block truncate text-[0.6875rem]"
              :title="trace.rootService"
            >
              {{ trace.rootService || "Unknown service" }}
            </span>
          </span>
          <time
            class="text-tertiary shrink-0 text-[0.6875rem] tabular-nums"
            :datetime="trace.startTime"
          >
            {{ formatRelativeTime(trace.startTime) }}
          </time>
        </RouterLink>
      </li>
    </ul>
    <p v-else class="text-tertiary px-4 py-6 text-sm">
      No failed traces in this window.
    </p>
  </section>
</template>
