import { apiV1Fetch } from './client';
import { User } from '@/types/user';

type UserDraft = Omit<User, 'id'>;

export async function getSessionUser(): Promise<User> {
	return apiV1Fetch<User>('/session/user');
}

export type UserPatchBody = Partial<UserDraft> & {
	password?: string;
};
export async function updateUser(id: number, userDraft: UserPatchBody): Promise<User> {
	return apiV1Fetch<User>(`/user/${id}`, {
		method: 'PATCH',
		body: JSON.stringify(userDraft),
	});
}

export interface UpdateUserPasswordBody {
	currentPassword: string;
	newPassword: string;
}
export async function updatePassword(id: number, body: UpdateUserPasswordBody) {
	return apiV1Fetch(`/user/${id}/password`, {
		method: 'POST',
		body: JSON.stringify(body),
	});
}
