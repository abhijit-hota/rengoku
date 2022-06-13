import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte()],
  resolve: {
    alias: {
      "@lib": "./src/lib",
      "@store": "./src/lib/stores.js",
    },
  },
  build: {
    outDir: "../api/frontend-dist"
  }
});
