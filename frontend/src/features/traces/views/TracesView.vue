<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import arrowUpIcon from "@/assets/icons/arrow-up.svg";

import TraceRow from "@/features/traces/components/TraceRow.vue";
import { TRACES_PAGE_SIZE } from "@/features/traces/api";
import { useTracesQuery } from "@/features/traces/queries";
import {
  type TraceOrder,
  type TracesFilters,
  type TracesListParams,
  type TraceStatusFilter,
} from "@/features/traces/types";
import {
  useEnvironmentsQuery,
  useServicesQuery,
} from "@/shared/queries/filter-options";
import ActiveFilterChips, {
  type ActiveFilterChip,
} from "@/shared/components/ActiveFilterChips.vue";
import ChipGroup from "@/shared/components/ChipGroup.vue";
import DateRangePicker from "@/shared/components/DateRangePicker.vue";
import EmptyState from "@/shared/components/EmptyState.vue";
import FilterBar from "@/shared/components/FilterBar.vue";
import LiveToggle from "@/shared/components/LiveToggle.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import Pager from "@/shared/components/Pager.vue";
import PageHead from "@/shared/components/PageHead.vue";
import SearchBox from "@/shared/components/SearchBox.vue";
import ServiceSelect from "@/shared/components/ServiceSelect.vue";
import Button from "@/shared/components/ui/Button.vue";
import {
  readQueryParam,
  readQueryParamList,
  useDebouncedSearch,
  useRouteQuerySync,
} from "@/shared/composables/useRouteFilters";
import {
  extractTime,
  persistTime,
  resolveInitialTime,
  timeToQuery,
  type SharedTime,
} from "@/shared/composables/useGlobalTimeRange";
import { formatCount } from "@/shared/utils/format";
import {
  ALL_TIME,
  CUSTOM_RANGE,
  DAY_PRESETS,
  TIME_PRESETS,
} from "@/shared/utils/date-range";
import { diffMsPrecise, formatDateTimeRange } from "@/shared/utils/time";

const STATUS_OPTIONS: readonly {
  value: TraceStatusFilter;
  label: string;
  tone?: "error" | "ok";
}[] = [
  { value: "all", label: "All" },
  { value: "error", label: "Errors", tone: "error" },
  { value: "ok", label: "OK", tone: "ok" },
];

const SEARCH_DEBOUNCE_MS = 300;

const route = useRoute();
const router = useRouter();

function filtersFromRoute(): TracesFilters {
  const statusRaw = readQueryParam(route, "status");
  const orderRaw = readQueryParam(route, "order");
  return {
    search: readQueryParam(route, "search") ?? "",
    status:
      statusRaw === "error" || statusRaw === "ok" || statusRaw === "all"
        ? statusRaw
        : "all",
    service: readQueryParamList(route, "service"),
    environment: readQueryParamList(route, "environment"),
    order: orderRaw === "asc" ? "asc" : "desc",
    ...resolveInitialTime(route),
  };
}

const filters = ref<TracesFilters>(filtersFromRoute());
const searchInput = ref(filters.value.search);
const cursor = ref<string | undefined>(undefined);
const cursorHistory = ref<(string | undefined)[]>([]);
const debouncedQuery = useDebouncedSearch(searchInput, SEARCH_DEBOUNCE_MS);
const live = ref(false);

watch(debouncedQuery, (value) => {
  if (value === filters.value.search) return;
  filters.value = { ...filters.value, search: value };
  resetPage();
});

function filtersToQuery(): Record<string, string | string[]> {
  const query: Record<string, string | string[]> = {};
  if (filters.value.search !== "") query.search = filters.value.search;
  if (filters.value.status !== "all") query.status = filters.value.status;
  if (filters.value.service.length > 0) query.service = filters.value.service;
  if (filters.value.environment.length > 0)
    query.environment = filters.value.environment;
  if (filters.value.order !== "desc") query.order = filters.value.order;
  return { ...query, ...timeToQuery(filters.value) };
}

useRouteQuerySync(filters, filtersToQuery);

const listParams = computed((): TracesListParams => {
  if (cursor.value === undefined) return { ...filters.value };
  return { ...filters.value, cursor: cursor.value };
});

const {
  data: tracesData,
  isPending,
  isError,
  refetch,
} = useTracesQuery(listParams, live);
const servicesQuery = useServicesQuery();
const environmentsQuery = useEnvironmentsQuery();

const traces = computed(() => tracesData.value?.traces ?? []);
const total = computed(() => tracesData.value?.total ?? null);

const serviceOptions = computed(() => {
  const services = servicesQuery.data.value ?? [];
  const options = services.map((service) => ({
    value: service,
    label: service,
  }));
  for (const selected of filters.value.service) {
    if (!options.some((option) => option.value === selected)) {
      options.unshift({ value: selected, label: selected });
    }
  }
  return [{ value: "all", label: "All services" }, ...options];
});

const environmentOptions = computed(() => {
  const environments = environmentsQuery.data.value ?? [];
  const options = environments.map((environment) => ({
    value: environment,
    label: environment,
  }));
  for (const selected of filters.value.environment) {
    if (!options.some((option) => option.value === selected)) {
      options.unshift({ value: selected, label: selected });
    }
  }
  return [{ value: "all", label: "All environments" }, ...options];
});

const filtersActive = computed(
  () =>
    filters.value.search !== "" ||
    filters.value.status !== "all" ||
    filters.value.service.length > 0 ||
    filters.value.environment.length > 0 ||
    filters.value.range !== ALL_TIME,
);

const facetsActive = computed(
  () =>
    filters.value.search !== "" ||
    filters.value.status !== "all" ||
    filters.value.service.length > 0 ||
    filters.value.environment.length > 0,
);

const activeChips = computed((): ActiveFilterChip[] => {
  const chips: ActiveFilterChip[] = [];
  if (filters.value.status !== "all")
    chips.push({
      key: "status",
      label: "Status",
      value: filters.value.status === "error" ? "Errors" : "OK",
    });
  if (filters.value.service.length > 0)
    chips.push({
      key: "service",
      label: "Service",
      value: filters.value.service.join(", "),
    });
  if (filters.value.environment.length > 0)
    chips.push({
      key: "environment",
      label: "Environment",
      value: filters.value.environment.join(", "),
    });
  if (filters.value.search !== "")
    chips.push({
      key: "search",
      label: "Root operation",
      value: filters.value.search,
    });
  if (filters.value.range !== ALL_TIME) {
    const preset = TIME_PRESETS.find(
      (item) => item.value === filters.value.range,
    );
    const day = DAY_PRESETS.find((item) => item.value === filters.value.range);
    let value: string;
    if (preset !== undefined) value = `Last ${preset.label}`;
    else if (day !== undefined) value = day.label;
    else {
      value = formatDateTimeRange(filters.value.start, filters.value.end);
    }
    chips.push({ key: "timerange", label: "Time", value });
  }
  return chips;
});

const resultCount = computed(() => {
  const count = total.value;
  if (count === null) return null;
  return `${formatCount(count)} ${count === 1 ? "trace" : "traces"}`;
});

const sortAriaLabel = computed(() => {
  const current = filters.value.order === "desc" ? "newest first" : "oldest first";
  const next = filters.value.order === "desc" ? "oldest first" : "newest first";
  return `Sort by start time, currently ${current}. Activate to show ${next}.`;
});

const emptyTitle = computed(() =>
  filtersActive.value ? "No traces found" : "No traces recorded yet",
);

const emptyMessage = computed(() => {
  if (!filtersActive.value) return undefined;
  if (filters.value.search !== "")
    return `No traces match “${filters.value.search}” with these filters. Try widening the time range or clearing a filter.`;
  return "No traces match these filters. Try widening the time range or clearing a filter.";
});

const emptyActionLabel = computed(() =>
  filtersActive.value ? "Clear all filters" : undefined,
);

const maxDurationMs = computed(() => {
  let max = 0;
  for (const trace of traces.value) {
    const duration = diffMsPrecise(trace.startTime, trace.endTime);
    if (duration > max) max = duration;
  }
  return max;
});

const pageIndex = computed(() => cursorHistory.value.length);
const showingFrom = computed(() =>
  traces.value.length === 0 ? 0 : pageIndex.value * TRACES_PAGE_SIZE + 1,
);
const showingTo = computed(
  () => pageIndex.value * TRACES_PAGE_SIZE + traces.value.length,
);
const hasPrev = computed(() => cursorHistory.value.length > 0);
const hasNext = computed(() => {
  const next = tracesData.value?.nextCursor;
  return next !== undefined && next !== null;
});

function resetPage(): void {
  cursor.value = undefined;
  cursorHistory.value = [];
}

function updateTime(time: SharedTime): void {
  filters.value = {
    search: filters.value.search,
    status: filters.value.status,
    service: filters.value.service,
    environment: filters.value.environment,
    order: filters.value.order,
    ...time,
  };
  persistTime(filters.value);
  resetPage();
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

function toggleOrder(): void {
  const order: TraceOrder =
    filters.value.order === "desc" ? "asc" : "desc";
  filters.value = { ...filters.value, order };
  resetPage();
}

function clearFacets(): void {
  searchInput.value = "";
  debouncedQuery.value = "";
  const next: TracesFilters = {
    search: "",
    status: "all",
    service: [],
    environment: [],
    order: filters.value.order,
    ...extractTime(filters.value),
  };
  filters.value = next;
  resetPage();
}

function clearAllFilters(): void {
  searchInput.value = "";
  debouncedQuery.value = "";
  filters.value = {
    search: "",
    status: "all",
    service: [],
    environment: [],
    order: filters.value.order,
    range: ALL_TIME,
  };
  persistTime(filters.value);
  resetPage();
}

function removeChip(key: string): void {
  if (key === "status") filters.value = { ...filters.value, status: "all" };
  else if (key === "service") filters.value = { ...filters.value, service: [] };
  else if (key === "timerange") updateTime({ range: ALL_TIME });
  else if (key === "environment")
    filters.value = { ...filters.value, environment: [] };
  else if (key === "search") {
    searchInput.value = "";
    debouncedQuery.value = "";
    filters.value = { ...filters.value, search: "" };
  } else return;
  resetPage();
}

function nextPage(): void {
  const next = tracesData.value?.nextCursor;
  if (next === undefined || next === null) return;
  cursorHistory.value = [...cursorHistory.value, cursor.value];
  cursor.value = next;
}

function prevPage(): void {
  if (cursorHistory.value.length === 0) return;
  const history = [...cursorHistory.value];
  const previous = history.pop();
  cursorHistory.value = history;
  cursor.value = previous;
}

function openTrace(traceId: string): void {
  void router.push({ name: "trace-detail", params: { id: traceId } });
}
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <PageHead
      title="Traces"
      subtitle="Every request that entered the system, root operation first."
    />
    <FilterBar label="Filter traces">
      <div class="flex flex-wrap items-center gap-2.5">
        <div class="min-w-40 flex-1">
          <SearchBox
            v-model="searchInput"
            input-id="trace-search"
            label="Filter by root operation or trace ID"
            placeholder="Filter by root operation or trace ID"
          />
        </div>
        <LiveToggle v-model="live" @update:model-value="resetPage" />
        <DateRangePicker
          :presets="TIME_PRESETS"
          :day-presets="DAY_PRESETS"
          :active="filters.range"
          :start="filters.start"
          :end="filters.end"
          @preset="applyPreset"
          @custom="applyCustomRange"
        />
      </div>
      <div class="flex flex-wrap items-center gap-2.5">
        <ChipGroup
          v-model="filters.status"
          :options="STATUS_OPTIONS"
          group-label="Trace status"
          @update:model-value="resetPage"
        />
        <ServiceSelect
          v-model="filters.service"
          label="Service"
          :options="serviceOptions"
          trigger-class="max-w-64"
          @update:model-value="resetPage"
        />
        <ServiceSelect
          v-model="filters.environment"
          label="Environment"
          :options="environmentOptions"
          search-placeholder="Search environments…"
          trigger-class="max-w-64"
          @update:model-value="resetPage"
        />
        <div class="ml-auto flex shrink-0 items-center gap-3">
          <button
            v-if="facetsActive"
            type="button"
            class="text-secondary hover:text-primary shrink-0 cursor-pointer border-none bg-transparent p-0 text-[0.78125rem] underline decoration-strong underline-offset-2 transition-colors duration-120 ease-out"
            @click="clearFacets"
          >
            Clear all
          </button>
          <span
            v-if="resultCount !== null"
            class="text-tertiary text-[0.78125rem] whitespace-nowrap tabular-nums"
            role="status"
            >{{ resultCount }}</span
          >
        </div>
      </div>
      <ActiveFilterChips :chips="activeChips" @remove="removeChip" />
    </FilterBar>
    <div
      id="trace-list"
      class="border-subtle bg-panel mx-4 mb-3 min-h-0 flex-1 overflow-y-auto rounded-xl border"
      role="list"
      aria-label="Traces"
      tabindex="-1"
    >
      <LoadingState v-if="isPending" label="Loading traces" />
      <div
        v-else-if="isError"
        class="text-tertiary flex flex-col items-center gap-3.5 px-5 py-12 text-center text-[0.8125rem]"
        role="alert"
      >
        <span class="text-heading text-sm font-semibold"
          >Couldn’t load traces</span
        >
        <span class="leading-relaxed"
          >Something went wrong fetching traces. Try again.</span
        >
        <Button variant="secondary" @click="refetch()">Retry</Button>
      </div>
      <EmptyState
        v-else-if="traces.length === 0"
        :title="emptyTitle"
        :message="emptyMessage"
        :action-label="emptyActionLabel"
        @clear="clearAllFilters"
      />
      <template v-else>
        <div
          class="border-default bg-panel sticky top-0 z-5 flex items-center gap-3 border-b py-1.5 pr-4 pl-[1.1875rem]"
        >
          <div
            class="text-tertiary w-[3.75rem] shrink-0 text-[0.65625rem] leading-relaxed font-semibold tracking-wider whitespace-nowrap uppercase"
            aria-hidden="true"
          >
            ID
          </div>
          <div
            class="text-tertiary min-w-0 flex-1 truncate text-[0.65625rem] leading-relaxed font-semibold tracking-wider uppercase"
            aria-hidden="true"
          >
            Trace
          </div>
          <div
            class="text-tertiary w-[5.5rem] shrink-0 truncate text-[0.65625rem] leading-relaxed font-semibold tracking-wider whitespace-nowrap uppercase max-[53.75rem]:hidden"
            aria-hidden="true"
          >
            Environment
          </div>
          <div
            class="w-[9.375rem] shrink-0 max-[53.75rem]:hidden"
            aria-hidden="true"
          />
          <div
            class="text-tertiary w-[3.625rem] shrink-0 text-[0.65625rem] leading-relaxed font-semibold tracking-wider whitespace-nowrap uppercase"
            aria-hidden="true"
          >
            Duration
          </div>
          <div
            class="text-tertiary w-[3.875rem] shrink-0 text-[0.65625rem] leading-relaxed font-semibold tracking-wider whitespace-nowrap uppercase max-[53.75rem]:hidden"
            aria-hidden="true"
          >
            Spans
          </div>
          <button
            type="button"
            class="text-tertiary hover:text-primary flex w-[4.625rem] shrink-0 cursor-pointer items-center justify-end gap-1 border-none bg-transparent p-0 text-[0.65625rem] leading-relaxed font-semibold tracking-wider whitespace-nowrap uppercase transition-colors duration-120 ease-out max-[40rem]:hidden"
            :aria-label="sortAriaLabel"
            @click="toggleOrder"
          >
            <span>Time</span>
            <img
              :src="arrowUpIcon"
              alt=""
              class="size-3 shrink-0 transition-transform duration-120 ease-out"
              :class="{ 'rotate-180': filters.order === 'desc' }"
            />
          </button>
        </div>
        <TraceRow
          v-for="trace in traces"
          :key="trace.traceId"
          :trace="trace"
          :max-duration-ms="maxDurationMs"
          @open-trace="openTrace"
        />
      </template>
    </div>
    <Pager
      :from="showingFrom"
      :to="showingTo"
      :total="total"
      :has-prev="hasPrev"
      :has-next="hasNext"
      item-name="traces"
      @prev="prevPage"
      @next="nextPage"
    />
  </div>
</template>
