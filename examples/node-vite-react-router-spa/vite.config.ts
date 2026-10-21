import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [tailwindcss(), reactRouter()],
  // Keep React Router's preview server and prerender request on the same address family.
  preview: {
    host: "127.0.0.1",
  },
  resolve: {
    tsconfigPaths: true,
  },
});
