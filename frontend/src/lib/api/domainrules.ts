import { apiV1Fetch } from './client';
import type {
	ListDomainRulesResponse,
	AddDomainRuleResponse,
	RemoveDomainRuleResponse,
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
): Promise<AddDomainRuleResponse> {
	return apiV1Fetch<AddDomainRuleResponse>(`/domain/type/${type}/kind/${kind}`, {
		method: 'POST',
		body: JSON.stringify({ domain, comment: comment || undefined }),
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
