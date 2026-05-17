import { apiV1Fetch } from './client';
import type {
	StatsSummaryResponse,
	StatsHistoryResponse,
	StatsTopDomainsResponse,
	StatsTopClientsResponse,
	StatsRange,
} from '@/types/stats';

export async function getStatsSummary(): Promise<StatsSummaryResponse> {
	return apiV1Fetch<StatsSummaryResponse>('/stats/summary');
}

export async function getStatsHistory(range: StatsRange = '24h'): Promise<StatsHistoryResponse> {
	return apiV1Fetch<StatsHistoryResponse>(`/stats/history?range=${range}`);
}

export async function getStatsTopDomains(
	range: StatsRange = '24h',
	count = 10,
): Promise<StatsTopDomainsResponse> {
	return apiV1Fetch<StatsTopDomainsResponse>(
		`/stats/top_domains?range=${range}&count=${count}`,
	);
}

export async function getStatsTopClients(
	range: StatsRange = '24h',
	count = 10,
): Promise<StatsTopClientsResponse> {
	return apiV1Fetch<StatsTopClientsResponse>(
		`/stats/top_clients?range=${range}&count=${count}`,
	);
}
