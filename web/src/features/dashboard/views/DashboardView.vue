<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";

import arrowRightIcon from "@/assets/icons/arrow-right.svg?component";
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
import SpotlightCard from "@/shared/components/ui/SpotlightCard.vue";
import StatCard from "@/shared/components/ui/StatCard.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageHead from "@/shared/components/PageHead.vue";
import { formatCount } from "@/shared/utils/format";
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

const compactFormatter = new Intl.NumberFormat(undefined, {
  notation: "compact",
  maximumFractionDigits: 1,
});
const PERCENT_SCALE = 100;

const overview = computed(() => summary.value?.overview ?? null);
const errorRate = computed(() => {
  const data = overview.value;
  if (data === null || data.traceCount === 0) return "0.0%";
  return `${((data.errorTraceCount / data.traceCount) * PERCENT_SCALE).toFixed(1)}%`;
});
const hasFailedTraces = computed(
  () => (overview.value?.errorTraceCount ?? 0) > 0,
);
const attentionTrace = computed(
  () => summary.value?.recentErrorTraces[0] ?? null,
);
const attentionTitle = computed(() => {
  const count = overview.value?.errorTraceCount ?? 0;
  return `${formatCount(count)} failed trace${count === 1 ? "" : "s"} in this window`;
});
const attentionBody = computed(() => {
  const trace = attentionTrace.value;
  if (trace === null) return "Open the filtered trace list to see what failed.";
  const name = trace.rootName || "Unnamed operation";
  const service = trace.rootService || "unknown service";
  return `Most recent: ${name} in ${service}.`;
});

const allTracesTo = computed(() => ({
  name: "traces",
  query: { range: range.value },
}));
const errorTracesTo = computed(() => ({
  name: "traces",
  query: { status: "error", range: range.value },
}));
const servicesTo = computed(() => ({
  name: "services",
  query: { range: range.value },
}));
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <PageHead title="Overview" :description="updatedLabel">
      <template #actions>
        <Select
          v-model="range"
          label="Window"
          :options="rangeOptions"
          class="min-w-45"
        />
        <Button :disabled="isFetching" @click="refetch()"> Refresh </Button>
      </template>
    </PageHead>

    <div
      class="scroll-thin min-h-0 flex-1 overflow-y-auto px-10 pb-6 max-[40rem]:px-4 max-[40rem]:pb-5"
    >
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
        <Button @click="refetch()">Retry</Button>
      </div>
      <div v-else-if="summary && overview" class="flex flex-col gap-6">
        <SpotlightCard
          v-if="hasFailedTraces"
          eyebrow="Needs attention"
          :title="attentionTitle"
        >
          {{ attentionBody }}
          <template #action>
            <Button variant="primary" as="a" as-child>
              <RouterLink
                :to="errorTracesTo"
                class="inline-flex items-center gap-2"
              >
                Investigate failed traces
                <component :is="arrowRightIcon" aria-hidden="true" />
              </RouterLink>
            </Button>
          </template>
        </SpotlightCard>

        <section
          class="grid grid-cols-3 gap-4 max-[50rem]:grid-cols-1"
          aria-label="Key metrics"
        >
          <RouterLink :to="allTracesTo" class="group block rounded-xl">
            <StatCard
              label="Total traces"
              :value="compactFormatter.format(overview.traceCount)"
              class="group-hover:border-default h-full transition-colors"
            />
          </RouterLink>
          <RouterLink :to="errorTracesTo" class="group block rounded-xl">
            <StatCard
              label="Error rate"
              :value="errorRate"
              :tone="overview.errorTraceCount > 0 ? 'error' : 'default'"
              class="group-hover:border-default h-full transition-colors"
            />
          </RouterLink>
          <RouterLink :to="servicesTo" class="group block rounded-xl">
            <StatCard
              label="Active services"
              :value="compactFormatter.format(overview.activeServiceCount)"
              class="group-hover:border-default h-full transition-colors"
            />
          </RouterLink>
        </section>

        <DashboardTraceChart
          :buckets="summary.traceBuckets"
          :overview="overview"
          :range="range"
        />

        <div class="grid grid-cols-2 gap-4 max-[50rem]:grid-cols-1">
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
