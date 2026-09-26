<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, type RouteLocationRaw } from "vue-router";

import ServiceInspector from "@/features/services/components/ServiceInspector.vue";
import ServiceMapCanvas from "@/features/services/components/ServiceMapCanvas.vue";
import { useServiceGraphQuery } from "@/features/services/queries";
import type { ServiceGraphFilters } from "@/features/services/types";
import { computeMapLayout } from "@/features/services/layout";
import DateRangePicker from "@/shared/components/DateRangePicker.vue";
import EmptyState from "@/shared/components/EmptyState.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageHead from "@/shared/components/PageHead.vue";
import SearchBox from "@/shared/components/SearchBox.vue";
import Button from "@/shared/components/ui/Button.vue";
import {
  readQueryParam,
  useRouteQuerySync,
} from "@/shared/composables/useRouteFilters";
import {
  extractTime,
  persistTime,
  resolveInitialTime,
  timeToQuery,
  type SharedTime,
} from "@/shared/composables/useGlobalTimeRange";
import {
  ALL_TIME,
  CUSTOM_RANGE,
  DAY_PRESETS,
  TIME_PRESETS,
} from "@/shared/utils/date-range";

const route = useRoute();

function filtersFromRoute(): ServiceGraphFilters {
  const next: ServiceGraphFilters = { ...resolveInitialTime(route) };
  const selected = readQueryParam(route, "selected");
  if (selected !== undefined) next.selected = selected;
  return next;
}

const filters = ref<ServiceGraphFilters>(filtersFromRoute());
const searchInput = ref("");

function filtersToQuery(): Record<string, string | string[]> {
  const query: Record<string, string | string[]> = {
    ...timeToQuery(filters.value),
  };
  const selected = filters.value.selected;
  if (selected !== undefined) query.selected = selected;
  return query;
}

useRouteQuerySync(filters, filtersToQuery);

const timeOnly = computed((): SharedTime => extractTime(filters.value));
const graphQuery = useServiceGraphQuery(timeOnly);
const { isPending, isError, refetch } = graphQuery;

const graph = computed(() => graphQuery.data.value ?? null);

const layout = computed(() =>
  computeMapLayout(
    graph.value?.nodes.map((node) => node.id) ?? [],
    graph.value?.edges ?? [],
  ),
);

const selected = computed(() => filters.value.selected ?? null);

const calls = computed((): string[] => {
  if (selected.value === null || graph.value === null) return [];
  return graph.value.edges
    .filter((edge) => edge.from === selected.value)
    .map((edge) => edge.to)
    .sort((a, b) => a.localeCompare(b));
});

const calledBy = computed((): string[] => {
  if (selected.value === null || graph.value === null) return [];
  return graph.value.edges
    .filter((edge) => edge.to === selected.value)
    .map((edge) => edge.from)
    .sort((a, b) => a.localeCompare(b));
});

const tracesTo = computed((): RouteLocationRaw | null => {
  if (selected.value === null) return null;
  return {
    name: "traces",
    query: { service: selected.value, ...timeToQuery(filters.value) },
  };
});

const logsTo = computed((): RouteLocationRaw | null => {
  if (selected.value === null) return null;
  return {
    name: "logs",
    query: { service: selected.value, ...timeToQuery(filters.value) },
  };
});

const nodeCount = computed(() => graph.value?.nodes.length ?? 0);
const edgeCount = computed(() => graph.value?.edges.length ?? 0);

const countLabel = computed(() => {
  const services = nodeCount.value === 1 ? "service" : "services";
  const connections = edgeCount.value === 1 ? "connection" : "connections";
  return `${nodeCount.value} ${services} · ${edgeCount.value} ${connections}`;
});

const hasGraph = computed(() => nodeCount.value > 0);

watch(
  () => graphQuery.data.value,
  (data) => {
    if (data === undefined) return;
    const current = filters.value.selected;
    if (
      current !== undefined &&
      !data.nodes.some((node) => node.id === current)
    ) {
      clearSelection();
    }
  },
);

function selectService(id: string): void {
  if (filters.value.selected === id) {
    clearSelection();
    return;
  }
  filters.value = { ...filters.value, selected: id };
}

function clearSelection(): void {
  if (filters.value.selected === undefined) return;
  filters.value = { ...extractTime(filters.value) };
}

function updateTime(time: SharedTime): void {
  const next: ServiceGraphFilters = { ...extractTime(time) };
  const current = filters.value.selected;
  if (current !== undefined) next.selected = current;
  filters.value = next;
  persistTime(next);
}

function applyPreset(value: string): void {
  if (value === ALL_TIME) {
    updateTime({ range: ALL_TIME });
    return;
  }
  if (
    TIME_PRESETS.some((item) => item.value === value) ||
    DAY_PRESETS.some((item) => item.value === value)
  )
    updateTime({ range: value });
}

function applyCustomRange(range: { start?: string; end?: string }): void {
  if (range.start === undefined && range.end === undefined) return;
  const time: SharedTime = { range: CUSTOM_RANGE };
  if (range.start !== undefined) time.start = range.start;
  if (range.end !== undefined) time.end = range.end;
  updateTime(time);
}
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <PageHead title="Service map">
      <template #actions>
        <DateRangePicker
          :presets="TIME_PRESETS"
          :day-presets="DAY_PRESETS"
          :active="filters.range"
          :start="filters.start"
          :end="filters.end"
          @preset="applyPreset"
          @custom="applyCustomRange"
        />
      </template>
    </PageHead>
    <p class="text-secondary shrink-0 px-4 pt-1 text-[0.8125rem]">
      How your services call each other, built from trace data in the selected
      window.
    </p>
    <div class="flex shrink-0 items-center gap-2.5 px-4 py-3">
      <div class="max-w-70 min-w-40 flex-1">
        <SearchBox
          v-model="searchInput"
          input-id="service-search"
          label="Find a service"
          placeholder="Find a service…"
        />
      </div>
      <span class="text-tertiary shrink-0 text-xs tabular-nums" role="status">{{
        countLabel
      }}</span>
    </div>
    <div
      class="border-subtle bg-subtle mx-4 mb-3 grid min-h-0 flex-1 grid-cols-[minmax(0,1fr)_16.25rem] gap-px overflow-hidden rounded-xl border max-[53.75rem]:grid-cols-1 max-[53.75rem]:overflow-y-auto"
      @keydown.escape="clearSelection"
    >
      <div class="bg-panel min-h-0 min-w-0">
        <LoadingState v-if="isPending" label="Loading service graph" />
        <div
          v-else-if="isError"
          class="text-tertiary flex flex-col items-center gap-3.5 px-5 py-12 text-center text-[0.8125rem]"
          role="alert"
        >
          <span class="text-heading text-sm font-semibold"
            >Couldn’t load service graph</span
          >
          <span class="leading-relaxed"
            >Something went wrong fetching the service graph. Try again.</span
          >
          <Button variant="secondary" @click="refetch()">Retry</Button>
        </div>
        <EmptyState
          v-else-if="!hasGraph"
          title="No services found"
          message="No service calls in this time window. Try widening the time range."
        />
        <ServiceMapCanvas
          v-else
          :layout="layout"
          :selected="selected"
          :query="searchInput"
          @select="selectService"
        />
      </div>
      <aside class="bg-panel min-w-0" aria-label="Service inspector">
        <ServiceInspector
          :selected="selected"
          :calls="calls"
          :called-by="calledBy"
          :traces-to="tracesTo"
          :logs-to="logsTo"
          @select="selectService"
          @clear="clearSelection"
        />
      </aside>
    </div>
  </div>
</template>
