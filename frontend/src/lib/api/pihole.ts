import apiFetch from './client';
import { PiholeNode } from '../../types/pihole';
import { HttpScheme } from '../../types';

type PiholeNodeDraft = Omit<PiholeNode, 'id'>;

export async function getPiholeNodes(): Promise<PiholeNode[]> {
	return apiFetch<PiholeNode[]>('/piholes');
}

export type PiholeCreateBody = PiholeNodeDraft & {
	password: string;
};
export async function createPiholeNode(nodeDraft: PiholeCreateBody): Promise<PiholeNode> {
	return apiFetch<PiholeNode>('/piholes', {
		method: 'POST',
		body: JSON.stringify(nodeDraft),
	});
}

export async function deletePiholeNode(id: number): Promise<void> {
	return apiFetch<void>(`/piholes/${id}`, {
		method: 'DELETE',
	});
}

export type PiholePatchBody = Partial<PiholeNodeDraft> & {
	password?: string;
};
export async function editPiholeNode(id: number, nodeDraft: PiholePatchBody): Promise<PiholeNode> {
	// TODO: should I escape id?
	return apiFetch<PiholeNode>(`/piholes/${id}`, {
		method: 'PATCH',
		body: JSON.stringify(nodeDraft),
	});
}

export interface PiholeTestConnectionBody {
	scheme: HttpScheme;
	host: string;
	port: number;
	password: string;
}

export async function testPiholeInstanceConnection(node: PiholeTestConnectionBody) {
	return apiFetch<void>('/piholes/test', {
		method: 'POST',
		body: JSON.stringify(node),
	});
}

export type PiholeTestExistingConnectionBody = Partial<PiholeTestConnectionBody>;
export async function testExistingPiholeConnection(
	id: number,
	overrides: PiholeTestExistingConnectionBody,
) {
	return apiFetch<void>(`/piholes/${id}/test`, {
		method: 'POST',
		body: JSON.stringify(overrides),
	});
}
