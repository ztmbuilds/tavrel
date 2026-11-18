<script setup lang="ts">
import orbitArt from "@/assets/auth/orbit.svg";
import orbitStatic from "@/assets/auth/orbit-static.svg";
import logo from "@/assets/icons/logo.svg";
</script>

<template>
  <aside
    class="border-subtle bg-auth-panel relative min-h-[22.5rem] flex-[1.35_1_35rem] overflow-hidden rounded-3xl border"
  >
    <picture>
      <source
        media="(prefers-reduced-motion: reduce)"
        :srcset="orbitStatic"
        type="image/svg+xml"
      />
      <img
        :src="orbitArt"
        alt=""
        aria-hidden="true"
        class="pointer-events-none absolute top-[46%] left-1/2 size-[45rem] -translate-x-1/2 -translate-y-1/2"
      />
    </picture>
    <img :src="logo" alt="tavrel" class="absolute top-8 left-9 size-9" />
    <div
      class="from-auth-panel absolute inset-x-0 bottom-0 h-[13.75rem] bg-linear-to-t from-8% to-transparent"
    ></div>
    <div class="absolute inset-x-9 bottom-9 z-10 flex flex-col gap-2">
      <p
        class="text-heading text-3xl leading-[1.15] font-medium tracking-[-0.02em]"
      >
        Welcome back.
      </p>
      <p class="text-tertiary max-w-auth text-[0.9375rem] leading-normal">
        Your data, ready to explore.
      </p>
    </div>
  </aside>
</template>
