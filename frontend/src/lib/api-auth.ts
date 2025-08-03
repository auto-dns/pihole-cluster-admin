import apiFetch from './api-client';

export async function login(username: string, password: string) {
    return apiFetch('/login', {
        body: JSON.stringify({username, password}),
        method: 'POST',
    })
}
