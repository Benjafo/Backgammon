import react from "@vitejs/plugin-react";
import path from "path";
import { defineConfig } from "vite";

export default defineConfig({
    plugins: [react()],
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
    server: {
        host: "0.0.0.0", // Allow Docker to access
        port: 5173,
        proxy: {
            "/api": {
                target: "http://backend-dev:8080", // Docker service name
                changeOrigin: true,
                ws: true, // Enable WebSocket proxying
            },
        },
        watch: {
            usePolling: true,
        },
    },
    build: {
        outDir: "../static/dist/",
        emptyOutDir: true,
    },
});
