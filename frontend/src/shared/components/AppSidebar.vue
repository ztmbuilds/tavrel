<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";

import PreloadLink from "@/shared/components/PreloadLink.vue";

import { cn } from "@/shared/utils/cn";
import Tooltip from "@/shared/components/ui/Tooltip.vue";
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
import sidebarCollapseIcon from "@/assets/icons/sidebar-collapse.svg";
import sidebarExpandIcon from "@/assets/icons/sidebar-expand.svg";

interface NavItem {
  label: string;
  to: string;
  routeName: string;
  match: readonly string[];
  icon: string;
  activeIcon: string;
}

const OBSERVE_ITEMS: readonly NavItem[] = [
  {
    label: "Dashboard",
    to: "/",
    routeName: "dashboard",
    match: ["dashboard"],
    icon: dashboardIcon,
    activeIcon: dashboardActiveIcon,
  },
  {
    label: "Services",
    to: "/services",
    routeName: "services",
    match: ["services"],
    icon: servicesIcon,
    activeIcon: servicesActiveIcon,
  },
  {
    label: "Traces",
    to: "/traces",
    routeName: "traces",
    match: ["traces", "trace-detail"],
    icon: tracesIcon,
    activeIcon: tracesActiveIcon,
  },
  {
    label: "Spans",
    to: "/spans",
    routeName: "spans",
    match: ["spans"],
    icon: spansIcon,
    activeIcon: spansActiveIcon,
  },
  {
    label: "Logs",
    to: "/logs",
    routeName: "logs",
    match: ["logs"],
    icon: logsIcon,
    activeIcon: logsActiveIcon,
  },
];

const ITEM_CLASSES =
  "flex h-[1.875rem] w-auto cursor-pointer items-center justify-start gap-2 rounded-[0.3125rem] px-2 text-[0.8125rem] font-medium whitespace-nowrap transition-colors duration-120 ease-out hover:bg-hover hover:text-secondary max-[56.25rem]:w-8 max-[56.25rem]:justify-center max-[56.25rem]:px-0";

const COLLAPSED_ITEM_CLASSES = "w-8 justify-center px-0";

const ACTIVE_CLASSES =
  "bg-active text-ink hover:bg-active hover:text-ink";

const STORAGE_KEY = "app-sidebar-collapsed";

function getInitialCollapsed(): boolean {
  try {
    return localStorage.getItem(STORAGE_KEY) === "1";
  } catch {
    return false;
  }
}

const route = useRoute();
const activeRouteName = computed(() => String(route.name ?? ""));
const isCollapsed = ref(getInitialCollapsed());
const toggleLabel = computed(() =>
  isCollapsed.value ? "Expand sidebar" : "Collapse sidebar",
);
const tooltipText = computed(() => (isCollapsed.value ? "Expand" : "Collapse"));

function isActive(item: NavItem): boolean {
  return item.match.includes(activeRouteName.value);
}

function handleToggle(): void {
  isCollapsed.value = !isCollapsed.value;
  try {
    localStorage.setItem(STORAGE_KEY, isCollapsed.value ? "1" : "0");
  } catch {
    return;
  }
}
</script>

<template>
  <nav
    aria-label="Primary"
    :class="
      cn(
        'border-subtle bg-panel flex w-51 shrink-0 flex-col items-stretch gap-0.5 overflow-y-auto border-r pt-2 pr-4 pb-4 pl-2 transition-[width] duration-200 ease-out max-[56.25rem]:w-13 max-[56.25rem]:items-center max-[56.25rem]:px-1.5',
        isCollapsed && 'w-13 items-center px-1.5',
      )
    "
  >
    <div
      :class="
        cn(
          'flex items-center justify-end pb-1 max-[56.25rem]:hidden',
          isCollapsed && 'justify-center px-0',
        )
      "
    >
      <Tooltip :label="tooltipText">
        <button
          type="button"
          :aria-expanded="!isCollapsed"
          :aria-label="toggleLabel"
          class="hover:bg-hover flex size-6 shrink-0 cursor-pointer items-center justify-center rounded-md transition-colors duration-120 ease-out"
          @click="handleToggle"
        >
          <img
            :src="isCollapsed ? sidebarExpandIcon : sidebarCollapseIcon"
            alt=""
            aria-hidden="true"
            class="size-4 shrink-0"
          />
        </button>
      </Tooltip>
    </div>

    <PreloadLink
      v-for="item in OBSERVE_ITEMS"
      :key="item.label"
      :to="item.to"
      :route-name="item.routeName"
      :title="item.label"
      :aria-label="item.label"
      :aria-current="isActive(item) ? 'page' : undefined"
      :class="
        cn(
          ITEM_CLASSES,
          isCollapsed && COLLAPSED_ITEM_CLASSES,
          isActive(item) && ACTIVE_CLASSES,
        )
      "
    >
      <img
        :src="isActive(item) ? item.activeIcon : item.icon"
        alt=""
        aria-hidden="true"
        class="size-3.75 shrink-0"
      />
      <span
        :class="
          cn(
            'overflow-hidden text-ellipsis whitespace-nowrap max-[56.25rem]:hidden',
            isCollapsed && 'hidden',
          )
        "
        >{{ item.label }}</span
      >
    </PreloadLink>
  </nav>
</template>
