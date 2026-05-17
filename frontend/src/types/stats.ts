export type StatsSummary = {
	queriesTotal: number;
	queriesBlocked: number;
	blockedPercent: number;
	gravitySize: number;
	uniqueClients: number;
	uniqueDomains: number;
};

export type StatsHistoryPoint = {
	timestamp: string; // RFC3339
	total: number;
	blocked: number;
};

export type TopDomain = {
	domain: string;
	count: number;
};

export type TopClient = {
	ip: string;
	name: string;
	count: number;
};

export type StatsNode<T> = {
	node: { id: number; name: string; host: string };
	success: boolean;
	error?: string;
	data: T;
};

export type StatsSummaryResponse = {
	cluster: StatsSummary;
	nodes: StatsNode<StatsSummary>[];
};

export type StatsHistoryResponse = {
	cluster: StatsHistoryPoint[];
	nodes: StatsNode<StatsHistoryPoint[]>[];
};

export type StatsTopDomainsNodePayload = {
	topQueried: TopDomain[];
	topBlocked: TopDomain[];
};

export type StatsTopDomainsResponse = {
	clusterTopQueried: TopDomain[];
	clusterTopBlocked: TopDomain[];
	nodes: StatsNode<StatsTopDomainsNodePayload>[];
};

export type StatsTopClientsResponse = {
	clusterClients: TopClient[];
	nodes: StatsNode<TopClient[]>[];
};

export type StatsRange = '1h' | '6h' | '24h';
