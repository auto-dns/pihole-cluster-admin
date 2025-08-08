import apiFetch from './client';
import { FullInitStatus, PiholeInitStatus } from '../../types/initialization';
import { User } from '../../types/user';

export async function createUser(username: string, password: string): Promise<User> {
	return apiFetch<User>('/setup/user', {
		body: JSON.stringify({ username, password }),
		method: 'POST',
	});
}

export async function getPublicInitStatus(): Promise<boolean> {
	const res = await apiFetch<{ initialized: boolean }>('/setup/initialized');
	return res.initialized;
}

export async function getFullInitStatus(): Promise<FullInitStatus> {
	return apiFetch<FullInitStatus>('/setup/status');
}

export async function updatePiholeInitStatus(status: PiholeInitStatus): Promise<void> {
	return apiFetch<void>('/setup/status/pihole', {
		body: JSON.stringify({ status }),
		method: 'PATCH',
	});
}
