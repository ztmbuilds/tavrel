<script setup lang="ts">
import { computed } from "vue";
import { BaseEdge, getBezierPath, type EdgeProps } from "@vue-flow/core";

import type { ServiceMapEdgeData } from "../types";

type ServiceMapEdgeProps = Pick<
  EdgeProps<ServiceMapEdgeData>,
  | "id"
  | "sourceX"
  | "sourceY"
  | "sourcePosition"
  | "targetX"
  | "targetY"
  | "targetPosition"
  | "markerEnd"
  | "style"
  | "data"
>;

const props = defineProps<ServiceMapEdgeProps>();

const ROUTE_CORNER_RADIUS = 18;
const MIN_SEGMENT_LENGTH = 0.1;
const STRAIGHT_ANGLE_EPSILON = 0.01;

const path = computed(() => {
  if (props.data.routePoints.length === 0) {
    return getBezierPath({
      sourceX: props.sourceX,
      sourceY: props.sourceY,
      sourcePosition: props.sourcePosition,
      targetX: props.targetX,
      targetY: props.targetY,
      targetPosition: props.targetPosition,
    })[0];
  }

  return roundedPath([
    { x: props.sourceX, y: props.sourceY },
    ...props.data.routePoints,
    { x: props.targetX, y: props.targetY },
  ]);
});

function roundedPath(points: Array<{ x: number; y: number }>): string {
  const route = points.filter(
    (point, index) =>
      index === 0 || distance(point, points[index - 1]) > MIN_SEGMENT_LENGTH,
  );
  const first = route[0];
  if (first === undefined) return "";
  if (route.length === 1) return `M${first.x},${first.y}`;

  let result = `M${first.x},${first.y}`;
  for (let index = 1; index < route.length - 1; index += 1) {
    const previous = route[index - 1];
    const corner = route[index];
    const next = route[index + 1];
    if (previous === undefined || corner === undefined || next === undefined)
      continue;

    const incomingLength = distance(previous, corner);
    const outgoingLength = distance(corner, next);
    const turn =
      (corner.x - previous.x) * (next.y - corner.y) -
      (corner.y - previous.y) * (next.x - corner.x);
    if (
      Math.abs(turn) < STRAIGHT_ANGLE_EPSILON ||
      incomingLength < MIN_SEGMENT_LENGTH ||
      outgoingLength < MIN_SEGMENT_LENGTH
    ) {
      result += ` L${corner.x},${corner.y}`;
      continue;
    }

    const radius = Math.min(
      ROUTE_CORNER_RADIUS,
      incomingLength / 2,
      outgoingLength / 2,
    );
    const before = toward(corner, previous, radius / incomingLength);
    const after = toward(corner, next, radius / outgoingLength);
    result += ` L${before.x},${before.y} Q${corner.x},${corner.y} ${after.x},${after.y}`;
  }

  const last = route[route.length - 1];
  if (last !== undefined) result += ` L${last.x},${last.y}`;
  return result;
}

function distance(
  a: { x: number; y: number },
  b: { x: number; y: number } | undefined,
): number {
  if (b === undefined) return 0;
  return Math.hypot(a.x - b.x, a.y - b.y);
}

function toward(
  origin: { x: number; y: number },
  target: { x: number; y: number },
  amount: number,
): { x: number; y: number } {
  return {
    x: origin.x + (target.x - origin.x) * amount,
    y: origin.y + (target.y - origin.y) * amount,
  };
}
</script>

<template>
  <BaseEdge
    :id="props.id"
    :path="path"
    :marker-end="props.markerEnd"
    :style="props.style"
  />
</template>
