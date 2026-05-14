import type { PiholeNodeRef } from './pihole';

export type RuleType = 'allow' | 'deny';
export type RuleKind = 'exact' | 'regex';

export type DomainRule = {
	id: number;
	domain: string;
	unicode: string;
	type: RuleType;
	kind: RuleKind;
	comment?: string | null;
	groups: number[];
	enabled: boolean;
	createdAt: string;
	updatedAt: string;
};

export type ListNodeResult = {
	node: PiholeNodeRef;
	rules: DomainRule[];
	tookMs: number;
	error?: string;
};

export type ListSummary = {
	totalNodes: number;
	okNodes: number;
	errorNodes: number;
	totalRules: number;
};

export type ListDomainRulesResponse = {
	summary: ListSummary;
	nodes: Record<string, ListNodeResult>;
};

export type AddDomainRuleResponse = {
	nodes: Record<string, {
		node: PiholeNodeRef;
		result: {
			domains: DomainRule[];
			processed: {
				success: { item: string }[];
				errors: { item: string; error: string }[];
			};
			tookMs: number;
		};
		error?: string;
	}>;
};

export type RemoveDomainRuleResponse = {
	summary: {
		total: number;
		removed: number;
		failed: number;
		errors: number;
	};
	nodes: Record<string, {
		node: PiholeNodeRef;
		removed: boolean;
		error?: string;
	}>;
};

export type ConsolidatedRule = {
	key: string;
	domain: string;
	type: RuleType;
	kind: RuleKind;
	enabled: boolean;
	comment?: string | null;
	nodeIds: number[];
	totalNodes: number;
};
