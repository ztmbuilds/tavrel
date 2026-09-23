<script setup lang="ts">
import type { StyleValue } from "vue";

import chevronDownIcon from "@/assets/icons/chevron-down.svg?component";

import { cn } from "@/shared/utils/cn";
import type { AttributeVisibleNode } from "../composables/useAttributeTree";

import AttributeTreeNode from "./AttributeTreeNode.vue";
import CopyButton from "./CopyButton.vue";

defineProps<{
  node: AttributeVisibleNode;
  rowStyle: StyleValue | undefined;
  copiedPath: string | null;
}>();

const emit = defineEmits<{
  copy: [path: string, text: string];
  toggle: [path: string];
}>();
</script>

<template>
  <div v-if="node.kind === 'group'" :style="rowStyle">
    <button
      type="button"
      :aria-expanded="node.expanded"
      :aria-label="`${node.expanded ? 'Collapse' : 'Expand'} ${node.fullPath}`"
      class="border-subtle hover:bg-hover flex w-full cursor-pointer items-center gap-1.5 border-x-0 border-t-0 border-b bg-transparent py-2 text-left"
      @click="$emit('toggle', node.fullPath)"
    >
      <component
        :is="chevronDownIcon"
        aria-hidden="true"
        :class="
          cn(
            'text-tertiary size-3 shrink-0 transition-transform duration-120 ease-out',
            !node.expanded && '-rotate-90',
          )
        "
      />
      <span class="text-secondary shrink-0 font-mono text-xs font-medium">{{
        node.name
      }}</span>
      <span
        class="text-tertiary shrink-0 font-mono text-[0.6875rem] tabular-nums"
        >{{ node.childCount }}</span
      >
      <span
        v-if="node.value !== null"
        class="text-primary ml-auto min-w-0 flex-1 text-right font-mono text-xs break-all"
      >
        {{ node.value }}
      </span>
    </button>
    <div
      v-if="node.expanded"
      data-attr-children
      class="border-l-default ml-[0.9375rem] border-l"
    >
      <AttributeTreeNode
        v-for="child in node.children"
        :key="child.fullPath"
        :node="child"
        :row-style="rowStyle"
        :copied-path="copiedPath"
        @copy="(path, text) => emit('copy', path, text)"
        @toggle="$emit('toggle', $event)"
      />
    </div>
  </div>
  <div
    v-else
    data-attr-leaf
    :style="rowStyle"
    class="group border-subtle flex items-center gap-2 border-x-0 border-t-0 border-b py-2"
  >
    <span class="text-secondary w-32 shrink-0 font-mono text-xs break-words">{{
      node.name
    }}</span>
    <span
      class="text-primary min-w-0 flex-1 font-mono text-xs break-all tabular-nums"
      >{{ node.value }}</span
    >
    <CopyButton
      :label="`Copy value of ${node.fullPath}`"
      title="Copy value"
      :copied="copiedPath === node.fullPath"
      @copy="$emit('copy', node.fullPath, node.value)"
    />
  </div>
</template>

<style scoped>
[data-attr-children] > [data-attr-leaf] {
  padding-left: 0.875rem;
}

[data-attr-children] > [data-attr-leaf]::before {
  content: "";
  position: absolute;
  top: 50%;
  left: 0;
  width: 0.625rem;
  height: 1px;
  background: var(--color-default);
  transform: translateY(-50%);
}

[data-attr-leaf] {
  position: relative;
}
</style>
