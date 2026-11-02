import { computed, type Ref } from "vue";
import type { Attributes } from "@/shared/api/otel";
import type { Span } from "@/shared/types/entities/span";
import { serviceColor } from "@/shared/utils/service-color";
import { diffMsPrecise } from "@/shared/utils/time";

import type {
  SpanNode,
  TraceDetail,
  TraceDiagnosis,
  TraceIssue,
} from "../types";

export interface ServiceLegendEntry {
  name: string;
  color: string;
}

const ERRNO_PATTERN =
  /\b(ENOENT|EACCES|EPERM|ECONNREFUSED|ECONNRESET|ETIMEDOUT|EPIPE|ENOTFOUND|EAI_AGAIN|EHOSTUNREACH|ENETUNREACH|EADDRINUSE|EEXIST|ENOSPC|EROFS|EISDIR|ENOTDIR|ELOOP|EMSGSIZE|EPROTO|EBADF)\b/;

function textAttr(attrs: Attributes, key: string): string | undefined {
  const value = attrs[key];
  return typeof value === "string" && value !== "" ? value : undefined;
}

function exceptionMessage(span: Span): string | undefined {
  const event = span.events.find(
    (entry) => entry.name.toLowerCase() === "exception",
  );
  if (event === undefined) return undefined;
  return textAttr(event.attributes, "exception.message");
}

function errorTypeName(span: Span): string | undefined {
  const direct = textAttr(span.attributes, "error.type");
  if (direct !== undefined) return direct;
  const event = span.events.find(
    (entry) => entry.name.toLowerCase() === "exception",
  );
  if (event === undefined) return undefined;
  return textAttr(event.attributes, "exception.type");
}

function capitalize(value: string): string {
  if (value === "") return value;
  return value.charAt(0).toUpperCase() + value.slice(1);
}

function cleanMessage(message: string): string {
  const first = (message.split("\n")[0] ?? "").trim();
  const matches = first.match(/'[^']+'|"[^"]+"|\/[^\s'"]{4,}/g);
  let out = first;
  if (matches !== null) {
    const counts = new Map<string, number>();
    for (const token of matches) {
      counts.set(token, (counts.get(token) ?? 0) + 1);
    }
    for (const [token, count] of counts) {
      if (count < 2) continue;
      const at = out.indexOf(token);
      if (at === -1) continue;
      const head = out.slice(0, at + token.length);
      const tail = out
        .slice(at + token.length)
        .split(token)
        .join("");
      out = `${head}${tail}`;
    }
    out = out
      .replace(/\s{2,}/g, " ")
      .replace(/,?\s*open\s*$/i, "")
      .trim();
  }
  const errno = ERRNO_PATTERN.exec(out);
  if (errno !== null && errno.index !== undefined) {
    const tail = out
      .slice(errno.index)
      .replace(/,?\s*open\s*['"]{2}\s*$/i, "")
      .trim();
    if (tail !== "") return tail;
  }
  return out;
}

function httpCodeLabel(span: Span): string | undefined {
  for (const key of ["http.response.status_code", "http.status_code"]) {
    const value = span.attributes[key];
    const code =
      typeof value === "number"
        ? value
        : typeof value === "string" && value !== ""
          ? Number(value)
          : NaN;
    if (Number.isFinite(code)) return `HTTP ${code}`;
  }
  return undefined;
}

function errorHeadline(span: Span): string {
  const raw =
    exceptionMessage(span) ??
    (span.statusMessage === "" ? undefined : span.statusMessage);
  if (raw !== undefined) {
    const cleaned = cleanMessage(raw);
    if (cleaned !== "") return capitalize(cleaned);
  }
  const errorType = errorTypeName(span);
  if (errorType !== undefined && errorType.trim() !== "") {
    return errorType.trim();
  }
  return httpCodeLabel(span) ?? "No error message recorded";
}

function hasSignal(span: Span): boolean {
  return (
    exceptionMessage(span) !== undefined ||
    span.statusMessage !== "" ||
    textAttr(span.attributes, "error.type") !== undefined
  );
}

function findOrigin(nodes: SpanNode[]): Span | null {
  const flat: { span: Span; depth: number }[] = [];
  const walk = (list: SpanNode[], depth: number): void => {
    for (const node of list) {
      flat.push({ span: node.span, depth });
      walk(node.children, depth + 1);
    }
  };
  walk(nodes, 0);
  const errors = flat.filter((entry) => entry.span.status === "error");
  if (errors.length === 0) return null;
  const errorIds = new Set(errors.map((entry) => entry.span.spanId));
  const byId = new Map(flat.map((entry) => [entry.span.spanId, entry.span]));
  const hasErrorDescendant = new Set<string>();
  for (const entry of errors) {
    let parentId = entry.span.parentSpanId;
    while (parentId !== "") {
      if (errorIds.has(parentId)) hasErrorDescendant.add(parentId);
      const parent = byId.get(parentId);
      if (parent === undefined) break;
      parentId = parent.parentSpanId;
    }
  }
  const leaves = errors.filter(
    (entry) => !hasErrorDescendant.has(entry.span.spanId),
  );
  const informative = leaves.filter((entry) => hasSignal(entry.span));
  const pool = informative.length > 0 ? informative : leaves;
  let best = pool[0];
  if (best === undefined) return null;
  for (const entry of pool) {
    if (
      entry.depth > best.depth ||
      (entry.depth === best.depth && entry.span.startTime < best.span.startTime)
    ) {
      best = entry;
    }
  }
  return best.span;
}

export function useTraceGraph(detail: Ref<TraceDetail | undefined>): {
  flatSpans: Ref<Span[]>;
  rootNodes: Ref<SpanNode[]>;
  rootSpan: Ref<Span | null>;
  spanCount: Ref<number>;
  totalDurationMs: Ref<number>;
  serviceLegend: Ref<ServiceLegendEntry[]>;
  issues: Ref<TraceIssue[]>;
  diagnosis: Ref<TraceDiagnosis>;
  errorCount: Ref<number>;
  serviceCount: Ref<number>;
  spanStartMs: (span: Span) => number;
  spanDurationMs: (span: Span) => number;
  selfTimeMs: (spanId: string) => number;
  childCount: (spanId: string) => number;
  findSpan: (spanId: string) => Span | null;
  visibleSpans: (collapsed: Ref<Set<string>>) => Span[];
} {
  const nodes = computed((): SpanNode[] => detail.value?.spans ?? []);

  const flatSpans = computed((): Span[] => {
    const out: Span[] = [];
    const walk = (list: SpanNode[]): void => {
      for (const node of list) {
        out.push(node.span);
        walk(node.children);
      }
    };
    walk(nodes.value);
    return out;
  });

  const byId = computed((): Map<string, Span> => {
    const map = new Map<string, Span>();
    for (const span of flatSpans.value) map.set(span.spanId, span);
    return map;
  });

  const childDurations = computed((): Map<string, number> => {
    const map = new Map<string, number>();
    for (const span of flatSpans.value) {
      if (span.parentSpanId === "") continue;
      const duration = diffMsPrecise(span.startTime, span.endTime);
      map.set(
        span.parentSpanId,
        (map.get(span.parentSpanId) ?? 0) +
          (Number.isNaN(duration) ? 0 : Math.max(duration, 0)),
      );
    }
    return map;
  });

  const childCounts = computed((): Map<string, number> => {
    const map = new Map<string, number>();
    for (const span of flatSpans.value) {
      if (span.parentSpanId === "") continue;
      map.set(span.parentSpanId, (map.get(span.parentSpanId) ?? 0) + 1);
    }
    return map;
  });

  const rootNodes = nodes;

  const rootSpan = computed((): Span | null => {
    const firstRoot = nodes.value[0]?.span;
    if (firstRoot !== undefined) return firstRoot;
    return flatSpans.value[0] ?? null;
  });

  const spanCount = computed((): number => flatSpans.value.length);

  const totalDurationMs = computed((): number => {
    const trace = detail.value?.trace;
    if (trace === undefined) return 0;
    const value = diffMsPrecise(trace.startTime, trace.endTime);
    if (Number.isNaN(value)) return 0;
    return Math.max(value, 0);
  });

  const serviceLegend = computed((): ServiceLegendEntry[] => {
    const seen = new Set<string>();
    const legend: ServiceLegendEntry[] = [];
    for (const span of flatSpans.value) {
      if (seen.has(span.service)) continue;
      seen.add(span.service);
      legend.push({ name: span.service, color: serviceColor(span.service) });
    }
    return legend;
  });

  const errorSpans = computed((): Span[] =>
    flatSpans.value
      .filter((span) => span.status === "error")
      .sort((a, b) => {
        const durationA = diffMsPrecise(a.startTime, a.endTime);
        const durationB = diffMsPrecise(b.startTime, b.endTime);
        return (
          (Number.isNaN(durationB) ? 0 : durationB) -
          (Number.isNaN(durationA) ? 0 : durationA)
        );
      }),
  );

  const issues = computed((): TraceIssue[] =>
    errorSpans.value.map((span) => {
      const duration = diffMsPrecise(span.startTime, span.endTime);
      return {
        spanId: span.spanId,
        name: span.name,
        service: span.service,
        durationMs: Number.isNaN(duration) ? 0 : Math.max(duration, 0),
        status: span.status,
      };
    }),
  );

  const diagnosis = computed((): TraceDiagnosis => {
    const origin = findOrigin(nodes.value);
    if (origin === null) {
      return {
        kind: "ok",
        spanId: null,
        spanName: "",
        service: "",
        headline: "",
        tail: "",
      };
    }
    const traceOk = detail.value?.trace.status === "ok";
    return {
      kind: "error",
      spanId: origin.spanId,
      spanName: origin.name,
      service: origin.service,
      headline: errorHeadline(origin),
      tail: traceOk ? " The request still completed with an ok status." : "",
    };
  });

  const errorCount = computed((): number => errorSpans.value.length);

  const serviceCount = computed((): number => serviceLegend.value.length);

  function spanStartMs(span: Span): number {
    const traceStart = detail.value?.trace.startTime;
    if (traceStart === undefined) return 0;
    const value = diffMsPrecise(traceStart, span.startTime);
    if (Number.isNaN(value)) return 0;
    return Math.max(value, 0);
  }

  function spanDurationMs(span: Span): number {
    const value = diffMsPrecise(span.startTime, span.endTime);
    if (Number.isNaN(value)) return 0;
    return Math.max(value, 0);
  }

  function selfTimeMs(spanId: string): number {
    const span = byId.value.get(spanId);
    if (span === undefined) return 0;
    return Math.max(
      spanDurationMs(span) - (childDurations.value.get(spanId) ?? 0),
      0,
    );
  }

  function childCount(spanId: string): number {
    return childCounts.value.get(spanId) ?? 0;
  }

  function findSpan(spanId: string): Span | null {
    return byId.value.get(spanId) ?? null;
  }

  function visibleSpans(collapsed: Ref<Set<string>>): Span[] {
    const out: Span[] = [];
    const visit = (list: SpanNode[]): void => {
      for (const node of list) {
        out.push(node.span);
        if (!collapsed.value.has(node.span.spanId)) visit(node.children);
      }
    };
    visit(nodes.value);
    return out;
  }

  return {
    flatSpans,
    rootNodes,
    rootSpan,
    spanCount,
    totalDurationMs,
    serviceLegend,
    issues,
    diagnosis,
    errorCount,
    serviceCount,
    spanStartMs,
    spanDurationMs,
    selfTimeMs,
    childCount,
    findSpan,
    visibleSpans,
  };
}
