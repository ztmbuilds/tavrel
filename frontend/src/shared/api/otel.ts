import type { Log, LogLevel } from "@/shared/types/entities/log";
import type { Span, SpanKind, SpanStatus } from "@/shared/types/entities/span";
import type {
  SpanLink,
  SpanLinkTarget,
} from "@/shared/types/entities/span-link";
import type { Trace, TraceStatus } from "@/shared/types/entities/trace";
import { diffMsPrecise } from "@/shared/utils/time";

export type JsonValue =
  string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue };

export type Attributes = Record<string, JsonValue>;

export interface SpanEventDto {
  time: string;
  name: string;
  attributes: Attributes;
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

const SPAN_KIND_BY_INDEX: readonly SpanKind[] = [
  "unspecified",
  "internal",
  "server",
  "client",
  "producer",
  "consumer",
];

export function toSpanKind(kind: number): SpanKind {
  return SPAN_KIND_BY_INDEX[kind] ?? "unspecified";
}

export function toSpanStatus(statusCode: number): SpanStatus {
  if (statusCode === 1) return "ok";
  if (statusCode === 2) return "error";
  return "unset";
}

const WARN_SEVERITY_MIN = 13;
const WARN_SEVERITY_MAX = 16;
const ERROR_SEVERITY_MIN = 17;

export function toLogLevel(severityNumber: number): LogLevel {
  if (severityNumber >= ERROR_SEVERITY_MIN) return "error";
  if (
    severityNumber >= WARN_SEVERITY_MIN &&
    severityNumber <= WARN_SEVERITY_MAX
  )
    return "warn";
  return "info";
}

export function toLogMessage(body: unknown): string {
  if (typeof body === "string") return body;
  if (isRecord(body)) {
    const stringValue = body["stringValue"];
    if (typeof stringValue === "string") return stringValue;
    const message = body["message"];
    if (typeof message === "string") return message;
  }
  try {
    return JSON.stringify(body) ?? "[unserializable]";
  } catch {
    return "[unserializable]";
  }
}

const HOST_KEYS: readonly string[] = ["host.name", "host", "hostname"];

export function lookupHost(value: unknown): string | undefined {
  if (!isRecord(value)) return undefined;
  for (const key of HOST_KEYS) {
    const candidate = value[key];
    if (typeof candidate === "string" && candidate !== "") return candidate;
  }
  return undefined;
}

export function toLogHost(
  attributes: unknown,
  resourceAttributes: unknown,
): string {
  return lookupHost(attributes) ?? lookupHost(resourceAttributes) ?? "";
}

export function formatAttributeValue(value: JsonValue): string {
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean")
    return String(value);
  if (value === null) return "null";
  try {
    return JSON.stringify(value) ?? "[unserializable]";
  } catch {
    return "[unserializable]";
  }
}

export interface SpanDto {
  spanId: string;
  traceId: string;
  parentSpanId: string;
  service: string;
  environment: string;
  name: string;
  kind: number;
  statusCode: number;
  statusMessage: string;
  startTime: string;
  endTime: string;
  attributes: unknown;
  events: unknown;
}

export interface LogDto {
  id: number;
  traceId: string;
  spanId: string;
  timestamp: string;
  severityText: string;
  severityNumber: number;
  body: unknown;
  service: string;
  environment: string;
  attributes: unknown;
  resourceAttributes: unknown;
}

export interface TraceDto {
  traceId: string;
  rootService: string;
  rootName: string;
  environment: string;
  startTime: string;
  endTime: string;
  spanCount: number;
  error: boolean;
}

export function toTrace(dto: TraceDto): Trace {
  const status: TraceStatus = dto.error ? "error" : "ok";
  return {
    traceId: dto.traceId,
    rootService: dto.rootService,
    rootName: dto.rootName,
    environment: dto.environment,
    startTime: dto.startTime,
    endTime: dto.endTime,
    spanCount: dto.spanCount,
    status,
  };
}

export interface StringListResponse {
  message: string;
  data: string[];
  meta: { nextCursor: string | null };
}

function toAttributes(value: unknown): Attributes {
  if (!isRecord(value)) return {};
  return value as Attributes;
}

function toRawSpanEvent(value: unknown): SpanEventDto | null {
  if (!isRecord(value)) return null;
  if (typeof value["time"] !== "string") return null;
  if (typeof value["name"] !== "string") return null;
  return {
    time: value["time"],
    name: value["name"],
    attributes: toAttributes(value["attributes"]),
  };
}

export function toSpanEvents(events: unknown): SpanEventDto[] {
  if (!Array.isArray(events)) return [];
  const out: SpanEventDto[] = [];
  for (const item of events) {
    const raw = toRawSpanEvent(item);
    if (raw !== null) out.push(raw);
  }
  return out;
}

export interface SpanLinkDto {
  traceId: string;
  spanId: string;
  attributes: unknown;
  target: unknown;
}

function toSpanLinkTarget(value: unknown): SpanLinkTarget | null {
  if (!isRecord(value)) return null;
  const name = value["name"];
  const service = value["service"];
  const statusCode = value["statusCode"];
  if (
    typeof name !== "string" ||
    typeof service !== "string" ||
    typeof statusCode !== "number"
  )
    return null;
  return { name, service, status: toSpanStatus(statusCode) };
}

export function toSpanLink(dto: SpanLinkDto): SpanLink {
  return {
    traceId: dto.traceId,
    spanId: dto.spanId,
    attributes: toAttributes(dto.attributes),
    target: toSpanLinkTarget(dto.target),
  };
}

export function toSpan(dto: SpanDto): Span {
  return {
    spanId: dto.spanId,
    traceId: dto.traceId,
    parentSpanId: dto.parentSpanId,
    service: dto.service,
    environment: dto.environment,
    name: dto.name,
    kind: toSpanKind(dto.kind),
    status: toSpanStatus(dto.statusCode),
    statusMessage: dto.statusMessage,
    startTime: dto.startTime,
    endTime: dto.endTime,
    attributes: toAttributes(dto.attributes),
    events: toSpanEvents(dto.events),
  };
}

export function toLog(dto: LogDto): Log {
  return {
    id: dto.id,
    traceId: dto.traceId,
    spanId: dto.spanId,
    timestamp: dto.timestamp,
    service: dto.service,
    environment: dto.environment,
    severityText: dto.severityText,
    severityNumber: dto.severityNumber,
    level: toLogLevel(dto.severityNumber),
    message: toLogMessage(dto.body),
    host: toLogHost(dto.attributes, dto.resourceAttributes),
    attributes: toAttributes(dto.attributes),
    resourceAttributes: toAttributes(dto.resourceAttributes),
  };
}

export function toEventDetail(attributes: Attributes): string | undefined {
  const message = attributes["exception.message"];
  const type = attributes["exception.type"];
  const parts: string[] = [];
  if (typeof type === "string" && type !== "") parts.push(type);
  if (typeof message === "string" && message !== "") parts.push(message);
  if (parts.length > 0) return parts.join(": ");
  const keys = Object.keys(attributes);
  if (keys.length === 0) return undefined;
  try {
    return JSON.stringify(attributes);
  } catch {
    return undefined;
  }
}

export interface SpanEventViewInput {
  offsetMs: number;
  name: string;
  type: "error" | "info";
  detail?: string;
}

export function toSpanEventViews(
  events: SpanEventDto[],
  spanStartTime: string,
): SpanEventViewInput[] {
  const views: SpanEventViewInput[] = [];
  for (const raw of events) {
    const offsetMs = Math.max(diffMsPrecise(spanStartTime, raw.time), 0);
    const view: SpanEventViewInput = {
      offsetMs: Number.isNaN(offsetMs) ? 0 : offsetMs,
      name: raw.name,
      type: raw.name.toLowerCase() === "exception" ? "error" : "info",
    };
    const detail = toEventDetail(raw.attributes);
    if (detail !== undefined) view.detail = detail;
    views.push(view);
  }
  return views.sort((a, b) => a.offsetMs - b.offsetMs);
}
