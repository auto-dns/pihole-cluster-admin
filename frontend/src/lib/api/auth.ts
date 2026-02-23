import { apiV1Fetch } from './client';
import { User } from '@/types/user';

export async function login(username: string, password: string): Promise<User> {
	return apiV1Fetch<User>('/login', {
		method: 'POST',
		body: JSON.stringify({ username, password }),
	});
}

export async function logout() {
	return apiV1Fetch<void>('/logout', { method: 'POST' });
}
