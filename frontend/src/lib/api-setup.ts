import apiFetch from './api-client';

export interface CreateUserResponse {
    username: string
    createdAt: string,
    updatedAt: string,
}

export async function createUser(username: string, password: string): Promise<CreateUserResponse> {
    return apiFetch<CreateUserResponse>('/setup/user', {
        body: JSON.stringify({username, password}),
        method: 'POST',
    })
}
