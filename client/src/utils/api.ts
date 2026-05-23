export const normalize = (url: string) => url.replace(/\/+$/, '');

export const API_BASE_URL = normalize("/api");

export const buildUrl = (path: string, baseUrl: string = API_BASE_URL) => {
  if (path.startsWith('http')) {
    return path;
  }
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return `${baseUrl}${normalizedPath}`;
};

export const apiFetch = (path: string, init?: RequestInit) => fetch(buildUrl(path), init);

