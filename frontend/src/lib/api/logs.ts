import { apiV1Fetch } from './client';
import type { QueryLogResponse } from '@/types/querylog';

export type QueryLogParams = {
	cursor?: string;
	length?: number;
	from?: string;
	until?: string;
	domain?: string;
	clientIp?: string;
	clientName?: string;
	status?: string;
	type?: string;
};

export async function getQueryLogs(params: QueryLogParams = {}): Promise<QueryLogResponse> {
	const qs = new URLSearchParams();
	const keyMap: Record<string, string> = { clientIp: 'client_ip', clientName: 'client_name' };
	for (const [k, v] of Object.entries(params)) {
		if (v != null && v !== '') {
			qs.set(keyMap[k] ?? k, String(v));
		}
	}
	const query = qs.toString();
	return apiV1Fetch<QueryLogResponse>(`/querylog${query ? `?${query}` : ''}`);
}
