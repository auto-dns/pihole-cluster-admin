import { apiV1Fetch } from './client';
import type {
	ListClientsResponse,
	ClientMutateResponse,
	RemoveClientResponse,
} from '@/types/pihole_client';

export async function listPiholeClients(): Promise<ListClientsResponse> {
	return apiV1Fetch<ListClientsResponse>('/clients/');
}

export async function updatePiholeClient(
	id: number,
	patch: { groups: number[]; comment?: string | null },
): Promise<ClientMutateResponse> {
	return apiV1Fetch<ClientMutateResponse>(`/clients/${id}`, {
		method: 'PUT',
		body: JSON.stringify(patch),
	});
}

export async function removePiholeClient(id: number): Promise<RemoveClientResponse> {
	return apiV1Fetch<RemoveClientResponse>(`/clients/${id}`, { method: 'DELETE' });
}
