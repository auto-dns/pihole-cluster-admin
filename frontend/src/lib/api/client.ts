export default async function apiFetch<T = unknown>(path: string, options: RequestInit = {}): Promise<T> {
  const resp = await fetch(`/api${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    credentials: "include", // for cookies
  });
  
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);

  // Handle empty body (204 No Content or empty 200 response)
  const contentLength = resp.headers.get("content-length");
  if (resp.status === 204 || contentLength === "0") {
    return undefined as unknown as T;
  }

  // Some APIs return chunked encoding without content-length but still no body
  const text = await resp.text();
  if (!text) {
    return undefined as unknown as T;
  }

  return JSON.parse(text) as T;
}
