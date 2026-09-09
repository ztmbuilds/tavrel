export interface CivilDate {
  year: number;
  month: number;
  day: number;
}

export interface CivilTime {
  hour: number;
  minute: number;
}

export const MS_PER_SECOND = 1000;
export const MS_PER_MINUTE = 60000;
const MINUTES_PER_HOUR = 60;
const HOURS_PER_DAY = 24;

function pad2(value: number): string {
  return String(value).padStart(2, "0");
}

export function civilToIso(date: CivilDate, time: CivilTime): string {
  return new Date(
    date.year,
    date.month - 1,
    date.day,
    time.hour,
    time.minute,
  ).toISOString();
}

export function isoToCivilDate(iso: string): CivilDate {
  const parsed = new Date(iso);
  return {
    year: parsed.getFullYear(),
    month: parsed.getMonth() + 1,
    day: parsed.getDate(),
  };
}

export function isoToCivilTime(iso: string): CivilTime {
  const parsed = new Date(iso);
  return { hour: parsed.getHours(), minute: parsed.getMinutes() };
}

export function formatCivilDateTime(iso: string): string {
  const date = isoToCivilDate(iso);
  const time = isoToCivilTime(iso);
  return `${date.year}-${pad2(date.month)}-${pad2(date.day)} ${pad2(time.hour)}:${pad2(time.minute)}`;
}

export function compareCivilDate(a: CivilDate, b: CivilDate): number {
  if (a.year !== b.year) return a.year - b.year;
  if (a.month !== b.month) return a.month - b.month;
  return a.day - b.day;
}

const TIME_TEXT_PATTERN = /^([01]\d|2[0-3]):([0-5]\d)$/;

export function isValidTimeText(value: string): boolean {
  return TIME_TEXT_PATTERN.test(value);
}

export function toTimeText(time: CivilTime): string {
  return `${pad2(time.hour)}:${pad2(time.minute)}`;
}

export function parseTimeText(value: string): CivilTime {
  const [hourText, minuteText] = value.split(":");
  return { hour: Number(hourText ?? ""), minute: Number(minuteText ?? "") };
}

export function formatCivilDate(date: CivilDate): string {
  return new Date(date.year, date.month - 1, date.day).toLocaleDateString(
    undefined,
    { year: "numeric", month: "long", day: "numeric" },
  );
}

export function formatDurationMs(ms: number): string {
  if (ms < MS_PER_SECOND) return `${Math.round(ms)}ms`;
  const seconds = ms / MS_PER_SECOND;
  return `${seconds.toFixed(Number.isInteger(seconds) ? 0 : 2)}s`;
}

export function formatRelativeTime(
  iso: string,
  now: number = Date.now(),
): string {
  const diffMs = now - Date.parse(iso);
  if (Number.isNaN(diffMs) || diffMs < 0) return "just now";
  const minutes = Math.floor(diffMs / MS_PER_MINUTE);
  if (minutes < 1) return "just now";
  if (minutes < MINUTES_PER_HOUR) return `${minutes}m ago`;
  const hours = Math.floor(minutes / MINUTES_PER_HOUR);
  if (hours < HOURS_PER_DAY) return `${hours}h ago`;
  return `${Math.floor(hours / HOURS_PER_DAY)}d ago`;
}
