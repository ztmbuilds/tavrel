import { shallowRef } from "vue";

const SERVICE_COLORS: readonly string[] = [
  "#F5B7A2",
  "#69E855",
  "#5DE4B1",
  "#F5AAEE",
  "#F18AA2",
  "#F09453",
  "#50C2C2",
  "#9CA8EF",
  "#F17BE0",
  "#A69B3E",
  "#45AE8B",
  "#AD7BEB",
  "#EE56AE",
  "#DA3E7E",
  "#E2413B",
  "#3A9A4C",
  "#3A9393",
  "#3788D0",
  "#B742E3",
  "#CD3EB3",
];

const FALLBACK_SERVICE_COLOR = SERVICE_COLORS[0] ?? "#F5B7A2";
const NAME_HASH_FACTOR = 31;
const serviceColorRegistry = shallowRef<ReadonlyMap<string, string>>(new Map());

function hashServiceName(name: string): number {
  let hash = 0;
  for (let i = 0; i < name.length; i += 1) {
    hash = (hash * NAME_HASH_FACTOR + name.charCodeAt(i)) >>> 0;
  }
  return hash;
}

function findLeastUsedSlot(
  preferredSlot: number,
  usageCounts: number[],
): number {
  let slot = preferredSlot;
  let lowestUsage = Number.POSITIVE_INFINITY;

  for (let offset = 0; offset < SERVICE_COLORS.length; offset += 1) {
    const candidate = (preferredSlot + offset) % SERVICE_COLORS.length;
    const usage = usageCounts[candidate] ?? 0;
    if (usage < lowestUsage) {
      slot = candidate;
      lowestUsage = usage;
    }
  }

  return slot;
}

export function serviceColor(name: string): string {
  const registeredColor = serviceColorRegistry.value.get(name);
  if (registeredColor !== undefined) return registeredColor;

  // Log-only services aren't in the span-backed registry, so this can collide.
  return (
    SERVICE_COLORS[hashServiceName(name) % SERVICE_COLORS.length] ??
    FALLBACK_SERVICE_COLOR
  );
}

function buildServiceColorMap(
  names: readonly string[],
): ReadonlyMap<string, string> {
  const sortedNames = [...new Set(names)].sort();
  const usageCounts = new Array<number>(SERVICE_COLORS.length).fill(0);
  const assignedSlots = new Array<boolean>(SERVICE_COLORS.length).fill(false);
  const colors = new Map<string, string>();

  for (const name of sortedNames) {
    const preferredSlot = hashServiceName(name) % SERVICE_COLORS.length;
    let slot = preferredSlot;

    if (assignedSlots.every(Boolean)) {
      slot = findLeastUsedSlot(preferredSlot, usageCounts);
    } else {
      while (assignedSlots[slot] === true) {
        slot = (slot + 1) % SERVICE_COLORS.length;
      }
    }

    const color = SERVICE_COLORS[slot] ?? FALLBACK_SERVICE_COLOR;
    colors.set(name, color);
    assignedSlots[slot] = true;
    usageCounts[slot] = (usageCounts[slot] ?? 0) + 1;
  }

  return colors;
}

export function setServiceColorRegistry(names: readonly string[]): void {
  serviceColorRegistry.value = buildServiceColorMap(names);
}
