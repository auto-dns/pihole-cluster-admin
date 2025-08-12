import apiFetch from './client';
import { User } from '../../types/user';

export async function login(username: string, password: string): Promise<User> {
	return apiFetch<User>('/login', {
		method: 'POST',
		body: JSON.stringify({ username, password }),
	});
}

export async function getUser(): Promise<User> {
	return apiFetch<User>('/session/user');
}

export async function logout() {
	return apiFetch<void>('/logout');
}
