const RAW_API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

export const API_BASE_URL = RAW_API_BASE_URL.replace(/\/$/, "");

export function apiUrl(path) {
  // If a fully-qualified URL is passed, return it unchanged
  if (/^https?:\/\//i.test(path)) return path;

  // If an API base is configured at build time, prefix it
  if (API_BASE_URL) {
    return `${API_BASE_URL}${path.startsWith("/") ? path : "/" + path}`;
  }

  // No base configured -> use same-origin relative path
  return path.startsWith("/") ? path : "/" + path;
}

export function wsUrl(path) {
  // If API_BASE_URL was set and looks like an absolute http(s) URL, convert to ws(s)
  if (API_BASE_URL && /^https?:\/\//i.test(API_BASE_URL)) {
    const wsBase = API_BASE_URL.replace(/^http/i, "ws");
    return `${wsBase}${path.startsWith("/") ? path : "/" + path}`;
  }

  // Otherwise use same-origin protocol/host and the backend's ws endpoint
  if (typeof window !== "undefined") {
    const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    return `${proto}//${window.location.host}${path.startsWith("/") ? path : "/" + path}`;
  }

  // Fallback: return the path as-is
  return path;
}
