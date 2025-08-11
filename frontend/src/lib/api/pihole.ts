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

export async function testPiholeConnection(node: PiholeTestConnectionBody) {
	return apiFetch<void>('/piholes/test', {
		method: 'POST',
		body: JSON.stringify(node),
	});
}
