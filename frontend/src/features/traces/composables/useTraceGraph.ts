import { computed, type Ref } from "vue";
import { toSpanEventViews } from "@/shared/api/otel";
import type { Span } from "@/shared/types/entities/span";
import { serviceColor } from "@/shared/utils/service-color";
import { formatDurationMs } from "@/shared/utils/time";

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

export interface AncestorEntry {
  spanId: string;
  name: string;
}

const FULL_PCT = 100;

function errorDetail(span: Span): string | undefined {
  const views = toSpanEventViews(span.events, span.startTime);
  const failure = views.find((view) => view.type === "error" && view.detail);
  if (failure?.detail === undefined) return undefined;
  const firstLine = failure.detail.split("\n")[0] ?? "";
  const separator = firstLine.indexOf(": ");
  const message =
    separator === -1 ? firstLine : firstLine.slice(separator + 2);
  if (message === "") return undefined;
  return message.charAt(0).toLowerCase() + message.slice(1);
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
  ancestors: (spanId: string) => AncestorEntry[];
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
      map.set(
        span.parentSpanId,
        (map.get(span.parentSpanId) ?? 0) +
          Math.max(
            Date.parse(span.endTime) - Date.parse(span.startTime),
            0,
          ),
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
    return Math.max(Date.parse(trace.endTime) - Date.parse(trace.startTime), 0);
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
      .sort(
        (a, b) =>
          Date.parse(b.endTime) -
          Date.parse(b.startTime) -
          (Date.parse(a.endTime) - Date.parse(a.startTime)),
      ),
  );

  const issues = computed((): TraceIssue[] =>
    errorSpans.value.map((span) => ({
      spanId: span.spanId,
      name: span.name,
      service: span.service,
      durationMs: Math.max(
        Date.parse(span.endTime) - Date.parse(span.startTime),
        0,
      ),
      status: span.status,
    })),
  );

  const diagnosis = computed((): TraceDiagnosis => {
    const errors = errorSpans.value;
    const top = errors[0];
    if (top === undefined) {
      return { kind: "ok", spanId: null, spanName: "", suffix: "", tail: "" };
    }
    const durationMs = Math.max(
      Date.parse(top.endTime) - Date.parse(top.startTime),
      0,
    );
    const total = totalDurationMs.value;
    const percent = total > 0 ? Math.round((durationMs / total) * FULL_PCT) : 0;
    const detailText = errorDetail(top);
    const traceOk = detail.value?.trace.status === "ok";
    return {
      kind: "error",
      spanId: top.spanId,
      spanName: top.name,
      suffix: ` failed in ${top.service}${detailText !== undefined && detailText !== "" ? ` — ${detailText}` : ""}, using ${formatDurationMs(durationMs)} (${percent}% of the trace).`,
      tail: traceOk ? " The request still completed with an ok status." : "",
    };
  });

  const errorCount = computed((): number => errorSpans.value.length);

  const serviceCount = computed((): number => serviceLegend.value.length);

  function spanStartMs(span: Span): number {
    const traceStart = detail.value?.trace.startTime;
    if (traceStart === undefined) return 0;
    return Math.max(Date.parse(span.startTime) - Date.parse(traceStart), 0);
  }

  function spanDurationMs(span: Span): number {
    return Math.max(
      Date.parse(span.endTime) - Date.parse(span.startTime),
      0,
    );
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

  function ancestors(spanId: string): AncestorEntry[] {
    const chain: AncestorEntry[] = [];
    let current = byId.value.get(spanId);
    while (current !== undefined) {
      chain.unshift({ spanId: current.spanId, name: current.name });
      if (current.parentSpanId === "") break;
      current = byId.value.get(current.parentSpanId);
    }
    return chain;
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
    ancestors,
    findSpan,
    visibleSpans,
  };
}
