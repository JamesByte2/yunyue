// API 封装：H5 同源直连；微信小程序需要完整地址（上线时替换为 https 合法域名）。
// #ifdef H5
const BASE = "";
// #endif
// #ifdef MP-WEIXIN
const BASE = "http://8.216.24.166:8080";
// #endif

interface Resp<T = any> {
  statusCode: number;
  data: T;
}

function request<T = any>(method: "GET" | "POST", path: string, body?: any): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    uni.request({
      url: BASE + path,
      method,
      data: body,
      header: { "Content-Type": "application/json" },
      success: (res: Resp) => {
        if (res.statusCode >= 400) {
          const msg = (res.data && (res.data.error ?? res.data.message)) || `请求失败（${res.statusCode}）`;
          reject(new Error(msg));
        } else {
          resolve(res.data as T);
        }
      },
      fail: () => reject(new Error("网络异常，请稍后重试")),
    });
  });
}

export const apiGet = <T = any>(path: string) => request<T>("GET", path);
export const apiPost = <T = any>(path: string, body?: any) => request<T>("POST", path, body);

export function toast(title: string) {
  uni.showToast({ title, icon: "none" });
}
