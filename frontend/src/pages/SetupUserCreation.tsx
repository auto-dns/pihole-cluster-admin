import { FormEvent, useState } from 'react';
import { useAuth } from '../app/AuthProvider';
import { useInitializationStatus } from '../app/InitializationStatusProvider';
import useInput from '../lib/hooks/useInput';
import { createUser } from '../lib/api/setup';
import classNames from 'classnames';
import '../styles/pages/user-setup.scss';

function ErrorText({ show, message }: { show: boolean; message: string }) {
	return <span className="error-text">{show ? message : '\u00A0'}</span>;
}

export default function Login() {
	const auth = useAuth();
	const init = useInitializationStatus();
	const username = useInput('');
	const password = useInput('');
	const passwordVerify = useInput('');
	const [errors, setErrors] = useState<{
		username: string;
		password: string;
		passwordVerify: string;
	}>({ username: '', password: '', passwordVerify: '' });
	const [touched, setTouched] = useState<{
		username: boolean;
		password: boolean;
		passwordVerify: boolean;
	}>({ username: false, password: false, passwordVerify: false });
	const [submitted, setSubmitted] = useState<boolean>(false);

	function validateUsername(value: string) {
		if (!value.trim()) {
			return 'Username cannot be empty';
		}
		return '';
	}

	function validatePassword(value: string) {
		if (!value.trim()) {
			return 'Password cannot be empty';
		}
		if (value.trim().length < 8) {
			return 'Password must be at least 8 characters';
		}
		return '';
	}

	function validatePasswordVerify(password: string, verify: string) {
		if (password !== verify) {
			return 'Passwords do not match';
		}
		return '';
	}

	function handleUsernameChange(e: React.ChangeEvent<HTMLInputElement>) {
		username.onChange(e);
		if (errors.username) {
			setErrors((prev) => ({ ...prev, username: '' }));
		}
	}

	function handlePasswordChange(e: React.ChangeEvent<HTMLInputElement>) {
		password.onChange(e);
		if (passwordVerify.value) {
			setErrors((prev) => ({
				...prev,
				passwordVerify: validatePasswordVerify(e.target.value, passwordVerify.value),
			}));
		}
		if (errors.password) {
			setErrors((prev) => ({ ...prev, password: '' }));
		}
	}

	function handlePasswordVerifyChange(e: React.ChangeEvent<HTMLInputElement>) {
		passwordVerify.onChange(e);
		if (errors.passwordVerify) {
			setErrors((prev) => ({ ...prev, passwordVerify: '' }));
		}
	}

	function handleUsernameBlur() {
		setTouched((prev) => ({ ...prev, username: true }));
		setErrors((prev) => ({ ...prev, username: validateUsername(username.value) }));
	}

	function handlePasswordBlur() {
		setTouched((prev) => ({ ...prev, password: true }));
		setErrors((prev) => ({ ...prev, password: validatePassword(password.value) }));
	}

	function handlePasswordVerifyBlur() {
		setTouched((prev) => ({ ...prev, passwordVerify: true }));
		setErrors((prev) => ({
			...prev,
			passwordVerify: validatePasswordVerify(password.value, passwordVerify.value),
		}));
	}

	function handleFormSubmission(e: FormEvent<HTMLFormElement>) {
		e.preventDefault();
		submitForm();
	}

	async function submitForm() {
		setSubmitted(true);

		const usernameError = validateUsername(username.value);
		const passwordError = validatePassword(password.value);
		const passwordVerifyError = validatePasswordVerify(password.value, passwordVerify.value);

		if (usernameError || passwordError || passwordVerifyError) {
			setErrors({
				username: usernameError,
				password: passwordError,
				passwordVerify: passwordVerifyError,
			});
			return;
		}

		try {
			const newUser = await createUser(username.value, password.value);
			await auth.setUser(newUser);
			await init.refreshPublic();
			await init.refreshFull();
		} catch (e: unknown) {
			console.error(e);
		}
	}

	return (
		<div className="user-setup-page">
			<div className="setup-card">
				<h1>Welcome to Pihole Cluster Admin!</h1>
				<p>Please set up an admin user to begin</p>
				<form onSubmit={handleFormSubmission}>
					<label htmlFor="user-creation-username">
						Username
						<input
							id="user-creation-username"
							className={classNames({
								'input-error':
									(submitted ||
										(touched?.username && !!username?.value.length)) &&
									!!errors?.username,
							})}
							value={username.value}
							onChange={handleUsernameChange}
							onBlur={handleUsernameBlur}
						/>
						<ErrorText
							show={
								(submitted || (touched?.username && !!username?.value.length)) &&
								!!errors?.username
							}
							message={errors.username || ''}
						/>
					</label>
					<label htmlFor="user-creation-password">
						Password
						<input
							id="user-creation-password"
							className={classNames({
								'input-error':
									(submitted ||
										(touched?.password && !!password?.value.length)) &&
									!!errors?.password,
							})}
							type="password"
							value={password.value}
							onChange={handlePasswordChange}
							onBlur={handlePasswordBlur}
						/>
						<ErrorText
							show={
								(submitted || (touched?.password && !!password?.value.length)) &&
								!!errors?.password
							}
							message={errors.password || ''}
						/>
					</label>
					<label htmlFor="user-creation-password-verification">
						Verify Password
						<input
							id="user-creation-password-verification"
							className={classNames({
								'input-error':
									(submitted ||
										(touched?.passwordVerify &&
											!!passwordVerify?.value.length)) &&
									!!errors?.passwordVerify,
							})}
							type="password"
							value={passwordVerify.value}
							onChange={handlePasswordVerifyChange}
							onBlur={handlePasswordVerifyBlur}
						/>
						<ErrorText
							show={
								(submitted ||
									(touched?.passwordVerify && !!passwordVerify?.value.length)) &&
								!!errors?.passwordVerify
							}
							message={errors.passwordVerify || ''}
						/>
					</label>
					<button type="submit">Create User</button>
				</form>
			</div>
		</div>
	);
}
