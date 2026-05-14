import type { PiholeNodeRef } from './pihole';

export type QueryLogEntry = {
	id: number;
	time: string; // RFC3339
	qtype: string;
	status: string;
	dnssec: string;
	domain: string;
	upstream?: string;
	replyType: string;
	replyTimeMs: number;
	clientIp: string;
	clientName?: string;
	listId?: number;
	edeCode: number;
	edeText?: string;
	cname?: string;
};

export type QueryLogPage = {
	cursor: number;
	recordsTotal: number;
	recordsFiltered: number;
	tookMs: number;
	entries: QueryLogEntry[];
};

export type QueryLogNode = {
	node: PiholeNodeRef;
	success: boolean;
	error?: string;
	page?: QueryLogPage;
};

export type QueryLogResponse = {
	cursor: string;
	endOfResults: boolean;
	nodes: QueryLogNode[];
};

export type MergedEntry = QueryLogEntry & { nodeId: number; nodeName: string };
