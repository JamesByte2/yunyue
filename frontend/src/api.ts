import axios from "axios";

export const api = axios.create({ baseURL: "/api" });

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("yy_token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(undefined, (error) => {
  if (error.response?.status === 401) {
    localStorage.removeItem("yy_token");
    localStorage.removeItem("yy_user");
    location.hash = "#/login";
  }
  return Promise.reject(error);
});

export function money(cents: number): string {
  return (cents / 100).toFixed(2);
}

export function minToTime(min: number): string {
  return `${String(Math.floor(min / 60)).padStart(2, "0")}:${String(min % 60).padStart(2, "0")}`;
}

export const BOOKING_STATUS: Record<string, { label: string; type: "warning" | "primary" | "success" | "info" }> = {
  pending: { label: "待确认", type: "warning" },
  confirmed: { label: "已确认", type: "primary" },
  done: { label: "已完成", type: "success" },
  canceled: { label: "已取消", type: "info" },
};
