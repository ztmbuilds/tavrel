import type { RouteLocationNormalizedLoaded } from "vue-router";

import { readQueryParam } from "@/shared/composables/useRouteFilters";
import {
  ALL_TIME,
  TIME_PRESETS,
  dayPresetRange,
  defaultWindowStart,
} from "@/shared/utils/date-range";
import { MS_PER_MINUTE } from "@/shared/utils/time";

export const RANGE_ALL = ALL_TIME;
export type TimeRangeValue = typeof ALL_TIME;

export interface SharedTime {
  start?: string;
  end?: string;
  range?: TimeRangeValue;
  preset?: string;
}

const STORAGE_KEY = "observe-time-range";

function isValidIso(value: string): boolean {
  return value.includes("T") && !Number.isNaN(Date.parse(value));
}

function takeValidIso(value: string | undefined): string | undefined {
  if (value === undefined || !isValidIso(value)) return undefined;
  return value;
}

export function readTimeFromRoute(
  route: RouteLocationNormalizedLoaded,
): SharedTime | undefined {
  if (readQueryParam(route, "range") === RANGE_ALL) return { range: RANGE_ALL };
  const start = takeValidIso(readQueryParam(route, "start"));
  const end = takeValidIso(readQueryParam(route, "end"));
  if (start === undefined && end === undefined) return undefined;
  const time: SharedTime = {};
  if (start !== undefined) time.start = start;
  if (end !== undefined) time.end = end;
  return time;
}

function readStoredTime(): SharedTime | undefined {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw === null) return undefined;
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== "object" || parsed === null) return undefined;
    const record = parsed as Record<string, unknown>;
    if (record["range"] === RANGE_ALL) return { range: RANGE_ALL };
    const time: SharedTime = {};
    let hasValue = false;
    const storedStart = record["start"];
    const storedEnd = record["end"];
    if (typeof storedStart === "string" && isValidIso(storedStart)) {
      time.start = storedStart;
      hasValue = true;
    }
    if (typeof storedEnd === "string" && isValidIso(storedEnd)) {
      time.end = storedEnd;
      hasValue = true;
    }
    const storedPreset = record["preset"];
    if (
      hasValue &&
      typeof storedPreset === "string" &&
      (TIME_PRESETS.some((item) => item.value === storedPreset) ||
        dayPresetRange(storedPreset) !== undefined)
    )
      time.preset = storedPreset;
    return hasValue ? time : undefined;
  } catch {
    return undefined;
  }
}

function refreshStoredTime(stored: SharedTime): SharedTime {
  if (stored.preset === undefined) return stored;
  const relative = TIME_PRESETS.find((item) => item.value === stored.preset);
  if (relative !== undefined)
    return {
      start: new Date(Date.now() - relative.minutes * MS_PER_MINUTE).toISOString(),
      preset: stored.preset,
    };
  const day = dayPresetRange(stored.preset);
  if (day !== undefined) {
    const refreshed: SharedTime = { preset: stored.preset };
    if (day.start !== undefined) refreshed.start = day.start;
    if (day.end !== undefined) refreshed.end = day.end;
    return refreshed;
  }
  return stored;
}

export function resolveInitialTime(
  route: RouteLocationNormalizedLoaded,
): SharedTime {
  const fromRoute = readTimeFromRoute(route);
  if (fromRoute !== undefined) return fromRoute;
  const stored = readStoredTime();
  if (stored !== undefined) return refreshStoredTime(stored);
  return { start: defaultWindowStart() };
}

export function persistTime(time: SharedTime): void {
  try {
    if (time.range === RANGE_ALL) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ range: RANGE_ALL }));
      return;
    }
    if (time.start === undefined && time.end === undefined) return;
    const payload: Record<string, string> = {};
    if (time.start !== undefined) payload["start"] = time.start;
    if (time.end !== undefined) payload["end"] = time.end;
    if (time.preset !== undefined) payload["preset"] = time.preset;
    localStorage.setItem(STORAGE_KEY, JSON.stringify(payload));
  } catch {
    return;
  }
}

export function timeToQuery(time: SharedTime): Record<string, string> {
  if (time.range === RANGE_ALL) return { range: RANGE_ALL };
  const query: Record<string, string> = {};
  if (time.start !== undefined) query["start"] = time.start;
  if (time.end !== undefined) query["end"] = time.end;
  return query;
}

export function extractTime(time: SharedTime): SharedTime {
  const extracted: SharedTime = {};
  if (time.range === RANGE_ALL) {
    extracted.range = RANGE_ALL;
    return extracted;
  }
  if (time.start !== undefined) extracted.start = time.start;
  if (time.end !== undefined) extracted.end = time.end;
  if (time.preset !== undefined) extracted.preset = time.preset;
  return extracted;
}

export function preserveTimeQuery(
  query: RouteLocationNormalizedLoaded["query"],
): Record<string, string> {
  const range = query["range"];
  if (range === RANGE_ALL) return { range: RANGE_ALL };
  const preserved: Record<string, string> = {};
  const start = query["start"];
  const end = query["end"];
  if (typeof start === "string" && isValidIso(start)) preserved["start"] = start;
  if (typeof end === "string" && isValidIso(end)) preserved["end"] = end;
  return preserved;
}
