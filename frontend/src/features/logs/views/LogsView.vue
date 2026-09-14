<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import arrowUpIcon from "@/assets/icons/arrow-up.svg";
import refreshIcon from "@/assets/icons/refresh.svg";

import LogRow from "@/features/logs/components/LogRow.vue";
import {
  useInfiniteLogsQuery,
  useLogsServicesQuery,
} from "@/features/logs/queries";
import { type LogsFilters, type LogLevelFilter } from "@/features/logs/types";
import { useLogsInfiniteScroll } from "@/features/logs/composables/useLogsInfiniteScroll";
import ActiveFilterChips, {
  type ActiveFilterChip,
} from "@/shared/components/ActiveFilterChips.vue";
import ChipGroup from "@/shared/components/ChipGroup.vue";
import DateRangePicker from "@/shared/components/DateRangePicker.vue";
import EmptyState from "@/shared/components/EmptyState.vue";
import FilterBar from "@/shared/components/FilterBar.vue";
import LiveToggle from "@/shared/components/LiveToggle.vue";
import PageHead from "@/shared/components/PageHead.vue";
import SearchBox from "@/shared/components/SearchBox.vue";
import ServiceSelect from "@/shared/components/ServiceSelect.vue";
import Button from "@/shared/components/ui/Button.vue";
import {
  readQueryParam,
  useDebouncedSearch,
  useRouteQuerySync,
} from "@/shared/composables/useRouteFilters";
import {
  ALL_TIME,
  CUSTOM_RANGE,
  DAY_PRESETS,
  TIME_PRESETS,
  dayPresetRange,
  defaultWindowStart,
  matchDayPreset,
} from "@/shared/utils/date-range";
import { formatCount } from "@/shared/utils/format";
import { formatDateTimeRange, MS_PER_MINUTE } from "@/shared/utils/time";

const LEVEL_OPTIONS: readonly {
  value: LogLevelFilter;
  label: string;
  tone?: "info" | "warn" | "error";
}[] = [
  { value: "all", label: "All" },
  { value: "info", label: "Info", tone: "info" },
  { value: "warn", label: "Warn", tone: "warn" },
  { value: "error", label: "Error", tone: "error" },
];

const SEARCH_DEBOUNCE_MS = 300;
const PRESET_MATCH_TOLERANCE_MIN = 1;
const BACK_TO_TOP_THRESHOLD_PX = 600;

const route = useRoute();
const router = useRouter();

function filtersFromRoute(): LogsFilters {
  const levelRaw = readQueryParam(route, "level");
  const next: LogsFilters = {
    search: readQueryParam(route, "search") ?? "",
    level:
      levelRaw === "info" ||
      levelRaw === "warn" ||
      levelRaw === "error" ||
      levelRaw === "all"
        ? levelRaw
        : "all",
    service: readQueryParam(route, "service") ?? "all",
    traceId: readQueryParam(route, "traceId") ?? "",
    spanId: readQueryParam(route, "spanId") ?? "",
  };
  const start = readQueryParam(route, "start");
  const end = readQueryParam(route, "end");
  if (start !== undefined) next.start = start;
  if (end !== undefined) next.end = end;
  if (start === undefined && end === undefined) next.start = defaultWindowStart();
  return next;
}

const filters = ref<LogsFilters>(filtersFromRoute());
const searchInput = ref(filters.value.search);
const debouncedQuery = useDebouncedSearch(searchInput, SEARCH_DEBOUNCE_MS);
const live = ref(false);
const expanded = reactive(new Set<number>());
const scrollerEl = ref<Element | null>(null);
const sentinelEl = ref<Element | null>(null);
const showBackToTop = ref(false);

watch(debouncedQuery, (value) => {
  if (value === filters.value.search) return;
  filters.value = { ...filters.value, search: value };
  resetList();
});

function filtersToQuery(): Record<string, string> {
  const query: Record<string, string> = {};
  if (filters.value.search !== "") query.search = filters.value.search;
  if (filters.value.level !== "all") query.level = filters.value.level;
  if (filters.value.service !== "all") query.service = filters.value.service;
  if (filters.value.traceId !== "") query.traceId = filters.value.traceId;
  if (filters.value.spanId !== "") query.spanId = filters.value.spanId;
  if (filters.value.start !== undefined) query.start = filters.value.start;
  if (filters.value.end !== undefined) query.end = filters.value.end;
  return query;
}

useRouteQuerySync(filters, filtersToQuery);

const {
  data: logsData,
  isPending,
  isError,
  refetch,
  isFetching,
  fetchNextPage,
  hasNextPage,
  isFetchingNextPage,
  isFetchNextPageError,
} = useInfiniteLogsQuery(filters, live);
const servicesQuery = useLogsServicesQuery();

const logs = computed(
  () => logsData.value?.pages.flatMap((page) => page.logs) ?? [],
);
const total = computed(() => logsData.value?.pages[0]?.total ?? null);

const canLoadMore = computed(
  () => hasNextPage.value && !isFetchingNextPage.value,
);

function loadMore(): void {
  if (!canLoadMore.value) return;
  void fetchNextPage();
}

useLogsInfiniteScroll({
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
    return ALL_TIME;
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
    filters.value.level !== "all" ||
    filters.value.service !== "all" ||
    filters.value.traceId !== "" ||
    filters.value.spanId !== "" ||
    filters.value.start !== undefined ||
    filters.value.end !== undefined,
);

const facetsActive = computed(
  () =>
    filters.value.search !== "" ||
    filters.value.level !== "all" ||
    filters.value.service !== "all" ||
    filters.value.traceId !== "" ||
    filters.value.spanId !== "",
);

const LEVEL_LABELS: Record<LogLevelFilter, string> = {
  all: "All",
  info: "Info",
  warn: "Warn",
  error: "Error",
};

const activeChips = computed((): ActiveFilterChip[] => {
  const chips: ActiveFilterChip[] = [];
  if (filters.value.level !== "all")
    chips.push({
      key: "level",
      label: "Level",
      value: LEVEL_LABELS[filters.value.level],
    });
  if (filters.value.service !== "all")
    chips.push({
      key: "service",
      label: "Service",
      value: filters.value.service,
    });
  if (filters.value.traceId !== "")
    chips.push({
      key: "traceId",
      label: "Trace",
      value: filters.value.traceId,
    });
  if (filters.value.spanId !== "")
    chips.push({ key: "spanId", label: "Span", value: filters.value.spanId });
  if (filters.value.search !== "")
    chips.push({
      key: "search",
      label: "Message",
      value: filters.value.search,
    });
  if (filters.value.start !== undefined || filters.value.end !== undefined) {
    const preset = TIME_PRESETS.find(
      (item) => item.value === activePreset.value,
    );
    const day = DAY_PRESETS.find((item) => item.value === activePreset.value);
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
  return `${formatCount(count)} ${count === 1 ? "log line" : "log lines"}`;
});

const emptyTitle = computed(() =>
  filtersActive.value ? "No log lines found" : "No logs recorded yet",
);

const emptyMessage = computed(() => {
  if (!filtersActive.value) return undefined;
  if (filters.value.search !== "")
    return `No log lines match “${filters.value.search}” with these filters. Try a shorter search or widen the time range.`;
  return "No log lines match these filters. Try widening the time range or clearing a filter.";
});

const emptyActionLabel = computed(() =>
  filtersActive.value ? "Clear all filters" : undefined,
);

const loadedCount = computed(() => {
  if (total.value === null) return null;
  return `${formatCount(logs.value.length)} of ${formatCount(total.value)} logs`;
});

const endOfListLabel = computed(() => {
  const count = total.value;
  if (count === null) return "No more logs";
  return count === 1
    ? "Showing the only log"
    : `Showing all ${formatCount(count)} logs`;
});

function resetList(): void {
  expanded.clear();
  showBackToTop.value = false;
  scrollerEl.value?.scrollTo({ top: 0 });
}

function onScrollerScroll(event: Event): void {
  if (!(event.target instanceof HTMLElement)) return;
  showBackToTop.value = event.target.scrollTop > BACK_TO_TOP_THRESHOLD_PX;
}

function scrollToTop(): void {
  scrollerEl.value?.scrollTo({ top: 0, behavior: "smooth" });
}

function setTimeRange(
  start: string | undefined,
  end: string | undefined,
): void {
  const next: LogsFilters = {
    search: filters.value.search,
    level: filters.value.level,
    service: filters.value.service,
    traceId: filters.value.traceId,
    spanId: filters.value.spanId,
  };
  if (start !== undefined) next.start = start;
  if (end !== undefined) next.end = end;
  filters.value = next;
  resetList();
}

function applyPreset(value: string): void {
  if (value === ALL_TIME) {
    setTimeRange(undefined, undefined);
    return;
  }
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
  const next: LogsFilters = {
    search: "",
    level: "all",
    service: "all",
    traceId: "",
    spanId: "",
  };
  if (filters.value.start !== undefined) next.start = filters.value.start;
  if (filters.value.end !== undefined) next.end = filters.value.end;
  filters.value = next;
  resetList();
}

function clearAllFilters(): void {
  searchInput.value = "";
  debouncedQuery.value = "";
  filters.value = {
    search: "",
    level: "all",
    service: "all",
    traceId: "",
    spanId: "",
    start: defaultWindowStart(),
  };
  resetList();
}

function removeChip(key: string): void {
  if (key === "level") filters.value = { ...filters.value, level: "all" };
  else if (key === "service")
    filters.value = { ...filters.value, service: "all" };
  else if (key === "traceId") filters.value = { ...filters.value, traceId: "" };
  else if (key === "spanId") filters.value = { ...filters.value, spanId: "" };
  else if (key === "timerange") setTimeRange(undefined, undefined);
  else if (key === "search") {
    searchInput.value = "";
    debouncedQuery.value = "";
    filters.value = { ...filters.value, search: "" };
  } else return;
  resetList();
}

function toggleLog(id: number): void {
  if (expanded.has(id)) expanded.delete(id);
  else expanded.add(id);
}

function openTrace(traceId: string): void {
  void router.push({ name: "trace-detail", params: { id: traceId } });
}
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <PageHead
      title="Logs"
      subtitle="Log lines, linked to the trace and span that produced them."
    />
    <FilterBar label="Filter logs">
      <div class="flex flex-wrap items-center gap-2.5">
        <div class="min-w-40 flex-1">
          <SearchBox
            v-model="searchInput"
            input-id="log-search"
            label="Search log messages"
            placeholder="Search log messages, e.g. card_declined"
          />
        </div>
        <LiveToggle v-model="live" @update:model-value="resetList" />
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
          v-model="filters.level"
          :options="LEVEL_OPTIONS"
          group-label="Log level"
          @update:model-value="resetList"
        />
        <ServiceSelect
          v-model="filters.service"
          label="Service"
          :options="serviceOptions"
          trigger-class="max-w-64"
          @update:model-value="resetList"
        />
        <div class="ml-auto flex shrink-0 items-center gap-3">
          <button
            v-if="facetsActive"
            type="button"
            class="text-secondary hover:text-primary decoration-strong shrink-0 cursor-pointer border-none bg-transparent p-0 text-[0.78125rem] underline underline-offset-2 transition-colors duration-120 ease-out"
            @click="clearFacets"
          >
            Clear all
          </button>
          <img
            v-if="isFetching && !isPending"
            :src="refreshIcon"
            alt=""
            aria-hidden="true"
            class="size-3.5 shrink-0 animate-spin opacity-70"
          />
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
    <div class="relative mx-4 mb-3 min-h-0 flex-1">
      <div
        id="log-list"
        ref="scrollerEl"
        class="border-subtle bg-panel h-full overflow-y-auto rounded-xl border"
        role="list"
        aria-label="Logs"
        tabindex="-1"
        @scroll="onScrollerScroll"
      >
        <div
          v-if="isPending"
          class="text-tertiary flex flex-col items-center gap-1.5 px-5 py-12 text-center text-[0.8125rem]"
          role="status"
        >
          <span class="text-heading text-sm font-semibold">Loading logs…</span>
        </div>
        <div
          v-else-if="isError"
          class="text-tertiary flex flex-col items-center gap-3.5 px-5 py-12 text-center text-[0.8125rem]"
          role="alert"
        >
          <span class="text-heading text-sm font-semibold"
            >Couldn’t load logs</span
          >
          <span class="leading-relaxed"
            >Something went wrong fetching logs. Try again.</span
          >
          <Button variant="secondary" @click="refetch()">Retry</Button>
        </div>
        <EmptyState
          v-else-if="logs.length === 0"
          :title="emptyTitle"
          :message="emptyMessage"
          :action-label="emptyActionLabel"
          @clear="clearAllFilters"
        />
        <template v-else>
          <div
            class="border-default bg-panel sticky top-0 z-5 flex items-center gap-2.5 border-b px-4 py-1.5"
            aria-hidden="true"
          >
            <div class="w-[0.1875rem] shrink-0" />
            <div
              class="text-tertiary w-10 shrink-0 text-[0.65625rem] leading-relaxed font-semibold tracking-wider uppercase max-[40rem]:hidden"
            >
              Level
            </div>
            <div
              class="text-tertiary w-[10.75rem] shrink-0 text-[0.65625rem] leading-relaxed font-semibold tracking-wider uppercase max-[40rem]:hidden"
            >
              Time
            </div>
            <div
              class="text-tertiary w-[7.375rem] shrink-0 text-[0.65625rem] leading-relaxed font-semibold tracking-wider uppercase"
            >
              Service
            </div>
            <div
              class="text-tertiary min-w-0 flex-1 truncate text-[0.65625rem] leading-relaxed font-semibold tracking-wider uppercase"
            >
              Message
            </div>
            <div
              class="text-tertiary w-[4.625rem] shrink-0 text-right text-[0.65625rem] leading-relaxed font-semibold tracking-wider uppercase max-[56.25rem]:hidden"
            >
              Trace
            </div>
          </div>
          <LogRow
            v-for="log in logs"
            :key="log.id"
            :log="log"
            :expanded="expanded.has(log.id)"
            @toggle="toggleLog"
            @open-trace="openTrace"
          />
          <div
            ref="sentinelEl"
            class="flex flex-col items-center gap-2 px-5 py-4 text-center"
          >
            <template v-if="isFetchingNextPage">
              <img
                :src="refreshIcon"
                alt=""
                aria-hidden="true"
                class="size-3.5 shrink-0 animate-spin opacity-70"
              />
              <span class="text-tertiary text-[0.78125rem]" role="status"
                >Loading more logs…</span
              >
            </template>
            <template v-else-if="isFetchNextPageError">
              <span class="text-tertiary text-[0.78125rem]" role="alert"
                >Couldn’t load more logs.</span
              >
              <Button variant="secondary" @click="loadMore()">Retry</Button>
            </template>
            <template v-else-if="hasNextPage">
              <span
                v-if="loadedCount !== null"
                class="text-tertiary text-[0.78125rem] tabular-nums"
                role="status"
                >{{ loadedCount }}</span
              >
              <Button variant="secondary" @click="loadMore()"
                >Load more logs</Button
              >
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
      <button
        v-show="showBackToTop"
        type="button"
        aria-label="Back to top of logs"
        title="Back to top"
        class="border-default bg-raised hover:border-strong hover:bg-hover absolute right-4 bottom-4 z-10 flex size-10 cursor-pointer items-center justify-center rounded-full border shadow-lg transition-colors duration-120 ease-out"
        @click="scrollToTop"
      >
        <img :src="arrowUpIcon" alt="" aria-hidden="true" class="size-4" />
      </button>
    </div>
  </div>
</template>
