<script setup lang="ts">
import { Handle, Position, type NodeProps } from "@vue-flow/core";

import { cn } from "@/shared/utils/cn";
import { serviceColor } from "@/shared/utils/service-color";

import type { ServiceMapNodeData } from "../types";

type ServiceMapNodeProps = Pick<NodeProps<ServiceMapNodeData>, "id" | "data">;

const props = defineProps<ServiceMapNodeProps>();

const KEYBOARD_MOVE_STEP = 24;
const KEYBOARD_MOVE_MULTIPLIER = 4;

const sides = [
  { id: "left", position: Position.Left },
  { id: "right", position: Position.Right },
  { id: "top", position: Position.Top },
  { id: "bottom", position: Position.Bottom },
];

function handleKeydown(event: KeyboardEvent): void {
  if (!event.altKey) return;
  const delta = {
    ArrowUp: { x: 0, y: -1 },
    ArrowDown: { x: 0, y: 1 },
    ArrowLeft: { x: -1, y: 0 },
    ArrowRight: { x: 1, y: 0 },
  }[event.key];
  if (delta === undefined) return;

  event.preventDefault();
  const step = event.shiftKey
    ? KEYBOARD_MOVE_STEP * KEYBOARD_MOVE_MULTIPLIER
    : KEYBOARD_MOVE_STEP;
  props.data.onMove({ x: delta.x * step, y: delta.y * step });
}

function dotStyle(name: string): Record<string, string> {
  const color = serviceColor(name);
  return {
    background: color,
    boxShadow: `0 0 0 3px color-mix(in srgb, ${color} 16%, transparent)`,
  };
}
</script>

<template>
  <div class="relative h-full w-full">
    <Handle
      v-for="side in sides"
      :id="`source-${side.id}`"
      :key="`source-${side.id}`"
      type="source"
      :position="side.position"
      :connectable="false"
      aria-hidden="true"
      class="pointer-events-none absolute size-2 border-none bg-transparent opacity-0"
    />
    <Handle
      v-for="side in sides"
      :id="`target-${side.id}`"
      :key="`target-${side.id}`"
      type="target"
      :position="side.position"
      :connectable="false"
      aria-hidden="true"
      class="pointer-events-none absolute size-2 border-none bg-transparent opacity-0"
    />
    <button
      type="button"
      :class="
        cn(
          'focus-visible:outline-signal relative flex h-full w-full cursor-grab items-center gap-2.5 rounded-lg border px-3.5 text-left transition-[border-color,background-color,color,opacity] duration-150 ease-out focus-visible:outline-2 focus-visible:outline-offset-[3px] active:cursor-grabbing',
          data.dimmed && 'opacity-40',
          data.focused
            ? 'border-signal bg-surface text-heading ring-signal/25 ring-1'
            : 'border-default bg-panel text-secondary hover:border-strong hover:text-primary',
        )
      "
      :aria-label="`${props.id}. Drag to reposition. Hold Alt and use an arrow key to move. Activate to inspect.`"
      :aria-pressed="props.data.focused"
      :title="props.id"
      aria-keyshortcuts="Alt+ArrowUp Alt+ArrowDown Alt+ArrowLeft Alt+ArrowRight"
      @click.stop="props.data.onSelect"
      @keydown.stop="handleKeydown"
    >
      <span
        class="size-2 shrink-0 rounded-full"
        :style="dotStyle(props.id)"
        aria-hidden="true"
      />
      <span
        class="min-w-0 flex-1 truncate font-sans text-[0.8125rem] font-medium"
      >
        {{ props.id }}
      </span>
    </button>
  </div>
</template>
