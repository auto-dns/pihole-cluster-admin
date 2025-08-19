import apiFetch from './client';
import { User } from '../../types/user';

type UserDraft = Omit<User, 'id'>;

export async function getSessionUser(): Promise<User> {
	return apiFetch<User>('/session/user');
}

export type UserPatchBody = Partial<UserDraft> & {
	password?: string;
};
export async function updateUser(id: number, userDraft: UserPatchBody): Promise<User> {
	return apiFetch<User>(`/user/${id}`, {
		method: 'PATCH',
		body: JSON.stringify(userDraft),
	});
}

export interface UpdateUserPasswordBody {
	currentPassword: string;
	newPassword: string;
}
export async function updatePassword(id: number, body: UpdateUserPasswordBody) {
	return apiFetch(`/user/${id}/password`, {
		method: 'POST',
		body: JSON.stringify(body),
	});
}
