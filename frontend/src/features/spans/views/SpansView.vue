<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import SpanRow from "@/features/spans/components/SpanRow.vue";
import { SPANS_PAGE_SIZE } from "@/features/spans/api";
import { useServicesQuery, useSpansQuery } from "@/features/spans/queries";
import {
  CUSTOM_RANGE,
  DEFAULT_SPANS_FILTERS,
  TIME_PRESETS,
  type SpansFilters,
  type SpansListParams,
  type SpanStatusFilter,
} from "@/features/spans/types";
import ChipGroup from "@/shared/components/ChipGroup.vue";
import DateRangePicker from "@/shared/components/DateRangePicker.vue";
import EmptyState from "@/shared/components/EmptyState.vue";
import FilterBar from "@/shared/components/FilterBar.vue";
import Pager from "@/shared/components/Pager.vue";
import PageHead from "@/shared/components/PageHead.vue";
import ResultStrip from "@/shared/components/ResultStrip.vue";
import SearchBox from "@/shared/components/SearchBox.vue";
import Button from "@/shared/components/ui/Button.vue";
import Select from "@/shared/components/ui/Select.vue";
import type { SpanKind } from "@/shared/types/entities/span";
import { MS_PER_MINUTE } from "@/shared/utils/time";

const STATUS_OPTIONS: readonly { value: SpanStatusFilter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "error", label: "Errors" },
  { value: "ok", label: "OK" },
];

const KIND_OPTIONS: readonly { value: "all" | SpanKind; label: string }[] = [
  { value: "all", label: "All kinds" },
  { value: "unspecified", label: "unspecified" },
  { value: "internal", label: "internal" },
  { value: "server", label: "server" },
  { value: "client", label: "client" },
  { value: "producer", label: "producer" },
  { value: "consumer", label: "consumer" },
];

const SEARCH_DEBOUNCE_MS = 300;
const PRESET_MATCH_TOLERANCE_MIN = 1;

const route = useRoute();
const router = useRouter();

function queryParam(name: string): string | undefined {
  const raw = route.query[name];
  if (typeof raw === "string" && raw !== "") return raw;
  return undefined;
}

function filtersFromRoute(): SpansFilters {
  const statusRaw = queryParam("status");
  const kindRaw = queryParam("kind");
  const next: SpansFilters = {
    search: queryParam("search") ?? "",
    status:
      statusRaw === "error" || statusRaw === "ok" || statusRaw === "all"
        ? statusRaw
        : "all",
    kind:
      kindRaw === "all" ||
      kindRaw === "unspecified" ||
      kindRaw === "internal" ||
      kindRaw === "server" ||
      kindRaw === "client" ||
      kindRaw === "producer" ||
      kindRaw === "consumer"
        ? kindRaw
        : "all",
    service: queryParam("service") ?? "all",
  };
  const start = queryParam("start");
  const end = queryParam("end");
  if (start !== undefined) next.start = start;
  if (end !== undefined) next.end = end;
  return next;
}

const filters = ref<SpansFilters>(filtersFromRoute());
const searchInput = ref(filters.value.search);
const cursor = ref<string | undefined>(undefined);
const cursorHistory = ref<(string | undefined)[]>([]);
const debouncedQuery = ref(filters.value.search);

let searchTimer: ReturnType<typeof setTimeout> | undefined;
watch(searchInput, (value) => {
  if (searchTimer !== undefined) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    debouncedQuery.value = value;
  }, SEARCH_DEBOUNCE_MS);
});
onUnmounted(() => {
  if (searchTimer !== undefined) clearTimeout(searchTimer);
});

watch(debouncedQuery, (value) => {
  if (value === filters.value.search) return;
  filters.value = { ...filters.value, search: value };
  resetPage();
});

function filtersToQuery(): Record<string, string> {
  const query: Record<string, string> = {};
  if (filters.value.search !== "") query.search = filters.value.search;
  if (filters.value.status !== "all") query.status = filters.value.status;
  if (filters.value.kind !== "all") query.kind = filters.value.kind;
  if (filters.value.service !== "all") query.service = filters.value.service;
  if (filters.value.start !== undefined) query.start = filters.value.start;
  if (filters.value.end !== undefined) query.end = filters.value.end;
  return query;
}

watch(
  filters,
  () => {
    void router.replace({ query: filtersToQuery() });
  },
  { deep: true },
);

const listParams = computed((): SpansListParams => {
  if (cursor.value === undefined) return { ...filters.value };
  return { ...filters.value, cursor: cursor.value };
});

const {
  data: spansData,
  isPending,
  isError,
  refetch,
  isFetching,
} = useSpansQuery(listParams);
const servicesQuery = useServicesQuery();

const spans = computed(() => spansData.value?.spans ?? []);
const total = computed(() => spansData.value?.total ?? null);

const serviceOptions = computed(() => {
  const services = servicesQuery.data.value ?? [];
  const options = services.map((service) => ({
    value: service,
    label: service,
  }));
  if (
    filters.value.service !== "all" &&
    !options.some((option) => option.value === filters.value.service)
  ) {
    options.unshift({
      value: filters.value.service,
      label: filters.value.service,
    });
  }
  return [{ value: "all", label: "All services" }, ...options];
});

const activePreset = computed(() => {
  if (filters.value.start === undefined && filters.value.end === undefined)
    return "1h";
  if (filters.value.end !== undefined || filters.value.start === undefined)
    return CUSTOM_RANGE;
  const minutesAgo =
    (Date.now() - Date.parse(filters.value.start)) / MS_PER_MINUTE;
  const match = TIME_PRESETS.find(
    (preset) =>
      Math.abs(preset.minutes - minutesAgo) < PRESET_MATCH_TOLERANCE_MIN,
  );
  return match?.value ?? CUSTOM_RANGE;
});

const filtersActive = computed(
  () =>
    filters.value.search !== "" ||
    filters.value.status !== "all" ||
    filters.value.kind !== "all" ||
    filters.value.service !== "all" ||
    filters.value.start !== undefined ||
    filters.value.end !== undefined,
);

const emptyMessage = computed(() =>
  filters.value.search !== ""
    ? `No spans match “${filters.value.search}” with these filters. Try a shorter name or widen the time range.`
    : "No spans match these filters. Try widening the time range or clearing a filter.",
);

const maxDurationMs = computed(() => {
  let max = 0;
  for (const span of spans.value) {
    const duration = Date.parse(span.endTime) - Date.parse(span.startTime);
    if (duration > max) max = duration;
  }
  return max;
});

const pageIndex = computed(() => cursorHistory.value.length);
const showingFrom = computed(() =>
  spans.value.length === 0 ? 0 : pageIndex.value * SPANS_PAGE_SIZE + 1,
);
const showingTo = computed(
  () => pageIndex.value * SPANS_PAGE_SIZE + spans.value.length,
);
const hasPrev = computed(() => cursorHistory.value.length > 0);
const hasNext = computed(() => {
  const next = spansData.value?.nextCursor;
  return next !== undefined && next !== null;
});

function resetPage(): void {
  cursor.value = undefined;
  cursorHistory.value = [];
}

function setTimeRange(
  start: string | undefined,
  end: string | undefined,
): void {
  const next: SpansFilters = {
    search: filters.value.search,
    status: filters.value.status,
    kind: filters.value.kind,
    service: filters.value.service,
  };
  if (start !== undefined) next.start = start;
  if (end !== undefined) next.end = end;
  filters.value = next;
  resetPage();
}

function applyPreset(value: string): void {
  const preset = TIME_PRESETS.find((item) => item.value === value);
  if (preset === undefined) return;
  setTimeRange(
    new Date(Date.now() - preset.minutes * MS_PER_MINUTE).toISOString(),
    undefined,
  );
}

function applyCustomRange(range: { start?: string; end?: string }): void {
  setTimeRange(range.start, range.end);
}

function clearFilters(): void {
  searchInput.value = "";
  debouncedQuery.value = "";
  filters.value = { ...DEFAULT_SPANS_FILTERS };
  resetPage();
}

function nextPage(): void {
  const next = spansData.value?.nextCursor;
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
      title="Spans"
      subtitle="Every unit of work recorded, independent of which trace it belongs to."
    />
    <FilterBar
      label="Filter spans"
      :show-clear="filtersActive"
      @clear="clearFilters"
    >
      <SearchBox
        v-model="searchInput"
        input-id="span-search"
        label="Filter by span name"
        placeholder="Filter by span name, e.g. GET /cart"
      />
      <ChipGroup
        v-model="filters.status"
        :options="STATUS_OPTIONS"
        group-label="Span status"
        @update:model-value="resetPage"
      />
      <Select
        v-model="filters.kind"
        label="Kind"
        :options="KIND_OPTIONS"
        @update:model-value="resetPage"
      />
      <DateRangePicker
        :presets="TIME_PRESETS"
        :active="activePreset"
        :start="filters.start"
        :end="filters.end"
        @preset="applyPreset"
        @custom="applyCustomRange"
      />
      <Select
        v-model="filters.service"
        label="Service"
        :options="serviceOptions"
        trigger-class="max-w-[9.375rem]"
        @update:model-value="resetPage"
      />
    </FilterBar>
    <ResultStrip
      v-if="!isPending && !isError"
      :count="total ?? spans.length"
      noun="span"
    />
    <div
      id="span-list"
      class="min-h-0 flex-1 overflow-y-auto"
      role="list"
      aria-label="Spans"
      tabindex="-1"
      :class="isFetching && 'opacity-60'"
    >
      <div
        v-if="isPending"
        class="text-tertiary flex flex-col items-center gap-1.5 px-5 py-12 text-center text-[0.8125rem]"
        role="status"
      >
        <span class="text-heading text-sm font-semibold">Loading spans…</span>
      </div>
      <div
        v-else-if="isError"
        class="text-tertiary flex flex-col items-center gap-3.5 px-5 py-12 text-center text-[0.8125rem]"
        role="alert"
      >
        <span class="text-heading text-sm font-semibold"
          >Couldn’t load spans</span
        >
        <span class="leading-relaxed"
          >Something went wrong fetching spans. Try again.</span
        >
        <Button variant="secondary" @click="refetch()">Retry</Button>
      </div>
      <EmptyState
        v-else-if="spans.length === 0"
        title="No spans found"
        :message="emptyMessage"
        action-label="Clear all filters"
        @clear="clearFilters"
      />
      <template v-else>
        <div
          class="border-default bg-panel sticky top-0 z-5 flex items-center gap-3 border-b py-1.5 pr-4 pl-[1.1875rem]"
          aria-hidden="true"
        >
          <div class="w-4 shrink-0" />
          <div
            class="text-tertiary min-w-0 flex-1 truncate text-[0.65625rem] leading-relaxed font-semibold tracking-wider uppercase"
          >
            Span
          </div>
          <div class="w-[8.125rem] shrink-0 max-[53.75rem]:hidden" />
          <div
            class="text-tertiary w-[3.625rem] shrink-0 text-[0.65625rem] leading-relaxed font-semibold tracking-wider whitespace-nowrap uppercase"
          >
            Duration
          </div>
          <div
            class="text-tertiary w-[4.625rem] shrink-0 text-[0.65625rem] leading-relaxed font-semibold tracking-wider whitespace-nowrap uppercase"
          >
            Time
          </div>
        </div>
        <SpanRow
          v-for="span in spans"
          :key="span.spanId"
          :span="span"
          :max-duration-ms="maxDurationMs"
          @open-trace="openTrace"
        />
      </template>
    </div>
    <Pager
      :from="showingFrom"
      :to="showingTo"
      :has-prev="hasPrev"
      :has-next="hasNext"
      item-name="spans"
      @prev="prevPage"
      @next="nextPage"
    />
  </div>
</template>
