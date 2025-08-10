import apiFetch from './client';
import { PiholeCreateBody, PiholePatchBody, PiholeNode } from '../../types/pihole';

export async function getPiholeNodes(): Promise<PiholeNode[]> {
	return apiFetch<PiholeNode[]>('/piholes');
}

export async function createPiholeNode(nodeDraft: PiholeCreateBody): Promise<PiholeNode> {
	return apiFetch<PiholeNode>('/piholes', {
		body: JSON.stringify({ nodeDraft }),
		method: 'POST',
	});
}

export async function editPiholeNode(id: number, nodeDraft: PiholePatchBody): Promise<PiholeNode> {
	// TODO: should I escape id?
	return apiFetch<PiholeNode>(`/piholes/${id}`, {
		body: JSON.stringify(nodeDraft),
		method: 'PATCH',
	});
}
