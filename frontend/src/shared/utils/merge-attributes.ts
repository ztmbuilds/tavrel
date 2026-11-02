import type { Attributes, JsonValue } from "@/shared/api/otel";

function isAttributes(value: JsonValue | undefined): value is Attributes {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function mergeAttributes(
  base: Attributes,
  override: Attributes,
): Attributes {
  const merged: Attributes = { ...base };

  for (const [key, value] of Object.entries(override)) {
    const existing = merged[key];
    merged[key] =
      isAttributes(existing) && isAttributes(value)
        ? mergeAttributes(existing, value)
        : value;
  }

  return merged;
}
