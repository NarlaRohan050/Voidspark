import { defineConfig } from "vite";

export default defineConfig({
  server: {
    port: 5173,
    proxy: {
      "/generate": "http://localhost:8080",
      "/party": "http://localhost:8080",
      "/explore": "http://localhost:8080",
      "/state": "http://localhost:8080",
      "/worlds": "http://localhost:8080",
      "/api": "http://localhost:8080",
      "/sse": "http://localhost:8080"
    }
  }
});
