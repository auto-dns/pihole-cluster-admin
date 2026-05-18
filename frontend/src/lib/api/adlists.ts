import { apiV1Fetch } from './client';
import type {
	ListAdlistsResponse,
	AddAdlistResponse,
	RemoveAdlistResponse,
	GravityRebuildResponse,
	AdlistType,
} from '@/types/adlist';

export async function listAdlists(): Promise<ListAdlistsResponse> {
	return apiV1Fetch<ListAdlistsResponse>('/adlists/');
}

export async function addAdlist(
	address: string,
	type: AdlistType,
	comment?: string,
	enabled = true,
): Promise<AddAdlistResponse> {
	return apiV1Fetch<AddAdlistResponse>('/adlists/', {
		method: 'POST',
		body: JSON.stringify({ address, type, comment: comment || undefined, enabled }),
	});
}

export async function updateAdlist(
	id: number,
	patch: { enabled?: boolean; comment?: string | null; groups?: number[] },
): Promise<AddAdlistResponse> {
	return apiV1Fetch<AddAdlistResponse>(`/adlists/${id}`, {
		method: 'PUT',
		body: JSON.stringify(patch),
	});
}

export async function removeAdlist(id: number): Promise<RemoveAdlistResponse> {
	return apiV1Fetch<RemoveAdlistResponse>(`/adlists/${id}`, { method: 'DELETE' });
}

export async function rebuildGravity(): Promise<GravityRebuildResponse> {
	return apiV1Fetch<GravityRebuildResponse>('/gravity/rebuild', { method: 'POST' });
}
