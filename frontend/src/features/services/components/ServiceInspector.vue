<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router";

import xIcon from "@/assets/icons/x.svg?component";
import PreloadLink from "@/shared/components/PreloadLink.vue";
import Button from "@/shared/components/ui/Button.vue";
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
        <Button
          variant="ghost"
          size="sm"
          class="size-6 shrink-0 p-0 [&_svg]:size-3.5"
          aria-label="Clear selection"
          @click="$emit('clear')"
        >
          <component :is="xIcon" aria-hidden="true" />
        </Button>
      </div>
      <div class="min-h-0 flex-1 overflow-y-auto pr-1">
        <div class="flex flex-col gap-4">
          <section>
            <p
              class="text-tertiary mb-1.5 text-[0.65625rem] font-semibold tracking-wider uppercase tabular-nums"
            >
              Calls ({{ calls.length }})
            </p>
            <div class="flex flex-col gap-1">
              <Button
                v-for="id in calls"
                :key="id"
                variant="default"
                size="sm"
                class="h-7 w-full justify-start gap-2 px-2 font-mono"
                @click="$emit('select', id)"
              >
                <span
                  class="h-1.5 w-1.5 shrink-0 rounded-[0.125rem]"
                  :style="{ background: serviceColor(id) }"
                  aria-hidden="true"
                />
                {{ id }}
              </Button>
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
              class="text-tertiary mb-1.5 text-[0.65625rem] font-semibold tracking-wider uppercase tabular-nums"
            >
              Called by ({{ calledBy.length }})
            </p>
            <div class="flex flex-col gap-1">
              <Button
                v-for="id in calledBy"
                :key="id"
                variant="default"
                size="sm"
                class="h-7 w-full justify-start gap-2 px-2 font-mono"
                @click="$emit('select', id)"
              >
                <span
                  class="h-1.5 w-1.5 shrink-0 rounded-[0.125rem]"
                  :style="{ background: serviceColor(id) }"
                  aria-hidden="true"
                />
                {{ id }}
              </Button>
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
        <Button as="a" as-child variant="primary" class="flex-1">
          <PreloadLink :to="tracesTo" route-name="traces"
            >View traces</PreloadLink
          >
        </Button>
        <Button as="a" as-child class="flex-1">
          <PreloadLink :to="logsTo" route-name="logs">View logs</PreloadLink>
        </Button>
      </div>
    </template>
  </div>
</template>
