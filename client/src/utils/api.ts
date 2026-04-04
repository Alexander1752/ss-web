const DEFAULT_BASE_URL = 'http://127.0.0.1:8080';

export const normalize = (url: string) => url.replace(/\/+$/, '');

export const API_BASE_URL = normalize(import.meta.env.VITE_API_BASE_URL ?? DEFAULT_BASE_URL);

export const buildUrl = (path: string, baseUrl: string = API_BASE_URL) => {
  if (path.startsWith('http')) {
    return path;
  }
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return `${baseUrl}${normalizedPath}`;
};

export const apiFetch = (path: string, init?: RequestInit) => fetch(buildUrl(path), init);

