export interface validationResult {
	valid: boolean;
	reason?: string;
}

export function validateUsername(username: string, formerUsername?: string): validationResult {
	if (!username.trim()) return { valid: false, reason: 'Username cannot be empty' };
	if (formerUsername !== undefined && username.trim() === formerUsername.trim())
		return { valid: false, reason: 'New username cannot be the same as your current username' };
	return { valid: true };
}

export function validatePassword(
	password: string,
	passwordConfirmation?: string,
): validationResult {
	if (!password.trim()) return { valid: false, reason: 'Password is required' };
	if (password.trim().length < 8)
		return { valid: false, reason: 'Password must be at least 8 characters' };
	if (passwordConfirmation !== undefined && password.trim() !== passwordConfirmation.trim())
		return { valid: false, reason: 'Password and confirmation do not match' };
	return { valid: true };
}
