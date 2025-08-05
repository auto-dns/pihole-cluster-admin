import apiFetch from './client';
import { User } from '../../types';

export async function login(username: string, password: string): Promise<void> {
    return apiFetch<void>('/login', {
        body: JSON.stringify({username, password}),
        method: 'POST',
    })
}

export async function getUser(): Promise<User> {
    return apiFetch<User>('/session/user');
}

export async function logout() {
    return apiFetch<void>('/logout')
}
