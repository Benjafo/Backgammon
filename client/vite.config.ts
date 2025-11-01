import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
    plugins: [react()],
    server: {
        host: '0.0.0.0', // Allow Docker to access
        port: 5173,
        proxy: {
            "/api": {
                target: "http://backend-dev:8080", // Docker service name
                changeOrigin: true,
            },
        },
    },
    build: {
        outDir: "../static/dist/",
        emptyOutDir: true,
    },
});
