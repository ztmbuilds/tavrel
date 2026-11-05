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

const rangeDetail = computed(() => {
  if (filters.value.range === CUSTOM_RANGE) return "in a custom range";
  if (filters.value.range === ALL_TIME) return "across all time";
  const day = DAY_PRESETS.find((item) => item.value === filters.value.range);
  if (day !== undefined) return `for ${day.label.toLowerCase()}`;
  const preset = TIME_PRESETS.find(
    (item) => item.value === filters.value.range,
  );
  return preset === undefined ? "" : `in the last ${preset.label}`;
});

const description = computed(() => {
  const detail = rangeDetail.value;
  return detail === "" ? countLabel.value : `${countLabel.value} ${detail}`;
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
    <PageHead title="Services" :description="description" description-live>
      <template #actions>
        <div
          class="w-56 max-[40rem]:w-auto max-[40rem]:min-w-0 max-[40rem]:flex-1"
        >
          <SearchBox
            v-model="searchInput"
            input-id="service-search"
            label="Find a service"
            placeholder="Find a service…"
          />
        </div>
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

    <div
      class="scroll-thin mx-10 mb-4 grid max-h-[40rem] min-h-0 flex-1 grid-cols-[minmax(0,1fr)_18rem] gap-4 max-[65rem]:grid-cols-1 max-[65rem]:overflow-y-auto max-[40rem]:mx-4"
      @keydown.escape="clearSelection"
    >
      <div
        class="border-subtle bg-canvas min-h-0 min-w-0 overflow-hidden rounded-xl border max-[65rem]:h-[24rem]"
      >
        <div v-if="isPending" class="flex h-full items-center justify-center">
          <LoadingState label="Loading service graph" />
        </div>
        <div
          v-else-if="isError"
          class="text-tertiary flex h-full flex-col items-center justify-center gap-3.5 px-5 py-12 text-center text-[0.8125rem]"
          role="alert"
        >
          <span class="text-heading text-sm font-semibold"
            >Couldn’t load service graph</span
          >
          <span class="leading-relaxed"
            >Something went wrong fetching the service graph. Try again.</span
          >
          <Button @click="refetch()">Retry</Button>
        </div>
        <div
          v-else-if="!hasGraph"
          class="flex h-full items-center justify-center"
        >
          <EmptyState
            title="No services found"
            message="No service calls in this time window. Try widening the time range."
          />
        </div>
        <ServiceMapCanvas
          v-else
          :layout="layout"
          :selected="selected"
          :query="searchInput"
          @select="selectService"
        />
      </div>
      <aside
        class="border-subtle bg-panel min-h-0 min-w-0 overflow-hidden rounded-xl border max-[65rem]:min-h-56"
        aria-label="Service inspector"
      >
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
