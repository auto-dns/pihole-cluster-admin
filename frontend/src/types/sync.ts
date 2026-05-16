export type SyncNodeResult = {
	nodeId: number;
	nodeName: string;
	added: number;
	removed: number;
	success: boolean;
	error?: string;
};

export type SyncResponse = {
	sourceNodeId: number;
	nodes: SyncNodeResult[];
};
