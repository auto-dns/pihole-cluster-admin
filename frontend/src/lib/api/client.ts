import { HttpError } from '../../types';

export default async function apiFetch<T = unknown>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const resp = await fetch(`/api${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
    credentials: 'include',
  });

  const text = await resp.text();

  // Handle error responses (non-2xx)
  if (!resp.ok) {
    let message = text;
    try {
      const parsed = JSON.parse(text);
      message = parsed.error || parsed.message || message;
    } catch {
      // fallback: plain text
    }
    const err: HttpError = new Error(message || `HTTP ${resp.status}`);
    err.status = resp.status;
    throw err;
  }

  // Handle empty/no content responses
  if (resp.status === 204 || text === '' || resp.headers.get('content-length') === '0') {
    return undefined as unknown as T;
  }

  try {
    return JSON.parse(text) as T;
  } catch {
    throw new Error('Failed to parse JSON response');
  }
}
