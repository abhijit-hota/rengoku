import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import sveltePreprocess from "svelte-preprocess";
import * as path from "path";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    svelte({
      preprocess: sveltePreprocess(),
    }),
  ],
  resolve: {
    alias: {
      "@components": path.resolve(__dirname, "./src/components"),
      "@lib": path.resolve(__dirname, "./src/lib"),
      "@Modal": path.resolve(__dirname, "./src/components/Modal.svelte"),
      "@utils/dev": path.resolve(__dirname, "./src/lib/dev-utils"),
      "@icons": "@fortawesome/free-solid-svg-icons",
    },
  },
  build: {
    outDir: "../api/frontend-dist",
    emptyOutDir: true,
  },
  server: {
    port: 3000,
  },
});
