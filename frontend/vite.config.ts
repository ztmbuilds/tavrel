import tailwindcss from "@tailwindcss/vite";
import vue from "@vitejs/plugin-vue";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";

const DEFAULT_APP_PORT = "5678";

function resolveAppPort(): string {
  return process.env.APP_PORT?.trim() || DEFAULT_APP_PORT;
}

// https://vite.dev/config/
export default defineConfig(() => {
  const appPort = resolveAppPort();

  return {
    plugins: [vue(), tailwindcss()],
    resolve: {
      alias: {
        "@": fileURLToPath(new URL("./src", import.meta.url)),
      },
    },
    server: {
      port: 5173,
      proxy: {
        "/api/v1": {
          target: `http://localhost:${appPort}`,
          changeOrigin: true,
        },
      },
    },
    build: {
      outDir: "dist",
      emptyOutDir: true,
    },
  };
});
