function getFallbackApiBaseUrl() {
  if (typeof window === "undefined") {
    return "http://98.81.215.235:8080";
  }

  return `${window.location.protocol}//${window.location.hostname}:8080`;
}

const RAW_API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || getFallbackApiBaseUrl();

export const API_BASE_URL = RAW_API_BASE_URL.replace(/\/$/, "");

export function apiUrl(path) {
  return `${API_BASE_URL}${path}`;
}

export function wsUrl(path) {
  const base = API_BASE_URL || window.location.origin;
  const wsBase = base.replace(/^http/i, "ws");
  return `${wsBase}${path}`;
}
