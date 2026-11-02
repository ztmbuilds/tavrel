<script setup lang="ts">
import { RouterLink } from "vue-router";

import ArrowLink from "@/shared/components/ArrowLink.vue";
import Card from "@/shared/components/ui/Card.vue";
import CardRow from "@/shared/components/ui/CardRow.vue";
import type { DashboardRange } from "@/features/dashboard/types";
import type { Trace } from "@/shared/types/entities/trace";
import { formatRelativeTime } from "@/shared/utils/time";

defineProps<{
  traces: Trace[];
  range: DashboardRange;
}>();
</script>

<template>
  <Card title="Failed traces" flush>
    <template #action>
      <ArrowLink
        :to="{ name: 'traces', query: { status: 'error', range } }"
        route-name="traces"
        class="shrink-0"
      >
        View all
      </ArrowLink>
    </template>
    <ul v-if="traces.length > 0" class="m-0 list-none p-0">
      <li v-for="trace in traces" :key="trace.traceId">
        <CardRow as-child>
          <RouterLink
            :to="{ name: 'trace-detail', params: { id: trace.traceId } }"
            class="flex w-full min-w-0 items-center justify-between gap-3"
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
        </CardRow>
      </li>
    </ul>
    <p v-else class="text-tertiary px-5 py-14 text-center text-[0.8125rem]">
      No failed traces in this window.
    </p>
  </Card>
</template>
