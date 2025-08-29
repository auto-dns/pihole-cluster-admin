import apiV1Fetch from './client';
import { FullInitStatus, PiholeInitStatus } from '../../types/initialization';
import { User } from '../../types/user';

export async function createUser(username: string, password: string): Promise<User> {
	return apiV1Fetch<User>('/setup/user', {
		method: 'POST',
		body: JSON.stringify({ username, password }),
	});
}

export async function getPublicInitStatus(): Promise<boolean> {
	const res = await apiV1Fetch<{ initialized: boolean }>('/setup/initialized');
	return res.initialized;
}

export async function getFullInitStatus(): Promise<FullInitStatus> {
	return apiV1Fetch<FullInitStatus>('/setup/status');
}

export async function updatePiholeInitStatus(status: PiholeInitStatus): Promise<void> {
	return apiV1Fetch<void>('/setup/status/pihole', {
		method: 'PATCH',
		body: JSON.stringify({ status }),
	});
}
