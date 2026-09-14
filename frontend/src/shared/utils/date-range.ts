import { MS_PER_MINUTE } from "@/shared/utils/time";

export interface TimePreset {
  value: string;
  label: string;
  minutes: number;
}

export const TIME_PRESETS: readonly TimePreset[] = [
  { value: "15m", label: "15m", minutes: 15 },
  { value: "1h", label: "1h", minutes: 60 },
  { value: "6h", label: "6h", minutes: 360 },
  { value: "24h", label: "24h", minutes: 1440 },
  { value: "7d", label: "7d", minutes: 10080 },
];

export const CUSTOM_RANGE = "custom";
export const ALL_TIME = "all";

export const DEFAULT_WINDOW_PRESET = "24h";
export const DEFAULT_WINDOW_MINUTES = 1440;

export function defaultWindowStart(now: number = Date.now()): string {
  return new Date(now - DEFAULT_WINDOW_MINUTES * MS_PER_MINUTE).toISOString();
}

export const TODAY_PRESET = "today";
export const YESTERDAY_PRESET = "yesterday";

export interface DayPreset {
  value: string;
  label: string;
}

export const DAY_PRESETS: readonly DayPreset[] = [
  { value: TODAY_PRESET, label: "Today" },
  { value: YESTERDAY_PRESET, label: "Yesterday" },
];

const MS_PER_DAY = 86400000;

function startOfToday(): Date {
  const now = new Date();
  return new Date(now.getFullYear(), now.getMonth(), now.getDate());
}

export function dayPresetRange(value: string): { start?: string; end?: string } | undefined {
  if (value === TODAY_PRESET) return { start: startOfToday().toISOString() };
  if (value === YESTERDAY_PRESET) {
    const today = startOfToday();
    return {
      start: new Date(today.getTime() - MS_PER_DAY).toISOString(),
      end: today.toISOString(),
    };
  }
  return undefined;
}

export function matchDayPreset(
  start: string | undefined,
  end: string | undefined,
  toleranceMs: number,
): string | undefined {
  if (start === undefined) return undefined;
  const todayMs = startOfToday().getTime();
  if (Math.abs(Date.parse(start) - todayMs) < toleranceMs && end === undefined)
    return TODAY_PRESET;
  if (
    end !== undefined &&
    Math.abs(Date.parse(start) - (todayMs - MS_PER_DAY)) < toleranceMs &&
    Math.abs(Date.parse(end) - todayMs) < toleranceMs
  )
    return YESTERDAY_PRESET;
  return undefined;
}
