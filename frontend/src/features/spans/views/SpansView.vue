<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import SpanRow from "@/features/spans/components/SpanRow.vue";
import { SPANS_PAGE_SIZE } from "@/features/spans/api";
import { useServicesQuery, useSpansQuery } from "@/features/spans/queries";
import {
  CUSTOM_RANGE,
  DAY_PRESETS,
  TIME_PRESETS,
  dayPresetRange,
  matchDayPreset,
} from "@/shared/utils/date-range";
import {
  type SpansFilters,
  type SpansListParams,
  type SpanStatusFilter,
} from "@/features/spans/types";
import ActiveFilterChips, {
  type ActiveFilterChip,
} from "@/shared/components/ActiveFilterChips.vue";
import ChipGroup from "@/shared/components/ChipGroup.vue";
import DateRangePicker from "@/shared/components/DateRangePicker.vue";
import EmptyState from "@/shared/components/EmptyState.vue";
import FilterBar from "@/shared/components/FilterBar.vue";
import LiveToggle from "@/shared/components/LiveToggle.vue";
import Pager from "@/shared/components/Pager.vue";
import PageHead from "@/shared/components/PageHead.vue";
import SearchBox from "@/shared/components/SearchBox.vue";
import ServiceSelect from "@/shared/components/ServiceSelect.vue";
import Button from "@/shared/components/ui/Button.vue";
import Select from "@/shared/components/ui/Select.vue";
import type { SpanKind } from "@/shared/types/entities/span";
import { formatCount } from "@/shared/utils/format";
import { MS_PER_MINUTE } from "@/shared/utils/time";

const STATUS_OPTIONS: readonly {
  value: SpanStatusFilter;
  label: string;
  tone?: "error" | "ok";
}[] = [
  { value: "all", label: "All" },
  { value: "error", label: "Errors", tone: "error" },
  { value: "ok", label: "OK", tone: "ok" },
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
const live = ref(false);

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
} = useSpansQuery(listParams, live);
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
  const day = matchDayPreset(
    filters.value.start,
    filters.value.end,
    PRESET_MATCH_TOLERANCE_MIN * MS_PER_MINUTE,
  );
  if (day !== undefined) return day;
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

const facetsActive = computed(
  () =>
    filters.value.search !== "" ||
    filters.value.status !== "all" ||
    filters.value.kind !== "all" ||
    filters.value.service !== "all",
);

const activeChips = computed((): ActiveFilterChip[] => {
  const chips: ActiveFilterChip[] = [];
  if (filters.value.status !== "all")
    chips.push({
      key: "status",
      label: "Status",
      value: filters.value.status === "error" ? "Errors" : "OK",
    });
  if (filters.value.kind !== "all")
    chips.push({ key: "kind", label: "Kind", value: filters.value.kind });
  if (filters.value.service !== "all")
    chips.push({
      key: "service",
      label: "Service",
      value: filters.value.service,
    });
  if (filters.value.search !== "")
    chips.push({ key: "search", label: "Name", value: filters.value.search });
  return chips;
});

const resultCount = computed(() => {
  const count = total.value;
  if (count === null) return null;
  return `${formatCount(count)} ${count === 1 ? "span" : "spans"}`;
});

const emptyTitle = computed(() =>
  filtersActive.value ? "No spans found" : "No spans recorded yet",
);

const emptyMessage = computed(() =>
  filtersActive.value
    ? filters.value.search !== ""
      ? `No spans match “${filters.value.search}” with these filters. Try a shorter name or widen the time range.`
      : "No spans match these filters. Try widening the time range or clearing a filter."
    : undefined,
);

const emptyActionLabel = computed(() =>
  filtersActive.value ? "Clear all filters" : undefined,
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
  const day = dayPresetRange(value);
  if (day !== undefined) {
    setTimeRange(day.start, day.end);
    return;
  }
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

function clearFacets(): void {
  searchInput.value = "";
  debouncedQuery.value = "";
  const next: SpansFilters = {
    search: "",
    status: "all",
    kind: "all",
    service: "all",
  };
  if (filters.value.start !== undefined) next.start = filters.value.start;
  if (filters.value.end !== undefined) next.end = filters.value.end;
  filters.value = next;
  resetPage();
}

function removeChip(key: string): void {
  if (key === "status") filters.value = { ...filters.value, status: "all" };
  else if (key === "kind") filters.value = { ...filters.value, kind: "all" };
  else if (key === "service")
    filters.value = { ...filters.value, service: "all" };
  else if (key === "search") {
    searchInput.value = "";
    debouncedQuery.value = "";
    filters.value = { ...filters.value, search: "" };
  } else return;
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
    <FilterBar label="Filter spans">
      <div class="flex flex-wrap items-center gap-2.5">
        <div class="min-w-40 flex-1">
          <SearchBox
            v-model="searchInput"
            input-id="span-search"
            label="Filter by span name"
            placeholder="Filter by span name, e.g. GET /cart"
          />
        </div>
        <LiveToggle v-model="live" @update:model-value="resetPage" />
        <DateRangePicker
          :presets="TIME_PRESETS"
          :day-presets="DAY_PRESETS"
          :active="activePreset"
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
          group-label="Span status"
          @update:model-value="resetPage"
        />
        <Select
          v-model="filters.kind"
          label="Kind"
          :options="KIND_OPTIONS"
          @update:model-value="resetPage"
        />
        <ServiceSelect
          v-model="filters.service"
          label="Service"
          :options="serviceOptions"
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
        :title="emptyTitle"
        :message="emptyMessage"
        :action-label="emptyActionLabel"
        @clear="clearFacets"
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
      :total="total"
      :has-prev="hasPrev"
      :has-next="hasNext"
      item-name="spans"
      @prev="prevPage"
      @next="nextPage"
    />
  </div>
</template>
