const countFormatter = new Intl.NumberFormat("en-US");

export function formatCount(value: number): string {
  return countFormatter.format(value);
}

export const SHORT_LIST_ID_LENGTH = 7;
const SHORT_TRACE_ID_LENGTH = 8;

export function truncateTraceId(
  traceId: string,
  length = SHORT_TRACE_ID_LENGTH,
): string {
  if (traceId.length <= length) return traceId;
  return traceId.slice(0, length);
}
