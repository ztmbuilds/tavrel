<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import DashboardErrorLogs from "@/features/dashboard/components/DashboardErrorLogs.vue";
import DashboardErrorTraces from "@/features/dashboard/components/DashboardErrorTraces.vue";
import DashboardTraceChart from "@/features/dashboard/components/DashboardTraceChart.vue";
import { useDashboardSummaryQuery } from "@/features/dashboard/queries";
import {
  DASHBOARD_WINDOWS,
  DEFAULT_DASHBOARD_RANGE,
  type DashboardRange,
} from "@/features/dashboard/types";
import Button from "@/shared/components/ui/Button.vue";
import Select, { type SelectOption } from "@/shared/components/ui/Select.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageHead from "@/shared/components/PageHead.vue";
import { formatRelativeTime } from "@/shared/utils/time";

const route = useRoute();
const router = useRouter();

function isDashboardRange(value: unknown): value is DashboardRange {
  return (
    typeof value === "string" &&
    DASHBOARD_WINDOWS.some((window) => window.value === value)
  );
}

const initialRange = route.query["range"];
const range = ref<DashboardRange>(
  isDashboardRange(initialRange) ? initialRange : DEFAULT_DASHBOARD_RANGE,
);
const rangeOptions: SelectOption[] = DASHBOARD_WINDOWS.map((window) => ({
  value: window.value,
  label: window.label,
}));

watch(range, (value) => {
  if (route.query["range"] === value) return;
  void router.replace({ query: { range: value } });
});

watch(
  () => route.query["range"],
  (value) => {
    const next = isDashboardRange(value) ? value : DEFAULT_DASHBOARD_RANGE;
    if (range.value !== next) range.value = next;
  },
);

const { data, dataUpdatedAt, isPending, isFetching, isError, refetch } =
  useDashboardSummaryQuery(range);

const summary = computed(() => data.value ?? null);
const updatedLabel = computed(() => {
  if (dataUpdatedAt.value === 0) return "";
  return `Updated ${formatRelativeTime(new Date(dataUpdatedAt.value).toISOString())}`;
});
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <PageHead title="Overview">
      <template #actions>
        <span v-if="updatedLabel" class="text-tertiary text-xs tabular-nums">
          {{ updatedLabel }}
        </span>
        <Select
          v-model="range"
          label="Window"
          :options="rangeOptions"
          trigger-class="min-w-45"
        />
        <Button variant="secondary" :disabled="isFetching" @click="refetch()">
          Refresh
        </Button>
      </template>
    </PageHead>

    <div class="min-h-0 flex-1 overflow-y-auto px-4 pt-4 pb-5">
      <LoadingState v-if="isPending" label="Loading dashboard" />
      <div
        v-else-if="isError && summary === null"
        class="text-tertiary flex flex-col items-center gap-3.5 px-5 py-12 text-center text-[0.8125rem]"
        role="alert"
      >
        <span class="text-heading text-sm font-semibold"
          >Couldn’t load overview</span
        >
        <span class="leading-relaxed"
          >Something went wrong fetching dashboard data. Try again.</span
        >
        <Button variant="secondary" @click="refetch()">Retry</Button>
      </div>
      <div
        v-else-if="summary"
        class="mx-auto flex w-full max-w-screen-2xl flex-col gap-4"
      >
        <DashboardTraceChart
          :buckets="summary.traceBuckets"
          :overview="summary.overview"
          :range="range"
        />
        <div class="grid grid-cols-2 gap-3 max-[50rem]:grid-cols-1">
          <DashboardErrorTraces
            :traces="summary.recentErrorTraces"
            :range="range"
          />
          <DashboardErrorLogs :logs="summary.recentErrorLogs" :range="range" />
        </div>
      </div>
    </div>
  </div>
</template>
