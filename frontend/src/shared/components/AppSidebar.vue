<script setup lang="ts">
import { computed } from "vue";
import { RouterLink, useRoute } from "vue-router";

import { cn } from "@/shared/utils/cn";
import dashboardIcon from "@/assets/icons/dashboard.svg";
import dashboardActiveIcon from "@/assets/icons/dashboard-active.svg";
import logsIcon from "@/assets/icons/logs.svg";
import logsActiveIcon from "@/assets/icons/logs-active.svg";
import servicesIcon from "@/assets/icons/services.svg";
import servicesActiveIcon from "@/assets/icons/services-active.svg";
import spansIcon from "@/assets/icons/spans.svg";
import spansActiveIcon from "@/assets/icons/spans-active.svg";
import tracesIcon from "@/assets/icons/traces.svg";
import tracesActiveIcon from "@/assets/icons/traces-active.svg";

interface NavItem {
  label: string;
  to: string;
  match: readonly string[];
  icon: string;
  activeIcon: string;
}

const OBSERVE_ITEMS: readonly NavItem[] = [
  {
    label: "Dashboard",
    to: "/",
    match: ["dashboard"],
    icon: dashboardIcon,
    activeIcon: dashboardActiveIcon,
  },
  {
    label: "Services",
    to: "/services",
    match: ["services"],
    icon: servicesIcon,
    activeIcon: servicesActiveIcon,
  },
  {
    label: "Traces",
    to: "/traces",
    match: ["traces", "trace-detail"],
    icon: tracesIcon,
    activeIcon: tracesActiveIcon,
  },
  {
    label: "Spans",
    to: "/spans",
    match: ["spans"],
    icon: spansIcon,
    activeIcon: spansActiveIcon,
  },
  {
    label: "Logs",
    to: "/logs",
    match: ["logs"],
    icon: logsIcon,
    activeIcon: logsActiveIcon,
  },
];

const ITEM_CLASSES =
  "flex h-[1.875rem] w-auto cursor-pointer items-center justify-start gap-2 rounded-[0.3125rem] px-2 text-[0.78125rem] font-medium whitespace-nowrap transition-colors duration-120 ease-out hover:bg-hover hover:text-secondary max-[56.25rem]:w-8 max-[56.25rem]:justify-center max-[56.25rem]:px-0";

const ACTIVE_CLASSES =
  "bg-signal-muted text-signal hover:bg-signal-muted hover:text-signal";

const route = useRoute();
const activeRouteName = computed(() => String(route.name ?? ""));

function isActive(item: NavItem): boolean {
  return item.match.includes(activeRouteName.value);
}
</script>

<template>
  <nav
    aria-label="Primary"
    class="border-subtle bg-panel flex w-51 shrink-0 flex-col items-stretch gap-px border-r px-2 py-2.5 max-[56.25rem]:w-13 max-[56.25rem]:items-center max-[56.25rem]:px-1.5"
  >
    <div
      class="text-tertiary px-2 pt-2.5 pb-1 text-[0.625rem] font-semibold tracking-[0.08em] uppercase max-[56.25rem]:hidden"
    >
      Observe
    </div>
    <RouterLink
      v-for="item in OBSERVE_ITEMS"
      :key="item.label"
      :to="item.to"
      :title="item.label"
      :aria-label="item.label"
      :aria-current="isActive(item) ? 'page' : undefined"
      :class="cn(ITEM_CLASSES, isActive(item) && ACTIVE_CLASSES)"
    >
      <img
        :src="isActive(item) ? item.activeIcon : item.icon"
        alt=""
        aria-hidden="true"
        class="size-3.75 shrink-0"
      />
      <span
        class="overflow-hidden text-ellipsis whitespace-nowrap max-[56.25rem]:hidden"
        >{{ item.label }}</span
      >
    </RouterLink>
  </nav>
</template>
