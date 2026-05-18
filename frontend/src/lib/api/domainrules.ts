import { apiV1Fetch } from './client';
import type {
	ListDomainRulesResponse,
	AddDomainRuleResponse,
	RemoveDomainRuleResponse,
	SyncDomainRuleResponse,
	RuleType,
	RuleKind,
} from '@/types/domainrule';

export async function listDomainRules(): Promise<ListDomainRulesResponse> {
	return apiV1Fetch<ListDomainRulesResponse>('/domain/');
}

export async function addDomainRule(
	type: RuleType,
	kind: RuleKind,
	domain: string | string[],
	comment?: string,
	groups?: number[],
): Promise<AddDomainRuleResponse> {
	return apiV1Fetch<AddDomainRuleResponse>(`/domain/type/${type}/kind/${kind}`, {
		method: 'POST',
		body: JSON.stringify({
			domain,
			comment: comment || undefined,
			groups: groups && groups.length > 0 ? groups : undefined,
		}),
	});
}

export async function removeDomainRule(
	type: RuleType,
	kind: RuleKind,
	domain: string,
): Promise<RemoveDomainRuleResponse> {
	return apiV1Fetch<RemoveDomainRuleResponse>(
		`/domain/type/${type}/kind/${kind}/domain/${encodeURIComponent(domain)}`,
		{ method: 'DELETE' },
	);
}

export async function syncDomainRule(
	type: RuleType,
	kind: RuleKind,
	domain: string,
	comment?: string,
): Promise<SyncDomainRuleResponse> {
	return apiV1Fetch<SyncDomainRuleResponse>('/domain/parity/sync', {
		method: 'POST',
		body: JSON.stringify({ type, kind, domain, comment: comment || undefined }),
	});
}
