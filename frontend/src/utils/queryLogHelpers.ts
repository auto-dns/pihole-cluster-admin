import type { MergedEntry, QueryLogResponse } from '@/types/querylog';

export function isBlockedStatus(status: string): boolean {
	return (
		status.startsWith('BLOCKED') ||
		[
			'GRAVITY',
			'REGEX',
			'BLACKLIST',
			'GRAVITY_CNAME',
			'REGEX_CNAME',
			'BLACKLIST_CNAME',
		].includes(status) ||
		status.startsWith('EXTERNAL_BLOCKED')
	);
}

export function isForwardedStatus(status: string): boolean {
	return status.includes('FORWARD');
}

export function isCachedStatus(status: string): boolean {
	return status.includes('CACHE');
}

export function presetRange(minutes: number): { from: string; until: string } {
	const until = new Date();
	const from = new Date(until.getTime() - minutes * 60_000);
	return { from: from.toISOString(), until: until.toISOString() };
}

export function formatTime(iso: string): string {
	return new Date(iso).toLocaleTimeString();
}

export function formatDate(iso: string): string {
	const d = new Date(iso);
	return `${d.toLocaleDateString()} ${d.toLocaleTimeString()}`;
}

export function statusColor(status: string): string {
	if (isBlockedStatus(status)) return 'var(--accent-danger)';
	if (isForwardedStatus(status) || isCachedStatus(status)) return 'var(--accent-success)';
	return 'var(--text-secondary)';
}

export function shortStatus(status: string): string {
	if (status === 'BLOCKED_GRAVITY') return 'Gravity';
	if (status === 'BLOCKED_REGEX') return 'Regex';
	if (status === 'BLOCKED_BLACKLIST') return 'Exact';
	if (status === 'BLOCKED_EXTERNAL_IP') return 'Ext.IP';
	if (status === 'BLOCKED_EXTERNAL_NXDOMAIN') return 'Ext.NX';
	if (status === 'BLOCKED_EXTERNAL_REFUSED') return 'Ext.RF';
	if (status === 'OK_FORWARDED') return 'Forwarded';
	if (status === 'OK_CACHE') return 'Cached';
	if (status === 'OK_RETRIED') return 'Retried';
	if (status === 'SPECIAL_DOMAIN') return 'Special';
	if (isBlockedStatus(status)) return 'Blocked';
	if (isForwardedStatus(status)) return 'Forwarded';
	if (isCachedStatus(status)) return 'Cached';
	return status;
}

export function mergeAndSort(resp: QueryLogResponse): MergedEntry[] {
	const out: MergedEntry[] = [];
	for (const n of resp.nodes) {
		if (n.success && n.page) {
			for (const e of n.page.entries) {
				out.push({ ...e, nodeId: n.node.id, nodeName: n.node.name });
			}
		}
	}
	out.sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime());
	return out;
}
