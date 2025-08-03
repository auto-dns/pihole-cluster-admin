import apiFetch from './api-client';
import { User } from '../types';

export async function login(username: string, password: string): Promise<User> {
    return apiFetch('/login', {
        body: JSON.stringify({username, password}),
        method: 'POST',
    })
}

export async function getUser(): Promise<User> {
    return apiFetch('/session/user');
}

export async function logout() {
    return apiFetch('/logout')
}
