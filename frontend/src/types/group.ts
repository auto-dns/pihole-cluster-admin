import type { PiholeNodeRef } from './pihole';

export type Group = {
	id: number;
	name: string;
	description: string | null;
	enabled: boolean;
	dateAdded: string;
	dateModified: string;
};

// Per-node result

type ListGroupsNode = {
	node: PiholeNodeRef;
	groups: Group[];
	error?: string;
};

type ListGroupsSummary = {
	totalNodes: number;
	okNodes: number;
	errorNodes: number;
};

export type ListGroupsResponse = {
	summary: ListGroupsSummary;
	nodes: Record<string, ListGroupsNode>;
};

type GroupMutateNode = {
	node: PiholeNodeRef;
	groups: Group[];
	error?: string;
};

export type GroupMutateResponse = {
	nodes: Record<string, GroupMutateNode>;
};

type RemoveGroupNode = {
	node: PiholeNodeRef;
	removed: boolean;
	error?: string;
};

type RemoveGroupSummary = {
	total: number;
	removed: number;
	failed: number;
};

export type RemoveGroupResponse = {
	summary: RemoveGroupSummary;
	nodes: Record<string, RemoveGroupNode>;
};
