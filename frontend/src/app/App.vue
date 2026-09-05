<script setup lang="ts">
import { RouterLink, RouterView } from "vue-router";
</script>

<template>
  <div class="app-shell">
    <nav class="temp-nav" aria-label="Primary">
      <RouterLink to="/">Dashboard</RouterLink>
      <RouterLink to="/services">Services</RouterLink>
      <RouterLink to="/traces">Traces</RouterLink>
      <RouterLink to="/spans">Spans</RouterLink>
      <RouterLink to="/logs">Logs</RouterLink>
    </nav>
    <main>
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app-shell {
  display: flex;
  min-height: 100vh;
}

.temp-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 204px;
  flex-shrink: 0;
  padding: 10px 8px;
  background: var(--bg-base);
  border-right: 1px solid var(--border-subtle);
}

.temp-nav a {
  padding: 6px 8px;
  border-radius: 5px;
  color: var(--text-tertiary);
  text-decoration: none;
}

.temp-nav a.router-link-active {
  background: var(--color-signal-muted);
  color: var(--color-signal);
}

main {
  flex: 1;
  min-width: 0;
  padding: 12px 16px;
}
</style>
