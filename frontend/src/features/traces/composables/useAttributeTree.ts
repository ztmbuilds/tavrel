import {
  computed,
  ref,
  watch,
  type CSSProperties,
  type ComputedRef,
  type Ref,
} from "vue";

const AttributeViewMode = {
  Nested: "nested",
  Flat: "flat",
} as const;
export type AttributeViewMode =
  (typeof AttributeViewMode)[keyof typeof AttributeViewMode];

export type AttributeVisibleNode =
  | {
      kind: "group";
      name: string;
      fullPath: string;
      value: string | null;
      childCount: number;
      expanded: boolean;
      children: AttributeVisibleNode[];
    }
  | {
      kind: "leaf";
      name: string;
      fullPath: string;
      value: string;
    };

export type AttributeFlatRow = {
  path: string;
  namespace: string;
  leaf: string;
  value: string;
};

interface AttributeTreeNode {
  name: string;
  fullPath: string;
  value: string | null;
  leafCount: number;
  children: AttributeTreeNode[];
}

const VIEW_MODE_STORAGE_KEY = "span-attributes-view-mode";
const VIRTUALIZE_THRESHOLD = 200;
const COPY_FLASH_MS = 1600;

function readInitialMode(): AttributeViewMode {
  try {
    return localStorage.getItem(VIEW_MODE_STORAGE_KEY) ===
      AttributeViewMode.Flat
      ? AttributeViewMode.Flat
      : AttributeViewMode.Nested;
  } catch {
    return AttributeViewMode.Nested;
  }
}

export function useAttributeTree(source: Ref<unknown>): {
  mode: Ref<AttributeViewMode>;
  copiedPath: Ref<string | null>;
  nestedNodes: ComputedRef<AttributeVisibleNode[]>;
  flatRows: ComputedRef<AttributeFlatRow[]>;
  hasGroups: ComputedRef<boolean>;
  isEmpty: ComputedRef<boolean>;
  isLarge: ComputedRef<boolean>;
  virtualStyle: ComputedRef<CSSProperties | undefined>;
  setMode: (mode: AttributeViewMode) => void;
  toggle: (path: string) => void;
  expandAll: () => void;
  collapseAll: () => void;
  copyAttribute: (path: string, text: string) => void;
} {
  const mode = ref<AttributeViewMode>(readInitialMode());
  const copiedPath = ref<string | null>(null);
  const collapsed = ref<Set<string>>(new Set());
  let copyTimer: ReturnType<typeof setTimeout> | undefined;

  const roots = computed((): AttributeTreeNode[] => {
    if (!isRecord(source.value)) return [];
    const nodes: AttributeTreeNode[] = [];
    for (const [key, value] of Object.entries(source.value)) {
      insertDotted(nodes, "", key, value);
    }
    sortTree(nodes);
    for (const node of nodes) {
      tallyLeaves(node);
    }
    return nodes;
  });

  const totalLeaves = computed((): number => {
    let total = 0;
    for (const node of roots.value) {
      total += node.leafCount;
    }
    return total;
  });

  const isEmpty = computed((): boolean => totalLeaves.value === 0);

  const isLarge = computed(
    (): boolean => totalLeaves.value > VIRTUALIZE_THRESHOLD,
  );

  const virtualStyle = computed((): CSSProperties | undefined =>
    isLarge.value
      ? { contentVisibility: "auto", containIntrinsicSize: "auto 2.25rem" }
      : undefined,
  );

  const nestedNodes = computed((): AttributeVisibleNode[] => {
    const out: AttributeVisibleNode[] = [];
    for (const node of roots.value) {
      out.push(toVisibleNode(node, collapsed.value));
    }
    return out;
  });

  const flatRows = computed((): AttributeFlatRow[] => {
    const leaves: { path: string; value: string }[] = [];
    collectLeaves(roots.value, leaves);
    const rows: AttributeFlatRow[] = [];
    for (const leaf of leaves) {
      const separator = leaf.path.lastIndexOf(".");
      rows.push({
        path: leaf.path,
        namespace: separator === -1 ? "" : leaf.path.slice(0, separator + 1),
        leaf: separator === -1 ? leaf.path : leaf.path.slice(separator + 1),
        value: leaf.value,
      });
    }
    return rows;
  });

  const hasGroups = computed((): boolean =>
    roots.value.some((node) => node.children.length > 0),
  );

  watch(source, () => {
    collapsed.value = new Set();
  });

  function setMode(next: AttributeViewMode): void {
    mode.value = next;
    try {
      localStorage.setItem(VIEW_MODE_STORAGE_KEY, next);
    } catch {
      return;
    }
  }

  function toggle(path: string): void {
    const next = new Set(collapsed.value);
    if (next.has(path)) {
      next.delete(path);
    } else {
      next.add(path);
    }
    collapsed.value = next;
  }

  function expandAll(): void {
    collapsed.value = new Set();
  }

  function collapseAll(): void {
    const paths = new Set<string>();
    collectGroupPaths(roots.value, paths);
    collapsed.value = paths;
  }

  function copyAttribute(path: string, text: string): void {
    void navigator.clipboard?.writeText(text).catch(() => {});
    copiedPath.value = path;
    if (copyTimer !== undefined) clearTimeout(copyTimer);
    copyTimer = setTimeout(() => {
      copiedPath.value = null;
    }, COPY_FLASH_MS);
  }

  return {
    mode,
    copiedPath,
    nestedNodes,
    flatRows,
    hasGroups,
    isEmpty,
    isLarge,
    virtualStyle,
    setMode,
    toggle,
    expandAll,
    collapseAll,
    copyAttribute,
  };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function formatAttributeValue(value: unknown): string {
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean")
    return String(value);
  if (value === null || value === undefined) return "null";
  try {
    return JSON.stringify(value);
  } catch {
    return "[unserializable]";
  }
}

function getOrCreate(
  siblings: AttributeTreeNode[],
  name: string,
  fullPath: string,
): AttributeTreeNode {
  const existing = siblings.find((node) => node.name === name);
  if (existing !== undefined) return existing;
  const node: AttributeTreeNode = {
    name,
    fullPath,
    value: null,
    leafCount: 0,
    children: [],
  };
  siblings.push(node);
  return node;
}

function insertDotted(
  siblings: AttributeTreeNode[],
  basePath: string,
  dottedKey: string,
  rawValue: unknown,
): void {
  const parts = dottedKey.split(".").filter((part) => part.length > 0);
  const segments = parts.length > 0 ? parts : [dottedKey];
  let currentSiblings = siblings;
  let currentBase = basePath;
  for (let index = 0; index < segments.length; index += 1) {
    const segment = segments[index];
    if (segment === undefined) continue;
    const fullPath = currentBase === "" ? segment : `${currentBase}.${segment}`;
    const node = getOrCreate(currentSiblings, segment, fullPath);
    const isLast = index === segments.length - 1;
    if (!isLast) {
      currentSiblings = node.children;
      currentBase = fullPath;
      continue;
    }
    if (!isRecord(rawValue)) {
      node.value = formatAttributeValue(rawValue);
      continue;
    }
    const entries = Object.entries(rawValue);
    if (entries.length === 0) {
      node.value = "{}";
      continue;
    }
    for (const [key, value] of entries) {
      insertDotted(node.children, node.fullPath, key, value);
    }
  }
}

function sortTree(nodes: AttributeTreeNode[]): void {
  nodes.sort((a, b) => a.name.localeCompare(b.name));
  for (const node of nodes) {
    sortTree(node.children);
  }
}

function tallyLeaves(node: AttributeTreeNode): number {
  let total = node.children.length === 0 || node.value !== null ? 1 : 0;
  for (const child of node.children) {
    total += tallyLeaves(child);
  }
  node.leafCount = total;
  return total;
}

function toVisibleNode(
  node: AttributeTreeNode,
  collapsedPaths: Set<string>,
): AttributeVisibleNode {
  if (node.children.length === 0) {
    return {
      kind: "leaf",
      name: node.name,
      fullPath: node.fullPath,
      value: node.value ?? "null",
    };
  }
  const children: AttributeVisibleNode[] = [];
  for (const child of node.children) {
    children.push(toVisibleNode(child, collapsedPaths));
  }
  return {
    kind: "group",
    name: node.name,
    fullPath: node.fullPath,
    value: node.value,
    childCount: node.leafCount,
    expanded: !collapsedPaths.has(node.fullPath),
    children,
  };
}

function collectLeaves(
  nodes: AttributeTreeNode[],
  out: { path: string; value: string }[],
): void {
  for (const node of nodes) {
    if (node.children.length === 0) {
      out.push({ path: node.fullPath, value: node.value ?? "null" });
      continue;
    }
    if (node.value !== null) {
      out.push({ path: node.fullPath, value: node.value });
    }
    collectLeaves(node.children, out);
  }
}

function collectGroupPaths(nodes: AttributeTreeNode[], out: Set<string>): void {
  for (const node of nodes) {
    if (node.children.length === 0) continue;
    out.add(node.fullPath);
    collectGroupPaths(node.children, out);
  }
}
