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
                configure: (proxy, _options) => {
                    proxy.on("error", (err, _req, _res) => {
                        console.log("proxy error", err);
                    });
                    proxy.on("proxyReqWs", (proxyReq, req, socket) => {
                        console.log("WebSocket Proxy Request:", req.url);
                        socket.on("error", (err) => {
                            console.error("WebSocket socket error:", err);
                        });
                    });
                    proxy.on("open", (proxySocket) => {
                        console.log("WebSocket proxy connection opened");
                        proxySocket.on("error", (err) => {
                            console.error("WebSocket proxySocket error:", err);
                        });
                    });
                },
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
