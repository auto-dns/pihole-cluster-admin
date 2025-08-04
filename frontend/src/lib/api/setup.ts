import { FullInitStatus } from '../../types';
import apiFetch from './client';

export interface CreateUserResponse {
    username: string;
    createdAt: string;
    updatedAt: string;
}

export async function createUser(username: string, password: string): Promise<CreateUserResponse> {
    return apiFetch<CreateUserResponse>('/setup/user', {
        body: JSON.stringify({username, password}),
        method: 'POST',
    })
}

export async function getPublicInitStatus(): Promise<boolean> {
    const res = await apiFetch<{ initialized: boolean }>('/setup/initialized');
    return res.initialized;
}

export async function getFullInitStatus(): Promise<FullInitStatus> {
    return apiFetch<FullInitStatus>('/setup/status');
}
