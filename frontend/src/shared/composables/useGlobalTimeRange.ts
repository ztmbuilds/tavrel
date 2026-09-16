import type { RouteLocationNormalizedLoaded } from "vue-router";

import { readQueryParam } from "@/shared/composables/useRouteFilters";
import {
  ALL_TIME,
  CUSTOM_RANGE,
  DEFAULT_WINDOW_MINUTES,
  DEFAULT_WINDOW_PRESET,
  TIME_PRESETS,
  dayPresetRange,
} from "@/shared/utils/date-range";
import { MS_PER_MINUTE } from "@/shared/utils/time";

export interface SharedTime {
  range: string;
  start?: string;
  end?: string;
}

export interface ResolvedWindow {
  start?: string;
  end?: string;
}

const STORAGE_KEY = "observe-time-range";

function isValidIso(value: string): boolean {
  return value.includes("T") && !Number.isNaN(Date.parse(value));
}

function isKnownRange(value: string): boolean {
  return (
    value === ALL_TIME ||
    value === CUSTOM_RANGE ||
    TIME_PRESETS.some((item) => item.value === value) ||
    dayPresetRange(value) !== undefined
  );
}

function takeCustomBounds(record: {
  start?: unknown;
  end?: unknown;
}): SharedTime | undefined {
  const time: SharedTime = { range: CUSTOM_RANGE };
  let hasBound = false;
  if (typeof record.start === "string" && isValidIso(record.start)) {
    time.start = record.start;
    hasBound = true;
  }
  if (typeof record.end === "string" && isValidIso(record.end)) {
    time.end = record.end;
    hasBound = true;
  }
  return hasBound ? time : undefined;
}

export function readTimeFromRoute(
  route: RouteLocationNormalizedLoaded,
): SharedTime | undefined {
  const range = readQueryParam(route, "range");
  if (range === undefined || !isKnownRange(range)) return undefined;
  if (range !== CUSTOM_RANGE) return { range };
  return takeCustomBounds({
    start: readQueryParam(route, "start"),
    end: readQueryParam(route, "end"),
  });
}

function readStoredTime(): SharedTime | undefined {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw === null) return undefined;
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== "object" || parsed === null) return undefined;
    const record = parsed as Record<string, unknown>;
    const range = record["range"];
    if (typeof range !== "string" || !isKnownRange(range)) return undefined;
    if (range !== CUSTOM_RANGE) return { range };
    return takeCustomBounds({
      start: record["start"],
      end: record["end"],
    });
  } catch {
    return undefined;
  }
}

export function resolveInitialTime(
  route: RouteLocationNormalizedLoaded,
): SharedTime {
  const fromRoute = readTimeFromRoute(route);
  if (fromRoute !== undefined) return fromRoute;
  const stored = readStoredTime();
  if (stored !== undefined) return stored;
  return { range: DEFAULT_WINDOW_PRESET };
}

export function resolveWindow(
  time: SharedTime,
  now: number = Date.now(),
): ResolvedWindow {
  if (time.range === ALL_TIME) return {};
  if (time.range === CUSTOM_RANGE) {
    const window: ResolvedWindow = {};
    if (time.start !== undefined) window.start = time.start;
    if (time.end !== undefined) window.end = time.end;
    return window;
  }
  const relative = TIME_PRESETS.find((item) => item.value === time.range);
  if (relative !== undefined)
    return {
      start: new Date(now - relative.minutes * MS_PER_MINUTE).toISOString(),
    };
  const day = dayPresetRange(time.range);
  if (day !== undefined) {
    const window: ResolvedWindow = {};
    if (day.start !== undefined) window.start = day.start;
    if (day.end !== undefined) window.end = day.end;
    return window;
  }
  return {
    start: new Date(now - DEFAULT_WINDOW_MINUTES * MS_PER_MINUTE).toISOString(),
  };
}

export function persistTime(time: SharedTime): void {
  try {
    if (time.range !== CUSTOM_RANGE) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ range: time.range }));
      return;
    }
    const custom = takeCustomBounds({ start: time.start, end: time.end });
    if (custom === undefined) return;
    localStorage.setItem(STORAGE_KEY, JSON.stringify(custom));
  } catch {
    return;
  }
}

export function timeToQuery(time: SharedTime): Record<string, string> {
  if (time.range !== CUSTOM_RANGE) return { range: time.range };
  const query: Record<string, string> = { range: CUSTOM_RANGE };
  if (time.start !== undefined) query["start"] = time.start;
  if (time.end !== undefined) query["end"] = time.end;
  return query;
}

export function extractTime(time: SharedTime): SharedTime {
  if (time.range !== CUSTOM_RANGE) return { range: time.range };
  const extracted: SharedTime = { range: CUSTOM_RANGE };
  if (time.start !== undefined) extracted.start = time.start;
  if (time.end !== undefined) extracted.end = time.end;
  return extracted;
}

export function preserveTimeQuery(
  query: RouteLocationNormalizedLoaded["query"],
): Record<string, string> {
  const range = query["range"];
  if (typeof range !== "string" || !isKnownRange(range)) return {};
  if (range !== CUSTOM_RANGE) return { range };
  const preserved: Record<string, string> = { range };
  const start = query["start"];
  const end = query["end"];
  if (typeof start === "string" && isValidIso(start)) preserved["start"] = start;
  if (typeof end === "string" && isValidIso(end)) preserved["end"] = end;
  if (preserved["start"] === undefined && preserved["end"] === undefined)
    return {};
  return preserved;
}
