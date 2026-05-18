import type { PiholeNodeRef } from './pihole';

export type PiholeClient = {
	id: number;
	ip: string;
	name: string;
	comment: string | null;
	groups: number[];
	dateAdded: string;
	dateModified: string;
};

// Per-node result

type ListClientsNode = {
	node: PiholeNodeRef;
	clients: PiholeClient[];
	error?: string;
};

type ListClientsSummary = {
	totalNodes: number;
	okNodes: number;
	errorNodes: number;
};

export type ListClientsResponse = {
	summary: ListClientsSummary;
	nodes: Record<string, ListClientsNode>;
};

type ClientMutateNode = {
	node: PiholeNodeRef;
	clients: PiholeClient[];
	error?: string;
};

export type ClientMutateResponse = {
	nodes: Record<string, ClientMutateNode>;
};

type RemoveClientNode = {
	node: PiholeNodeRef;
	removed: boolean;
	error?: string;
};

type RemoveClientSummary = {
	total: number;
	removed: number;
	failed: number;
};

export type RemoveClientResponse = {
	summary: RemoveClientSummary;
	nodes: Record<string, RemoveClientNode>;
};
