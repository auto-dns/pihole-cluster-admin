export type AuditAction =
	| 'add_domain_rule'
	| 'remove_domain_rule'
	| 'set_cluster_blocking'
	| 'set_node_blocking';

export type AuditNodeResult = {
	nodeId: number;
	nodeName: string;
	success: boolean;
	error?: string;
};

export type AuditEntry = {
	id: number;
	actor: string;
	action: AuditAction;
	targetDomain?: string;
	ruleType?: string;
	ruleKind?: string;
	blockingEnabled?: boolean;
	blockingTimer?: number;
	targetNodeId?: number;
	targetNodeName?: string;
	nodeResults: AuditNodeResult[];
	createdAt: string;
};

export type AuditListResponse = {
	entries: AuditEntry[];
	total: number;
	limit: number;
	offset: number;
};

export type RollbackNodeResult = {
	nodeId: number;
	nodeName: string;
	success: boolean;
	error?: string;
};

export type RollbackResponse = {
	originalId: number;
	nodes: RollbackNodeResult[];
};
