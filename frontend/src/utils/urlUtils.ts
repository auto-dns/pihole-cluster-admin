// src/lib/urlUtils.ts
import type { HttpScheme } from '../types'; // 'http' | 'https'

export type ParsedPiholeUrl = {
	scheme: HttpScheme; // 'http' | 'https'
	host: string; // 'pi.hole' | '192.168.0.10'
	port: number; // 1..65535
};

const DEFAULT_PORT = {
	http: 80,
	https: 443,
} as const;

function normalizeScheme(s: string | null | undefined): HttpScheme {
	return s?.toLowerCase().startsWith('https') ? 'https' : 'http';
}

/**
 * Turn a single input (like "pi.hole", "192.168.0.1:8080", "https://x:8443")
 * into { scheme, host, port } with sensible defaults.
 */
export function parsePiholeUrl(rawInput: string): ParsedPiholeUrl {
	const input = rawInput.trim();

	// If there’s no scheme, assume http so URL() can parse it.
	const hasScheme = /^https?:\/\//i.test(input);
	const withScheme = hasScheme ? input : `http://${input}`;

	let url: URL;
	try {
		url = new URL(withScheme);
	} catch {
		throw new Error('Invalid URL. Example: "pi.hole", "192.168.0.1:8080", "https://host:443"');
	}

	const scheme = normalizeScheme(url.protocol);
	const host = url.hostname;
	if (!host) throw new Error('Host is required.');

	// url.port is '' when not specified
	const port = url.port ? Number(url.port) : DEFAULT_PORT[scheme];

	if (!Number.isInteger(port) || port < 1 || port > 65535) {
		throw new Error('Port must be between 1 and 65535.');
	}

	return { scheme, host, port };
}

/**
 * Build a nice, compact string for the input from parts.
 * We avoid showing default ports to keep edits clean:
 *  - http://host:80  -> "host"
 *  - https://host:443 -> "https://host"
 *  - custom ports     -> "scheme//host:port"
 */
export function formatPiholeUrl(parts: ParsedPiholeUrl): string {
	const isDefault = parts.port === DEFAULT_PORT[parts.scheme];
	if (parts.scheme === 'http' && isDefault) {
		return parts.host; // bare host
	}
	if (parts.scheme === 'https' && isDefault) {
		return `${parts.scheme}${parts.host}`; // https://host
	}
	return `${parts.scheme}${parts.host}:${parts.port}`;
}

/**
 * Helper for edit flows: turn existing node fields into the single input value.
 */
export function formatFromNode(scheme: HttpScheme, host: string, port: number): string {
	return formatPiholeUrl({ scheme, host, port });
}
