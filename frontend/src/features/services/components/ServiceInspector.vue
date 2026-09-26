<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router";

import PreloadLink from "@/shared/components/PreloadLink.vue";
import { serviceColor } from "@/shared/utils/service-color";

defineProps<{
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
</script>

<template>
  <div
    class="flex h-full min-h-0 flex-col gap-3.5 overflow-hidden p-4"
    aria-live="polite"
  >
    <div
      v-if="selected === null"
      class="text-tertiary flex flex-1 items-center justify-center px-2.5 text-center text-[0.78125rem] leading-relaxed"
    >
      Select a service<br />to see its connections.
    </div>
    <template v-else>
      <div class="flex shrink-0 items-center gap-2.5">
        <span
          class="h-2 w-2 shrink-0 rounded-[0.125rem]"
          :style="{ background: serviceColor(selected) }"
          aria-hidden="true"
        />
        <h3
          class="text-heading min-w-0 flex-1 font-mono text-sm font-medium break-all"
        >
          {{ selected }}
        </h3>
        <button
          type="button"
          aria-label="Clear selection"
          class="text-tertiary hover:text-primary ml-auto shrink-0 cursor-pointer rounded border-none bg-transparent px-1.5 py-1 text-[0.8125rem] transition-colors duration-120 ease-out"
          @click="$emit('clear')"
        >
          ✕
        </button>
      </div>
      <div class="min-h-0 flex-1 overflow-y-auto pr-1">
        <div class="flex flex-col gap-4">
          <section>
            <p
              class="text-tertiary mb-1.5 text-[0.65625rem] font-semibold tracking-wider uppercase"
            >
              Calls ({{ calls.length }})
            </p>
            <div class="flex flex-col gap-1">
              <button
                v-for="id in calls"
                :key="id"
                type="button"
                class="border-default bg-raised hover:border-strong text-primary flex w-full cursor-pointer items-center gap-2 rounded-md border px-2 py-1.5 text-left font-mono text-xs transition-colors duration-120 ease-out"
                @click="$emit('select', id)"
              >
                <span
                  class="h-1.5 w-1.5 shrink-0 rounded-[0.125rem]"
                  :style="{ background: serviceColor(id) }"
                  aria-hidden="true"
                />
                {{ id }}
              </button>
              <div
                v-if="calls.length === 0"
                class="text-tertiary font-sans text-xs"
              >
                None
              </div>
            </div>
          </section>
          <section>
            <p
              class="text-tertiary mb-1.5 text-[0.65625rem] font-semibold tracking-wider uppercase"
            >
              Called by ({{ calledBy.length }})
            </p>
            <div class="flex flex-col gap-1">
              <button
                v-for="id in calledBy"
                :key="id"
                type="button"
                class="border-default bg-raised hover:border-strong text-primary flex w-full cursor-pointer items-center gap-2 rounded-md border px-2 py-1.5 text-left font-mono text-xs transition-colors duration-120 ease-out"
                @click="$emit('select', id)"
              >
                <span
                  class="h-1.5 w-1.5 shrink-0 rounded-[0.125rem]"
                  :style="{ background: serviceColor(id) }"
                  aria-hidden="true"
                />
                {{ id }}
              </button>
              <div
                v-if="calledBy.length === 0"
                class="text-tertiary font-sans text-xs"
              >
                None
              </div>
            </div>
          </section>
        </div>
      </div>
      <div
        v-if="tracesTo !== null && logsTo !== null"
        class="border-subtle flex shrink-0 gap-2 border-t pt-3"
      >
        <PreloadLink
          :to="tracesTo"
          route-name="traces"
          class="bg-signal text-ink hover:bg-signal-hover inline-flex min-h-7 flex-1 cursor-pointer items-center justify-center rounded-md px-3 text-xs font-semibold whitespace-nowrap transition-colors duration-120 ease-out"
        >
          View traces
        </PreloadLink>
        <PreloadLink
          :to="logsTo"
          route-name="logs"
          class="border-default bg-raised text-primary hover:border-strong hover:bg-hover inline-flex min-h-7 flex-1 cursor-pointer items-center justify-center rounded-md border px-3 text-xs font-medium whitespace-nowrap transition-colors duration-120 ease-out"
        >
          View logs
        </PreloadLink>
      </div>
    </template>
  </div>
</template>
