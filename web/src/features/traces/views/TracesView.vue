<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, type RouteLocationNormalizedLoaded } from "vue-router";

import arrowUpIcon from "@/assets/icons/arrow-up.svg?component";

import TraceRow from "@/features/traces/components/TraceRow.vue";
import { useInfiniteTracesQuery } from "@/features/traces/queries";
import {
  type TraceOrder,
  type TracesFilters,
  type TraceStatusFilter,
} from "@/features/traces/types";
import { useInfiniteScroll } from "@/shared/composables/useInfiniteScroll";
import {
  useEnvironmentsQuery,
  useServicesQuery,
} from "@/shared/queries/filter-options";
import ActiveFilterChips, {
  type ActiveFilterChip,
} from "@/shared/components/ActiveFilterChips.vue";
import DateRangePicker from "@/shared/components/DateRangePicker.vue";
import EmptyState from "@/shared/components/EmptyState.vue";
import FilterBar from "@/shared/components/FilterBar.vue";
import LiveToggle from "@/shared/components/LiveToggle.vue";
import LoadingSpinner from "@/shared/components/LoadingSpinner.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageHead from "@/shared/components/PageHead.vue";
import SearchBox from "@/shared/components/SearchBox.vue";
import ServiceSelect from "@/shared/components/ServiceSelect.vue";
import Button from "@/shared/components/ui/Button.vue";
import SegmentedControl from "@/shared/components/ui/SegmentedControl.vue";
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
import {
  diffMsPrecise,
  formatDateTimeRange,
  formatDurationRange,
} from "@/shared/utils/time";

const STATUS_OPTIONS: readonly {
  value: TraceStatusFilter;
  label: string;
  dot?: string;
}[] = [
  { value: "all", label: "All" },
  { value: "error", label: "Errors", dot: "bg-error" },
  { value: "ok", label: "OK", dot: "bg-ok" },
];

const SEARCH_DEBOUNCE_MS = 300;
const BACK_TO_TOP_THRESHOLD_PX = 600;

const route = useRoute();

function readDurationMsParam(
  route: RouteLocationNormalizedLoaded,
  name: string,
): number | undefined {
  const raw = readQueryParam(route, name);
  if (raw === undefined) return undefined;
  const parsed = Number(raw);
  return Number.isInteger(parsed) && parsed >= 0 ? parsed : undefined;
}

function filtersFromRoute(): TracesFilters {
  const statusRaw = readQueryParam(route, "status");
  const orderRaw = readQueryParam(route, "order");
  const minDurationMs = readDurationMsParam(route, "minDurationMs");
  const maxDurationMs = readDurationMsParam(route, "maxDurationMs");
  const filters: TracesFilters = {
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
  if (minDurationMs !== undefined) filters.minDurationMs = minDurationMs;
  if (maxDurationMs !== undefined) filters.maxDurationMs = maxDurationMs;
  return filters;
}

const filters = ref<TracesFilters>(filtersFromRoute());
const searchInput = ref(filters.value.search);
const debouncedQuery = useDebouncedSearch(searchInput, SEARCH_DEBOUNCE_MS);
const live = ref(false);
const scrollerEl = ref<Element | null>(null);
const sentinelEl = ref<Element | null>(null);
const showBackToTop = ref(false);

watch(debouncedQuery, (value) => {
  if (value === filters.value.search) return;
  filters.value = { ...filters.value, search: value };
  resetList();
});

function filtersToQuery(): Record<string, string | string[]> {
  const query: Record<string, string | string[]> = {};
  if (filters.value.search !== "") query.search = filters.value.search;
  if (filters.value.status !== "all") query.status = filters.value.status;
  if (filters.value.service.length > 0) query.service = filters.value.service;
  if (filters.value.environment.length > 0)
    query.environment = filters.value.environment;
  if (filters.value.order !== "desc") query.order = filters.value.order;
  if (filters.value.minDurationMs !== undefined)
    query.minDurationMs = String(filters.value.minDurationMs);
  if (filters.value.maxDurationMs !== undefined)
    query.maxDurationMs = String(filters.value.maxDurationMs);
  return { ...query, ...timeToQuery(filters.value) };
}

useRouteQuerySync(filters, filtersToQuery);

const {
  data: tracesData,
  isPending,
  isError,
  refetch,
  fetchNextPage,
  hasNextPage,
  isFetchingNextPage,
  isFetchNextPageError,
} = useInfiniteTracesQuery(filters, live);
const servicesQuery = useServicesQuery();
const environmentsQuery = useEnvironmentsQuery();

const traces = computed(
  () => tracesData.value?.pages.flatMap((page) => page.traces) ?? [],
);

const canLoadMore = computed(
  () => hasNextPage.value && !isFetchingNextPage.value,
);

function loadMore(): void {
  if (!canLoadMore.value) return;
  void fetchNextPage();
}

useInfiniteScroll({
  sentinel: sentinelEl,
  scroller: scrollerEl,
  canLoadMore,
  onLoadMore: loadMore,
});

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
    filters.value.minDurationMs !== undefined ||
    filters.value.maxDurationMs !== undefined ||
    filters.value.range !== ALL_TIME,
);

const facetsActive = computed(
  () =>
    filters.value.search !== "" ||
    filters.value.status !== "all" ||
    filters.value.service.length > 0 ||
    filters.value.environment.length > 0 ||
    filters.value.minDurationMs !== undefined ||
    filters.value.maxDurationMs !== undefined,
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
  if (
    filters.value.minDurationMs !== undefined ||
    filters.value.maxDurationMs !== undefined
  )
    chips.push({
      key: "duration",
      label: "Duration",
      value: formatDurationRange(
        filters.value.minDurationMs ?? 0,
        filters.value.maxDurationMs,
      ),
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

const sortAriaLabel = computed(() => {
  const current =
    filters.value.order === "desc" ? "newest first" : "oldest first";
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

const maxTraceDurationMs = computed(() => {
  let max = 0;
  for (const trace of traces.value) {
    const duration = diffMsPrecise(trace.startTime, trace.endTime);
    if (duration > max) max = duration;
  }
  return max;
});

const loadedCount = computed(() => {
  const count = traces.value.length;
  return `${formatCount(count)} ${count === 1 ? "trace" : "traces"} loaded`;
});

const endOfListLabel = "No more traces";

function resetList(): void {
  showBackToTop.value = false;
  scrollerEl.value?.scrollTo({ top: 0 });
}

function onScrollerScroll(event: Event): void {
  if (!(event.target instanceof HTMLElement)) return;
  showBackToTop.value = event.target.scrollTop > BACK_TO_TOP_THRESHOLD_PX;
}

function scrollToTop(): void {
  const reduceMotion = window.matchMedia(
    "(prefers-reduced-motion: reduce)",
  ).matches;
  scrollerEl.value?.scrollTo({
    top: 0,
    behavior: reduceMotion ? "auto" : "smooth",
  });
}

function updateTime(time: SharedTime): void {
  filters.value = {
    search: filters.value.search,
    status: filters.value.status,
    service: filters.value.service,
    environment: filters.value.environment,
    order: filters.value.order,
    ...extractDuration(filters.value),
    ...time,
  };
  persistTime(filters.value);
  resetList();
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

function extractDuration(filters: TracesFilters): {
  minDurationMs?: number;
  maxDurationMs?: number;
} {
  const duration: { minDurationMs?: number; maxDurationMs?: number } = {};
  if (filters.minDurationMs !== undefined)
    duration.minDurationMs = filters.minDurationMs;
  if (filters.maxDurationMs !== undefined)
    duration.maxDurationMs = filters.maxDurationMs;
  return duration;
}

function toggleOrder(): void {
  const order: TraceOrder = filters.value.order === "desc" ? "asc" : "desc";
  filters.value = { ...filters.value, order };
  resetList();
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
  resetList();
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
  resetList();
}

function removeChip(key: string): void {
  if (key === "status") filters.value = { ...filters.value, status: "all" };
  else if (key === "service") filters.value = { ...filters.value, service: [] };
  else if (key === "timerange") {
    updateTime({ range: ALL_TIME });
    return;
  } else if (key === "duration") {
    const next = { ...filters.value };
    delete next.minDurationMs;
    delete next.maxDurationMs;
    filters.value = next;
  } else if (key === "environment")
    filters.value = { ...filters.value, environment: [] };
  else if (key === "search") {
    searchInput.value = "";
    debouncedQuery.value = "";
    filters.value = { ...filters.value, search: "" };
  } else return;
  resetList();
}
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <PageHead
      title="Traces"
      description="Every request that entered the system, root operation first."
    />
    <FilterBar label="Filter traces">
      <div class="flex flex-wrap items-center gap-2.5">
        <div class="min-w-40 flex-1 max-[40rem]:basis-full">
          <SearchBox
            v-model="searchInput"
            input-id="trace-search"
            label="Filter by root operation or trace ID"
            placeholder="Filter by root operation or trace ID"
          />
        </div>
        <LiveToggle v-model="live" @update:model-value="resetList" />
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
        <SegmentedControl
          v-model="filters.status"
          :options="STATUS_OPTIONS"
          aria-label="Trace status"
          @update:model-value="resetList"
        />
        <ServiceSelect
          v-model="filters.service"
          label="Service"
          :options="serviceOptions"
          trigger-class="max-w-64"
          @update:model-value="resetList"
        />
        <ServiceSelect
          v-model="filters.environment"
          label="Environment"
          :options="environmentOptions"
          search-placeholder="Search environments…"
          trigger-class="max-w-64"
          @update:model-value="resetList"
        />
        <div class="ml-auto flex shrink-0 items-center gap-3">
          <Button
            v-if="facetsActive"
            variant="ghost"
            size="sm"
            class="h-auto shrink-0 p-0"
            @click="clearFacets"
          >
            Clear all
          </Button>
        </div>
      </div>
      <ActiveFilterChips :chips="activeChips" @remove="removeChip" />
    </FilterBar>
    <div class="relative mx-10 mb-4 min-h-0 flex-1 max-[40rem]:mx-4">
      <div
        id="trace-list"
        ref="scrollerEl"
        class="border-subtle bg-panel h-full overflow-y-auto rounded-xl border"
        tabindex="-1"
        @scroll="onScrollerScroll"
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
          <Button @click="refetch()">Retry</Button>
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
            class="border-default bg-panel sticky top-0 z-5 flex h-11 items-center gap-3 border-b px-5"
          >
            <div
              class="text-tertiary w-[5.25rem] shrink-0 text-xs leading-relaxed font-normal tracking-[0.06em] whitespace-nowrap uppercase"
              aria-hidden="true"
            >
              ID
            </div>
            <div
              class="text-tertiary min-w-0 flex-1 truncate text-xs leading-relaxed font-normal tracking-[0.06em] uppercase"
              aria-hidden="true"
            >
              Trace
            </div>
            <div
              class="text-tertiary w-28 shrink-0 truncate text-xs leading-relaxed font-normal tracking-[0.06em] whitespace-nowrap uppercase max-[53.75rem]:hidden"
              aria-hidden="true"
            >
              Environment
            </div>
            <div
              class="text-tertiary w-[9.375rem] shrink-0 truncate text-xs leading-relaxed font-normal tracking-[0.06em] uppercase max-[53.75rem]:hidden"
              aria-hidden="true"
            >
              Timeline
            </div>
            <div
              class="text-tertiary w-[3.625rem] shrink-0 text-right text-xs leading-relaxed font-normal tracking-[0.06em] whitespace-nowrap uppercase"
              aria-hidden="true"
            >
              Duration
            </div>
            <div
              class="text-tertiary w-[3.875rem] shrink-0 text-right text-xs leading-relaxed font-normal tracking-[0.06em] whitespace-nowrap uppercase max-[53.75rem]:hidden"
              aria-hidden="true"
            >
              Spans
            </div>
            <button
              type="button"
              class="text-tertiary hover:text-primary flex w-[10.75rem] shrink-0 cursor-pointer items-center justify-end gap-1 border-none bg-transparent p-0 text-xs leading-relaxed font-normal tracking-[0.06em] whitespace-nowrap uppercase transition-colors duration-120 ease-out max-[40rem]:hidden"
              :aria-label="sortAriaLabel"
              @click="toggleOrder"
            >
              <span>Time</span>
              <component
                :is="arrowUpIcon"
                aria-hidden="true"
                class="size-3 shrink-0 transition-transform duration-120 ease-out"
                :class="{ 'rotate-180': filters.order === 'desc' }"
              />
            </button>
          </div>
          <div role="list" aria-label="Traces">
            <TraceRow
              v-for="trace in traces"
              :key="trace.traceId"
              :trace="trace"
              :max-duration-ms="maxTraceDurationMs"
            />
          </div>
          <div
            ref="sentinelEl"
            class="flex flex-col items-center gap-2 px-5 py-4 text-center"
          >
            <template v-if="isFetchingNextPage">
              <LoadingSpinner size="xs" label="Loading more traces" />
            </template>
            <template v-else-if="isFetchNextPageError">
              <span class="text-tertiary text-[0.78125rem]" role="alert"
                >Couldn’t load more traces.</span
              >
              <Button @click="loadMore()">Retry</Button>
            </template>
            <template v-else-if="hasNextPage">
              <span
                class="text-tertiary text-[0.78125rem] tabular-nums"
                role="status"
                >{{ loadedCount }}</span
              >
              <Button @click="loadMore()">Load more traces</Button>
            </template>
            <span
              v-else
              class="text-tertiary text-[0.78125rem] tabular-nums"
              role="status"
              >{{ endOfListLabel }}</span
            >
          </div>
        </template>
      </div>
      <Button
        v-show="showBackToTop"
        aria-label="Back to top of traces"
        title="Back to top"
        class="absolute right-4 bottom-4 z-10 size-10 rounded-full p-0"
        @click="scrollToTop"
      >
        <component :is="arrowUpIcon" aria-hidden="true" class="text-tertiary" />
      </Button>
    </div>
  </div>
</template>
