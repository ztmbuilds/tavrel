const countFormatter = new Intl.NumberFormat("en-US");

export function formatCount(value: number): string {
  return countFormatter.format(value);
}

const SHORT_TRACE_ID_LENGTH = 8;

export function truncateTraceId(traceId: string): string {
  if (traceId.length <= SHORT_TRACE_ID_LENGTH) return traceId;
  return traceId.slice(0, SHORT_TRACE_ID_LENGTH);
}
