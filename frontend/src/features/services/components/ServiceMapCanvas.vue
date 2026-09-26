<script setup lang="ts">
import { computed, ref } from "vue";

import { cn } from "@/shared/utils/cn";
import { serviceColor } from "@/shared/utils/service-color";

import { MAP_NODE_H, MAP_NODE_W, type MapLayout } from "../layout";

const props = defineProps<{
  layout: MapLayout;
  selected: string | null;
  query: string;
}>();

const emit = defineEmits<{
  select: [id: string];
  reset: [];
}>();

const ZOOM_MIN = 0.6;
const ZOOM_MAX = 1.6;
const ZOOM_STEP = 0.15;
const DRAG_THRESHOLD_PX = 3;

const zoom = ref(1);
const hovered = ref<string | null>(null);
const dragging = ref(false);
const scrollerRef = ref<HTMLDivElement | null>(null);

let dragMoved = false;
let captureActive = false;
let dragStartX = 0;
let dragStartY = 0;
let dragStartLeft = 0;
let dragStartTop = 0;

const neighborSet = computed((): Set<string> | null => {
  const focus = hovered.value ?? props.selected;
  if (focus === null) return null;
  const set = new Set<string>([focus]);
  for (const edge of props.layout.edgePaths) {
    if (edge.from === focus) set.add(edge.to);
    if (edge.to === focus) set.add(edge.from);
  }
  return set;
});

const trimmedQuery = computed(() => props.query.trim().toLowerCase());

function matchesQuery(id: string): boolean {
  if (trimmedQuery.value === "") return true;
  return id.toLowerCase().includes(trimmedQuery.value);
}

function isNodeDim(id: string): boolean {
  if (trimmedQuery.value !== "") return !matchesQuery(id);
  const set = neighborSet.value;
  if (set === null) return false;
  return !set.has(id);
}

function isEdgeDim(from: string, to: string): boolean {
  if (trimmedQuery.value !== "")
    return !matchesQuery(from) && !matchesQuery(to);
  const set = neighborSet.value;
  if (set === null) return false;
  return !(set.has(from) && set.has(to));
}

function isEdgeHot(from: string, to: string): boolean {
  if (trimmedQuery.value !== "") return false;
  const set = neighborSet.value;
  if (set === null) return false;
  return set.has(from) && set.has(to);
}

function edgeStroke(from: string, to: string): string {
  return isEdgeHot(from, to) ? "var(--color-secondary)" : "var(--color-strong)";
}

function zoomIn(): void {
  zoom.value = Math.min(ZOOM_MAX, +(zoom.value + ZOOM_STEP).toFixed(2));
}

function zoomOut(): void {
  zoom.value = Math.max(ZOOM_MIN, +(zoom.value - ZOOM_STEP).toFixed(2));
}

function resetView(): void {
  zoom.value = 1;
  emit("reset");
}

function handleNodeKeydown(event: KeyboardEvent, id: string): void {
  if (event.key === "Enter" || event.key === " ") {
    event.preventDefault();
    emit("select", id);
  }
}

function startPan(event: PointerEvent): void {
  const scroller = scrollerRef.value;
  if (scroller === null) return;
  dragMoved = false;
  captureActive = false;
  dragStartX = event.clientX;
  dragStartY = event.clientY;
  dragStartLeft = scroller.scrollLeft;
  dragStartTop = scroller.scrollTop;
}

function movePan(event: PointerEvent): void {
  const scroller = scrollerRef.value;
  if (scroller === null) return;
  const dx = event.clientX - dragStartX;
  const dy = event.clientY - dragStartY;
  if (Math.abs(dx) <= DRAG_THRESHOLD_PX && Math.abs(dy) <= DRAG_THRESHOLD_PX)
    return;
  if (!captureActive) {
    try {
      scroller.setPointerCapture(event.pointerId);
    } catch {
      return;
    }
    captureActive = true;
  }
  dragMoved = true;
  dragging.value = true;
  scroller.scrollLeft = dragStartLeft - dx;
  scroller.scrollTop = dragStartTop - dy;
}

function endPan(event: PointerEvent): void {
  const scroller = scrollerRef.value;
  if (scroller === null) return;
  dragging.value = false;
  if (!captureActive) return;
  captureActive = false;
  try {
    scroller.releasePointerCapture(event.pointerId);
  } catch {
    return;
  }
}

function suppressClickAfterDrag(event: MouseEvent): void {
  if (dragMoved) {
    event.stopPropagation();
    event.preventDefault();
    dragMoved = false;
  }
}
</script>

<template>
  <div class="bg-panel relative h-full min-h-[32.5rem]">
    <div
      class="border-default bg-surface absolute top-3 right-3 z-10 flex flex-col gap-px overflow-hidden rounded-lg border"
      role="group"
      aria-label="Zoom"
    >
      <button
        type="button"
        aria-label="Zoom in"
        class="text-secondary hover:text-primary hover:bg-hover flex h-[1.625rem] w-[1.625rem] cursor-pointer items-center justify-center border-none bg-transparent p-0 text-sm transition-colors duration-120 ease-out"
        @click="zoomIn"
      >
        +
      </button>
      <button
        type="button"
        aria-label="Zoom out"
        class="text-secondary hover:text-primary hover:bg-hover flex h-[1.625rem] w-[1.625rem] cursor-pointer items-center justify-center border-none bg-transparent p-0 text-sm transition-colors duration-120 ease-out"
        @click="zoomOut"
      >
        −
      </button>
      <button
        type="button"
        aria-label="Reset view"
        class="text-secondary hover:text-primary hover:bg-hover flex h-[1.625rem] w-[1.625rem] cursor-pointer items-center justify-center border-none bg-transparent p-0 text-[0.625rem] transition-colors duration-120 ease-out"
        @click="resetView"
      >
        ⟲
      </button>
    </div>
    <div
      ref="scrollerRef"
      :class="
        cn('h-full cursor-grab overflow-auto', dragging && 'cursor-grabbing')
      "
      @pointerdown="startPan"
      @pointermove="movePan"
      @pointerup="endPan"
      @pointercancel="endPan"
      @pointerleave="endPan"
      @click.capture="suppressClickAfterDrag"
    >
      <svg
        class="block"
        role="group"
        aria-label="Service dependency graph. Tab to a service, Enter to inspect it."
        :viewBox="`0 0 ${layout.width} ${layout.height}`"
        :style="{
          width: `${layout.width * zoom}px`,
          height: `${layout.height * zoom}px`,
        }"
      >
        <defs>
          <marker
            id="svc-arr"
            viewBox="0 0 8 8"
            refX="7"
            refY="4"
            markerWidth="6"
            markerHeight="6"
            orient="auto-start-reverse"
          >
            <path d="M0 0L8 4L0 8z" fill="var(--color-strong)" />
          </marker>
        </defs>
        <g v-for="edge in layout.edgePaths" :key="`${edge.from}→${edge.to}`">
          <path
            class="cursor-pointer"
            :d="edge.d"
            fill="none"
            stroke="transparent"
            stroke-width="16"
            @click="emit('select', edge.to)"
          />
          <path
            class="fill-none transition-opacity duration-150 ease-out"
            :d="edge.d"
            :stroke="edgeStroke(edge.from, edge.to)"
            :stroke-width="isEdgeHot(edge.from, edge.to) ? 1.6 : 1.4"
            :opacity="isEdgeDim(edge.from, edge.to) ? 0.15 : 1"
            marker-end="url(#svc-arr)"
            aria-hidden="true"
          />
        </g>
        <g
          v-for="(column, columnIndex) in layout.columns"
          :key="`col-${columnIndex}`"
        >
          <g
            v-for="id in column"
            :key="id"
            :class="
              cn(
                'cursor-pointer transition-opacity duration-150 ease-out',
                isNodeDim(id) && 'opacity-30',
              )
            "
            :transform="`translate(${layout.positions[id]?.x ?? 0},${layout.positions[id]?.y ?? 0})`"
            tabindex="0"
            role="button"
            :aria-label="`${id}: activate to inspect`"
            @click="emit('select', id)"
            @keydown="handleNodeKeydown($event, id)"
            @mouseenter="hovered = id"
            @mouseleave="hovered = null"
            @focus="hovered = id"
            @blur="hovered = null"
          >
            <rect
              :width="MAP_NODE_W"
              :height="MAP_NODE_H"
              rx="7"
              fill="var(--color-raised)"
              stroke="var(--color-default)"
              stroke-width="1.3"
            />
            <rect
              x="0"
              y="8"
              width="3"
              :height="MAP_NODE_H - 16"
              rx="1.5"
              :fill="serviceColor(id)"
            />
            <text
              class="font-mono"
              x="14"
              :y="MAP_NODE_H / 2 + 4"
              fill="var(--color-heading)"
              font-size="12"
              font-weight="500"
            >
              {{ id }}
            </text>
          </g>
        </g>
      </svg>
    </div>
  </div>
</template>
