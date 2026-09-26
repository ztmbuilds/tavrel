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
          'bg-raised relative flex h-full w-full cursor-grab items-center gap-3 rounded-xl border px-4 text-left shadow-lg shadow-black/30 transition-opacity duration-150 ease-out active:cursor-grabbing',
          data.dimmed && 'opacity-30',
          data.focused && 'border-signal ring-signal/40 ring-1',
          !data.focused && 'border-default',
        )
      "
      :aria-label="`${props.id}. Drag to reposition. Hold Alt and use an arrow key to move. Activate to inspect.`"
      :aria-pressed="props.data.focused"
      aria-keyshortcuts="Alt+ArrowUp Alt+ArrowDown Alt+ArrowLeft Alt+ArrowRight"
      @click.stop="props.data.onSelect"
      @keydown.stop="handleKeydown"
    >
      <span
        class="h-6 w-1 shrink-0 rounded-full"
        :style="{ background: serviceColor(props.id) }"
        aria-hidden="true"
      />
      <span
        class="text-heading min-w-0 flex-1 font-mono text-[0.8125rem] font-medium"
      >
        {{ props.id }}
      </span>
    </button>
  </div>
</template>

<style scoped>
button:focus-visible {
  outline: 2px solid var(--color-signal);
  outline-offset: 3px;
  border-radius: 0.75rem;
}
</style>
