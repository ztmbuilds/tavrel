<script setup lang="ts">
import { computed, ref, watch } from "vue";

import calendarIcon from "@/assets/icons/calendar.svg";
import chevronLeftIcon from "@/assets/icons/chevron-left.svg";
import chevronRightIcon from "@/assets/icons/chevron-right.svg";
import Button from "@/shared/components/ui/Button.vue";
import TimeInput from "@/shared/components/TimeInput.vue";
import { cn } from "@/shared/utils/cn";
import {
  civilToIso,
  compareCivilDate,
  formatCivilDate,
  formatCivilDateTime,
  isValidTimeText,
  isoToCivilDate,
  isoToCivilTime,
  parseTimeText,
  toTimeText,
  type CivilDate,
} from "@/shared/utils/time";

const props = defineProps<{
  presets: readonly DatePreset[];
  active: string;
  start: string | undefined;
  end: string | undefined;
}>();

const emit = defineEmits<{
  preset: [value: string];
  custom: [range: { start?: string; end?: string }];
}>();

export interface DatePreset {
  value: string;
  label: string;
  minutes: number;
}

const CUSTOM = "custom";
const DEFAULT_START_TEXT = "00:00";
const DEFAULT_END_TEXT = "23:59";
const WEEK_COLUMNS = 7;
const JANUARY_INDEX = 0;
const DECEMBER_INDEX = 11;
const MONTH_OVERFLOW_DAY = 0;

const WEEKDAY_LABELS: readonly string[] = [
  "Su",
  "Mo",
  "Tu",
  "We",
  "Th",
  "Fr",
  "Sa",
];

const dateOpen = ref(false);
const customOpen = ref(false);
const viewYear = ref(0);
const viewMonth = ref(0);
const selectedStart = ref<CivilDate | null>(null);
const selectedEnd = ref<CivilDate | null>(null);
const startTimeText = ref(DEFAULT_START_TEXT);
const endTimeText = ref(DEFAULT_END_TEXT);

const customSummary = computed(() => {
  const startLabel =
    props.start === undefined ? "…" : formatCivilDateTime(props.start);
  const endLabel =
    props.end === undefined ? "…" : formatCivilDateTime(props.end);
  return `${startLabel} – ${endLabel}`;
});

const dateLabel = computed(() => {
  if (props.active === CUSTOM) return customSummary.value;
  const preset = props.presets.find((item) => item.value === props.active);
  return preset === undefined ? "Last 1h" : `Last ${preset.label}`;
});

const monthLabel = computed(() =>
  new Date(viewYear.value, viewMonth.value, 1).toLocaleDateString(undefined, {
    month: "long",
    year: "numeric",
  }),
);

function daysInMonth(year: number, month: number): number {
  return new Date(year, month + 1, MONTH_OVERFLOW_DAY).getDate();
}

const monthWeeks = computed((): (CivilDate | null)[][] => {
  const cells: (CivilDate | null)[] = [];
  const leadingBlanks = new Date(viewYear.value, viewMonth.value, 1).getDay();
  for (let i = 0; i < leadingBlanks; i += 1) cells.push(null);
  const totalDays = daysInMonth(viewYear.value, viewMonth.value);
  for (let day = 1; day <= totalDays; day += 1) {
    cells.push({ year: viewYear.value, month: viewMonth.value + 1, day });
  }
  while (cells.length % WEEK_COLUMNS !== 0) cells.push(null);
  const weeks: (CivilDate | null)[][] = [];
  for (let w = 0; w < cells.length; w += WEEK_COLUMNS) {
    weeks.push(cells.slice(w, w + WEEK_COLUMNS));
  }
  return weeks;
});

function isSameCivilDate(a: CivilDate, b: CivilDate): boolean {
  return compareCivilDate(a, b) === 0;
}

function isSelectedDate(date: CivilDate): boolean {
  return (
    (selectedStart.value !== null &&
      isSameCivilDate(date, selectedStart.value)) ||
    (selectedEnd.value !== null && isSameCivilDate(date, selectedEnd.value))
  );
}

function isInRangeDate(date: CivilDate): boolean {
  if (selectedStart.value === null || selectedEnd.value === null) return false;
  return (
    compareCivilDate(date, selectedStart.value) >= 0 &&
    compareCivilDate(date, selectedEnd.value) <= 0
  );
}

function isTodayDate(date: CivilDate): boolean {
  const now = new Date();
  return (
    date.year === now.getFullYear() &&
    date.month === now.getMonth() + 1 &&
    date.day === now.getDate()
  );
}

function dayLabel(date: CivilDate): string {
  return formatCivilDate(date);
}

function selectDate(date: CivilDate): void {
  if (selectedStart.value === null || selectedEnd.value !== null) {
    selectedStart.value = date;
    selectedEnd.value = null;
    return;
  }
  if (compareCivilDate(date, selectedStart.value) < 0) {
    selectedEnd.value = selectedStart.value;
    selectedStart.value = date;
  } else {
    selectedEnd.value = date;
  }
}

function prevMonth(): void {
  if (viewMonth.value === JANUARY_INDEX) {
    viewMonth.value = DECEMBER_INDEX;
    viewYear.value -= 1;
  } else {
    viewMonth.value -= 1;
  }
}

function nextMonth(): void {
  if (viewMonth.value === DECEMBER_INDEX) {
    viewMonth.value = JANUARY_INDEX;
    viewYear.value += 1;
  } else {
    viewMonth.value += 1;
  }
}

function showMonthContaining(date: CivilDate): void {
  viewYear.value = date.year;
  viewMonth.value = date.month - 1;
}

function resetCustomInputs(): void {
  selectedStart.value = null;
  selectedEnd.value = null;
  startTimeText.value = DEFAULT_START_TEXT;
  endTimeText.value = DEFAULT_END_TEXT;
}

function prefillInputs(): void {
  if (props.start !== undefined && props.end !== undefined) {
    selectedStart.value = isoToCivilDate(props.start);
    selectedEnd.value = isoToCivilDate(props.end);
    showMonthContaining(selectedStart.value);
    startTimeText.value = toTimeText(isoToCivilTime(props.start));
    endTimeText.value = toTimeText(isoToCivilTime(props.end));
  } else {
    const now = new Date();
    resetCustomInputs();
    viewYear.value = now.getFullYear();
    viewMonth.value = now.getMonth();
  }
}

watch(dateOpen, (open) => {
  if (open) {
    customOpen.value = props.active === CUSTOM;
    prefillInputs();
  }
});

function toggleDatePop(): void {
  dateOpen.value = !dateOpen.value;
}

function closeDatePop(): void {
  dateOpen.value = false;
}

function selectPreset(value: string): void {
  dateOpen.value = false;
  customOpen.value = false;
  emit("preset", value);
}

const canApply = computed(
  () =>
    selectedStart.value !== null &&
    selectedEnd.value !== null &&
    isValidTimeText(startTimeText.value) &&
    isValidTimeText(endTimeText.value),
);

function applyCustom(): void {
  const start = selectedStart.value;
  const end = selectedEnd.value;
  if (start === null || end === null) return;
  if (
    !isValidTimeText(startTimeText.value) ||
    !isValidTimeText(endTimeText.value)
  )
    return;
  dateOpen.value = false;
  customOpen.value = false;
  emit("custom", {
    start: civilToIso(start, parseTimeText(startTimeText.value)),
    end: civilToIso(end, parseTimeText(endTimeText.value)),
  });
}

function cancelCustom(): void {
  customOpen.value = false;
  if (props.active !== CUSTOM) {
    resetCustomInputs();
  }
}
</script>

<template>
  <div class="relative">
    <button
      type="button"
      class="border-default bg-surface text-secondary hover:border-strong flex min-h-7 cursor-pointer items-center gap-1.5 rounded-md border px-2.5 py-1 font-sans text-xs whitespace-nowrap transition-colors duration-120 ease-out"
      :aria-expanded="dateOpen"
      aria-haspopup="dialog"
      aria-controls="span-date-pop"
      @click="toggleDatePop"
      @keydown.escape="closeDatePop"
    >
      <img :src="calendarIcon" alt="" class="size-3.5 shrink-0" />
      <span>{{ dateLabel }}</span>
      <span class="text-tertiary text-[0.625rem]" aria-hidden="true">▾</span>
    </button>
    <div
      v-if="dateOpen"
      id="span-date-pop"
      role="dialog"
      aria-label="Choose time range"
      class="border-default bg-raised absolute top-[calc(100%+0.375rem)] left-0 z-30 w-68 rounded-md border p-2 shadow-[0_0.25rem_0.75rem_rgba(0,0,0,0.4)]"
      @keydown.escape="closeDatePop"
    >
      <div
        class="grid grid-cols-5 gap-1"
        role="group"
        aria-label="Preset ranges"
      >
        <button
          v-for="preset in presets"
          :key="preset.value"
          type="button"
          :aria-pressed="active === preset.value"
          :class="
            cn(
              'border-subtle bg-surface text-secondary hover:text-primary min-h-[1.875rem] cursor-pointer rounded-md border px-1 py-1.5 font-sans text-xs transition-colors duration-120 ease-out',
              active === preset.value &&
                'bg-signal-muted text-signal border-transparent',
            )
          "
          @click="selectPreset(preset.value)"
        >
          {{ preset.label }}
        </button>
      </div>
      <button
        type="button"
        :aria-pressed="active === CUSTOM"
        :aria-expanded="customOpen"
        :class="
          cn(
            'text-tertiary hover:bg-surface hover:text-secondary mt-1 flex min-h-[1.875rem] w-full cursor-pointer items-center justify-center gap-1.5 rounded-md border border-transparent bg-transparent px-1 py-[0.4375rem] font-sans text-xs transition-colors duration-120 ease-out',
            active === CUSTOM && 'text-primary',
          )
        "
        @click="customOpen = !customOpen"
      >
        <span>{{
          active === CUSTOM && !customOpen ? customSummary : "Custom range"
        }}</span>
        <span
          class="text-[0.5625rem] transition-transform duration-120 ease-out"
          :class="customOpen && 'rotate-180'"
          aria-hidden="true"
          >▾</span
        >
      </button>
      <div v-if="customOpen" class="border-subtle mt-0.5 border-t pt-2.5">
        <div class="flex items-center justify-between">
          <Button
            variant="ghost"
            class="h-7 w-7 px-0"
            aria-label="Previous month"
            @click="prevMonth"
          >
            <img :src="chevronLeftIcon" alt="" class="size-3.5" />
          </Button>
          <span
            class="text-primary font-sans text-xs font-semibold"
            aria-live="polite"
            >{{ monthLabel }}</span
          >
          <Button
            variant="ghost"
            class="h-7 w-7 px-0"
            aria-label="Next month"
            @click="nextMonth"
          >
            <img :src="chevronRightIcon" alt="" class="size-3.5" />
          </Button>
        </div>
        <div class="mt-2" role="group" aria-label="Choose custom dates">
          <div class="grid grid-cols-7" aria-hidden="true">
            <span
              v-for="label in WEEKDAY_LABELS"
              :key="label"
              class="text-tertiary flex size-8 items-center justify-center text-[0.6875rem] font-medium"
            >
              {{ label }}
            </span>
          </div>
          <div
            v-for="(week, weekIndex) in monthWeeks"
            :key="`week-${weekIndex}`"
            class="grid grid-cols-7"
          >
            <template v-for="(date, dayIndex) in week" :key="`day-${dayIndex}`">
              <span v-if="date === null" class="size-8" />
              <button
                v-else
                type="button"
                :aria-pressed="isSelectedDate(date)"
                :aria-label="dayLabel(date)"
                :class="
                  cn(
                    'text-secondary hover:bg-hover hover:text-primary flex size-8 cursor-pointer items-center justify-center rounded-md border-none bg-transparent p-0 font-sans text-xs transition-colors duration-120 ease-out',
                    isInRangeDate(date) && 'bg-signal-muted/60 text-primary',
                    isSelectedDate(date) && 'bg-signal-muted text-signal',
                    isTodayDate(date) &&
                      !isSelectedDate(date) &&
                      'ring-strong ring-1 ring-inset',
                  )
                "
                @click="selectDate(date)"
              >
                {{ date.day }}
              </button>
            </template>
          </div>
        </div>
        <div class="mt-2 flex gap-2">
          <div class="flex min-w-0 flex-1 flex-col gap-1">
            <label for="span-start-time" class="text-tertiary text-[0.6875rem]"
              >Start time</label
            >
            <TimeInput
              v-model="startTimeText"
              input-id="span-start-time"
              label="Custom range start time"
            />
          </div>
          <div class="flex min-w-0 flex-1 flex-col gap-1">
            <label for="span-end-time" class="text-tertiary text-[0.6875rem]"
              >End time</label
            >
            <TimeInput
              v-model="endTimeText"
              input-id="span-end-time"
              label="Custom range end time"
            />
          </div>
        </div>
        <div class="mt-2.5 flex gap-1.5">
          <Button variant="secondary" @click="cancelCustom">Cancel</Button>
          <Button class="flex-1" :disabled="!canApply" @click="applyCustom">
            Apply range
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
