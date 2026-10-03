<script setup lang="ts">
import { RouterLink } from "vue-router";

import ArrowLink from "@/shared/components/ArrowLink.vue";
import type { DashboardRange } from "@/features/dashboard/types";
import type { Log } from "@/shared/types/entities/log";
import { formatRelativeTime } from "@/shared/utils/time";

defineProps<{
  logs: Log[];
  range: DashboardRange;
}>();
</script>

<template>
  <section class="border-subtle bg-panel overflow-hidden rounded-xl border">
    <header
      class="border-subtle flex items-center justify-between gap-3 border-b px-4 py-3"
    >
      <h2 class="text-heading text-sm font-semibold">Error logs</h2>
      <ArrowLink
        :to="{ name: 'logs', query: { level: 'error', range } }"
        route-name="logs"
        class="shrink-0"
      >
        View all
      </ArrowLink>
    </header>
    <ul v-if="logs.length > 0" class="m-0 list-none p-0">
      <li
        v-for="log in logs"
        :key="log.id"
        class="border-subtle flex min-w-0 items-start gap-3 border-b px-4 py-2.5 last:border-b-0"
      >
        <span
          class="bg-error mt-1.5 size-1.5 shrink-0 rounded-full"
          aria-hidden="true"
        />
        <div class="min-w-0 flex-1">
          <p
            class="text-primary m-0 truncate text-xs leading-relaxed"
            :title="log.message"
          >
            {{ log.message || "Empty error log" }}
          </p>
          <div
            class="text-tertiary mt-1 flex min-w-0 items-center gap-2 text-[0.6875rem]"
          >
            <span class="min-w-0 truncate" :title="log.service">{{
              log.service || "Unknown service"
            }}</span>
            <span aria-hidden="true">·</span>
            <time class="shrink-0 tabular-nums" :datetime="log.timestamp">
              {{ formatRelativeTime(log.timestamp) }}
            </time>
            <RouterLink
              v-if="log.traceId !== ''"
              :to="{ name: 'trace-detail', params: { id: log.traceId } }"
              class="text-signal ml-auto shrink-0 hover:underline hover:underline-offset-2"
            >
              Open trace
            </RouterLink>
          </div>
        </div>
      </li>
    </ul>
    <p v-else class="text-tertiary px-4 py-6 text-sm">
      No error logs in this window.
    </p>
  </section>
</template>
