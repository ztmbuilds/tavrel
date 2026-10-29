<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router";

import xIcon from "@/assets/icons/x.svg?component";
import PreloadLink from "@/shared/components/PreloadLink.vue";
import Button from "@/shared/components/ui/Button.vue";
import { serviceColor } from "@/shared/utils/service-color";

const props = defineProps<{
  selected: string | null;
  calls: string[];
  calledBy: string[];
  tracesTo: RouteLocationRaw | null;
  logsTo: RouteLocationRaw | null;
}>();

defineEmits<{
  select: [id: string];
  clear: [];
}>();

function dotStyle(name: string): Record<string, string> {
  const color = serviceColor(name);
  return {
    background: color,
    boxShadow: `0 0 0 3px color-mix(in srgb, ${color} 16%, transparent)`,
  };
}
</script>

<template>
  <div class="flex h-full min-h-0 flex-col overflow-hidden" aria-live="polite">
    <div
      v-if="props.selected === null"
      class="text-tertiary flex flex-1 items-center justify-center px-6 text-center text-[0.8125rem] leading-relaxed"
    >
      Select a service<br />to see its connections.
    </div>
    <template v-else>
      <header
        class="border-subtle flex h-14 shrink-0 items-center gap-2.5 border-b px-4"
      >
        <span
          class="size-2 shrink-0 rounded-full"
          :style="dotStyle(props.selected)"
          aria-hidden="true"
        />
        <h2
          class="text-heading min-w-0 flex-1 truncate text-sm font-semibold"
          :title="props.selected"
        >
          {{ props.selected }}
        </h2>
        <Button
          variant="ghost"
          size="sm"
          class="size-7 shrink-0 p-0 [&_svg]:size-3.5"
          aria-label="Clear selection"
          @click="$emit('clear')"
        >
          <component :is="xIcon" aria-hidden="true" />
        </Button>
      </header>

      <div class="min-h-0 flex-1 overflow-y-auto px-4 py-3.5">
        <div class="flex flex-col gap-4">
          <section>
            <p class="text-secondary m-0 mb-1.5 text-xs leading-relaxed">
              Calls
              <span class="text-tertiary tabular-nums"
                >· {{ props.calls.length }}</span
              >
            </p>
            <div class="flex flex-col gap-0.5">
              <Button
                v-for="id in props.calls"
                :key="id"
                variant="ghost"
                size="sm"
                class="text-primary h-8 w-full justify-start gap-2 px-2 font-sans text-[0.8125rem]"
                @click="$emit('select', id)"
              >
                <span
                  class="size-1.5 shrink-0 rounded-full"
                  :style="{ background: serviceColor(id) }"
                  aria-hidden="true"
                />
                {{ id }}
              </Button>
              <p
                v-if="props.calls.length === 0"
                class="text-tertiary m-0 px-2 py-1 text-[0.8125rem]"
              >
                None
              </p>
            </div>
          </section>
          <section>
            <p class="text-secondary m-0 mb-1.5 text-xs leading-relaxed">
              Called by
              <span class="text-tertiary tabular-nums"
                >· {{ props.calledBy.length }}</span
              >
            </p>
            <div class="flex flex-col gap-0.5">
              <Button
                v-for="id in props.calledBy"
                :key="id"
                variant="ghost"
                size="sm"
                class="text-primary h-8 w-full justify-start gap-2 px-2 font-sans text-[0.8125rem]"
                @click="$emit('select', id)"
              >
                <span
                  class="size-1.5 shrink-0 rounded-full"
                  :style="{ background: serviceColor(id) }"
                  aria-hidden="true"
                />
                {{ id }}
              </Button>
              <p
                v-if="props.calledBy.length === 0"
                class="text-tertiary m-0 px-2 py-1 text-[0.8125rem]"
              >
                None
              </p>
            </div>
          </section>
        </div>
      </div>

      <div
        v-if="props.tracesTo !== null && props.logsTo !== null"
        class="border-subtle flex shrink-0 gap-2 border-t p-4"
      >
        <Button as="a" as-child variant="primary" class="flex-1">
          <PreloadLink :to="props.tracesTo" route-name="traces"
            >View traces</PreloadLink
          >
        </Button>
        <Button as="a" as-child class="flex-1">
          <PreloadLink :to="props.logsTo" route-name="logs"
            >View logs</PreloadLink
          >
        </Button>
      </div>
    </template>
  </div>
</template>
