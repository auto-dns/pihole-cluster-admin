export type ConfigDNS = {
	upstreams: string[];
	interface: string;
	port: number;
	dnssec: boolean;
	domainNeeded: boolean;
	expandHosts: boolean;
	localTTL: number;
	blockingmode: string;
	blockingipv4: string;
	blockingipv6: string;
	ratelimit: { count: number; interval: number };
	revServer: { active: boolean; cidr: string; target: string; domain: string };
	piholePTR: string;
	querylog: { enabled: boolean };
};

export type ConfigMisc = {
	privacylevel: number;
	delay_startup: number;
	nice: number;
};

export type ConfigFTL = {
	query_display: string;
	delay_startup: number;
	database: { DBinterval: number; maxDBdays: number };
};

export type ConfigWebserver = {
	api: { excludeClients: string[]; excludeDomains: string[]; maxHistory: number };
};

export type ConfigResolver = {
	resolveIPv4: boolean;
	resolveIPv6: boolean;
	networkNames: boolean;
};

export type PiholeConfig = {
	dns: ConfigDNS;
	misc: ConfigMisc;
	ftl: ConfigFTL;
	webserver: ConfigWebserver;
	resolver: ConfigResolver;
};

export type GetConfigResponse = {
	consensus: PiholeConfig | null;
	// Keys are node IDs serialized as strings (Go encodes int64 map keys as JSON strings).
	perNode: Record<string, PiholeConfig>;
	drifted: boolean;
};

export type PatchConfigRequest = {
	dns?: Partial<{
		upstreams: string[];
		interface: string;
		port: number;
		dnssec: boolean;
		domainNeeded: boolean;
		expandHosts: boolean;
		localTTL: number;
		blockingmode: string;
		blockingipv4: string;
		blockingipv6: string;
		ratelimit: { count?: number; interval?: number };
		revServer: { active?: boolean; cidr?: string; target?: string; domain?: string };
		piholePTR: string;
		querylog: { enabled?: boolean };
	}>;
	misc?: Partial<ConfigMisc>;
	ftl?: Partial<{ query_display: string; delay_startup: number; database: Partial<{ DBinterval: number; maxDBdays: number }> }>;
	webserver?: { api?: Partial<{ excludeClients: string[]; excludeDomains: string[]; maxHistory: number }> };
	resolver?: Partial<ConfigResolver>;
};

export type PatchConfigNodeResult = {
	node: { id: number; name: string; host: string };
	success: boolean;
	error?: string;
};

export type PatchConfigResponse = {
	nodes: Record<string, PatchConfigNodeResult>;
};
