const SERVICE_COLORS: readonly string[] = [
  "#7C9CD6",
  "#8FBF9F",
  "#D6A5C9",
  "#E0B454",
  "#7FC4C4",
  "#C99B7C",
  "#A5A8E0",
  "#B8C47C",
  "#E29A9A",
  "#9AC7E0",
];

const FALLBACK_SERVICE_COLOR = "#7C9CD6";
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
