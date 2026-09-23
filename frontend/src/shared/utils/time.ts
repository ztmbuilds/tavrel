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
const HOURS_PER_HALF_DAY = 12;

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

export function formatShortTime(iso: string): string {
  const parsed = new Date(iso);
  const hour24 = parsed.getHours();
  const suffix = hour24 < HOURS_PER_HALF_DAY ? "AM" : "PM";
  const hour12 =
    hour24 % HOURS_PER_HALF_DAY === 0
      ? HOURS_PER_HALF_DAY
      : hour24 % HOURS_PER_HALF_DAY;
  return `${hour12}:${pad2(parsed.getMinutes())} ${suffix}`;
}

export function formatShortDateTime(iso: string): string {
  const parsed = new Date(iso);
  const month = parsed.toLocaleDateString("en-US", { month: "short" });
  const yearSuffix =
    parsed.getFullYear() === new Date().getFullYear()
      ? ""
      : `, ${parsed.getFullYear()}`;
  return `${month} ${parsed.getDate()}${yearSuffix}, ${formatShortTime(iso)}`;
}

export function formatShortDateTimeWithSeconds(iso: string): string {
  const parsed = new Date(iso);
  const month = parsed.toLocaleDateString("en-US", { month: "short" });
  const yearSuffix =
    parsed.getFullYear() === new Date().getFullYear()
      ? ""
      : `, ${parsed.getFullYear()}`;
  const hour24 = parsed.getHours();
  const suffix = hour24 < HOURS_PER_HALF_DAY ? "AM" : "PM";
  const hour12 =
    hour24 % HOURS_PER_HALF_DAY === 0
      ? HOURS_PER_HALF_DAY
      : hour24 % HOURS_PER_HALF_DAY;
  const clock = `${hour12}:${pad2(parsed.getMinutes())}:${pad2(parsed.getSeconds())} ${suffix}`;
  return `${month} ${parsed.getDate()}${yearSuffix}, ${clock}`;
}

export function formatDateTimeRange(
  start: string | undefined,
  end: string | undefined,
): string {
  if (start !== undefined && end !== undefined) {
    const startDate = new Date(start);
    const endDate = new Date(end);
    const sameDay =
      startDate.getFullYear() === endDate.getFullYear() &&
      startDate.getMonth() === endDate.getMonth() &&
      startDate.getDate() === endDate.getDate();
    if (sameDay) {
      const month = startDate.toLocaleDateString("en-US", { month: "short" });
      const yearSuffix =
        startDate.getFullYear() === new Date().getFullYear()
          ? ""
          : `, ${startDate.getFullYear()}`;
      return `${month} ${startDate.getDate()}${yearSuffix}, ${formatShortTime(start)} – ${formatShortTime(end)}`;
    }
    return `${formatShortDateTime(start)} – ${formatShortDateTime(end)}`;
  }
  if (start !== undefined && end === undefined)
    return `Since ${formatShortDateTime(start)}`;
  if (start === undefined && end !== undefined)
    return `Before ${formatShortDateTime(end)}`;
  return "…";
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

const US_PER_MS = 1000;
const NS_PER_MS = 1000000;
const NS_PER_SEC = 1000000000;
const FRACTION_DIGITS = 9;

function parseFractionNs(iso: string): number | null {
  const match = /\.(\d+)(?:Z|[+-]\d{2}:?\d{2})?$/.exec(iso);
  if (match === null) return 0;
  const digits = match[1] ?? "";
  if (!/^\d+$/.test(digits)) return null;
  return Number(digits.slice(0, FRACTION_DIGITS).padEnd(FRACTION_DIGITS, "0"));
}

export function diffMsPrecise(startIso: string, endIso: string): number {
  const startMs = Date.parse(startIso);
  const endMs = Date.parse(endIso);
  if (Number.isNaN(startMs) || Number.isNaN(endMs)) return NaN;
  const startNs = parseFractionNs(startIso);
  const endNs = parseFractionNs(endIso);
  if (startNs === null || endNs === null) return endMs - startMs;
  const startSec = Math.floor(startMs / MS_PER_SECOND);
  const endSec = Math.floor(endMs / MS_PER_SECOND);
  return ((endSec - startSec) * NS_PER_SEC + (endNs - startNs)) / NS_PER_MS;
}

export function formatDurationMs(ms: number): string {
  if (Number.isNaN(ms) || ms <= 0) return `0µs`;
  if (ms < 1) {
    if (ms < 1 / US_PER_MS)
      return `${Math.max(1, Math.round(ms * NS_PER_MS))}ns`;
    return `${Math.round(ms * US_PER_MS)}µs`;
  }
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
