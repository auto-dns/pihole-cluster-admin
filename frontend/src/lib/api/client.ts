import { HttpError } from '@/types';

const API_ROOT = '/api';
const API_VERSION = 'v1';
const API_PREFIX = `${API_ROOT}/${API_VERSION}`;

function join(base: string, path: string) {
	return `${base}${path.startsWith('/') ? path : `/${path}`}`;
}

export async function apiV1Fetch<T = unknown>(path: string, options: RequestInit = {}): Promise<T> {
	const url = join(API_PREFIX, path);

	const resp = await fetch(url, {
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

// (optional) unversioned fetch for healthcheck or future global endpoints
export async function apiFetchUnversioned<T = unknown>(path: string, options: RequestInit = {}) {
	const url = join(API_ROOT, path);
	return apiV1Fetch<T>(url.replace(API_PREFIX, ''), options); // reuse logic
}
