import { apiV1Fetch } from './client';
import type {
	ListGroupsResponse,
	GroupMutateResponse,
	RemoveGroupResponse,
} from '@/types/group';

export async function listGroups(): Promise<ListGroupsResponse> {
	return apiV1Fetch<ListGroupsResponse>('/groups/');
}

export async function addGroup(
	name: string,
	description?: string,
	enabled = true,
): Promise<GroupMutateResponse> {
	return apiV1Fetch<GroupMutateResponse>('/groups/', {
		method: 'POST',
		body: JSON.stringify({ name, description: description || undefined, enabled }),
	});
}

export async function updateGroup(
	name: string,
	patch: { description?: string | null; enabled?: boolean },
): Promise<GroupMutateResponse> {
	return apiV1Fetch<GroupMutateResponse>(`/groups/${encodeURIComponent(name)}`, {
		method: 'PUT',
		body: JSON.stringify(patch),
	});
}

export async function removeGroup(name: string): Promise<RemoveGroupResponse> {
	return apiV1Fetch<RemoveGroupResponse>(`/groups/${encodeURIComponent(name)}`, {
		method: 'DELETE',
	});
}
