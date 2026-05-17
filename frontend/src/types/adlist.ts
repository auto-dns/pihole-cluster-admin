import type { PiholeNodeRef } from './pihole';

export type AdlistType = 'block' | 'allow';

export type Adlist = {
	id: number;
	address: string;
	type: AdlistType;
	comment: string | null;
	groups: number[];
	enabled: boolean;
	dateAdded: string;
	dateModified: string;
	dateUpdated: string;
	number: number;
	invalidDomains: number;
	status: number;
};

export type ListAdlistsNodeResult = {
	node: PiholeNodeRef;
	lists: Adlist[];
	error?: string;
};

export type ListAdlistsSummary = {
	totalNodes: number;
	okNodes: number;
	errorNodes: number;
	totalLists: number;
};

export type ListAdlistsResponse = {
	summary: ListAdlistsSummary;
	nodes: Record<string, ListAdlistsNodeResult>;
};

export type AddAdlistResponse = {
	nodes: Record<
		string,
		{
			node: PiholeNodeRef;
			lists: Adlist[];
			error?: string;
		}
	>;
};

export type RemoveAdlistResponse = {
	summary: {
		total: number;
		removed: number;
		failed: number;
	};
	nodes: Record<
		string,
		{
			node: PiholeNodeRef;
			removed: boolean;
			error?: string;
		}
	>;
};

export type GravityRebuildResponse = {
	summary: {
		total: number;
		succeeded: number;
		failed: number;
	};
	nodes: Record<
		string,
		{
			node: PiholeNodeRef;
			success: boolean;
			error?: string;
		}
	>;
};

export type ConsolidatedAdlist = {
	id: number;
	address: string;
	type: AdlistType;
	enabled: boolean;
	comment: string | null;
	groups: number[];
	number: number;
	invalidDomains: number;
	dateUpdated: string;
	nodeIds: number[];
	totalNodes: number;
};
