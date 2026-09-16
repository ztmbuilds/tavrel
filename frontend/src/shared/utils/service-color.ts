const SERVICE_COLORS: readonly string[] = [
  "#5E8FD7",
  "#4CA381",
  "#BE6DB0",
  "#C0911A",
  "#3FA3A3",
  "#B07A45",
  "#7F83D6",
  "#8FA33A",
  "#CC6B78",
  "#4E9BCD",
];

const FALLBACK_SERVICE_COLOR = "#5E8FD7";
const NAME_HASH_FACTOR = 31;

function hashServiceName(name: string): number {
  let hash = 0;
  for (let i = 0; i < name.length; i += 1) {
    hash = (hash * NAME_HASH_FACTOR + name.charCodeAt(i)) >>> 0;
  }
  return hash;
}

export function serviceColor(name: string): string {
  if (SERVICE_COLORS.length === 0) return FALLBACK_SERVICE_COLOR;
  return (
    SERVICE_COLORS[hashServiceName(name) % SERVICE_COLORS.length] ??
    FALLBACK_SERVICE_COLOR
  );
}
