import { defineConfig } from "vite";
import uni from "@dcloudio/vite-plugin-uni";

export default defineConfig({
  plugins: [uni()],
  server: {
    proxy: { "/api": { target: "http://127.0.0.1:8010", changeOrigin: true } },
  },
  h5: {
    // 部署在 nginx 子路径 /m/ 下，与管理后台同域不同路径
    router: { base: "/m/" },
    devServer: {
      proxy: { "/api": { target: "http://127.0.0.1:8010", changeOrigin: true } },
    },
  },
});
