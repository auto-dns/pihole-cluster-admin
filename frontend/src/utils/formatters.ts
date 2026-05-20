export function formatCount(n: number): string {
	if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
	if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
	return String(n);
}

export function formatRelativeTime(unixSeconds: number): string {
	const elapsed = Math.floor(Date.now() / 1000) - unixSeconds;
	if (elapsed < 60) return 'just now';
	if (elapsed < 3600) return `${Math.floor(elapsed / 60)} min ago`;
	if (elapsed < 86400) return `${Math.floor(elapsed / 3600)} hr ago`;
	if (elapsed < 86400 * 30) return `${Math.floor(elapsed / 86400)} days ago`;
	if (elapsed < 86400 * 365) return `${Math.floor(elapsed / (86400 * 30))} mo ago`;
	return `${Math.floor(elapsed / (86400 * 365))} yr ago`;
}
