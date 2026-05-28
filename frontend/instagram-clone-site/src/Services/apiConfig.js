const RAW_API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

export const API_BASE_URL = RAW_API_BASE_URL.replace(/\/$/, "");

export function apiUrl(path) {
  return `${API_BASE_URL}${path}`;
}

export function wsUrl(path) {
  const base = API_BASE_URL || window.location.origin;
  const wsBase = base.replace(/^http/i, "ws");
  return `${wsBase}${path}`;
}
