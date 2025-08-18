import { FormEvent, useState } from 'react';
import { useAuth } from '../providers/AuthProvider';
import { useInitializationStatus } from '../providers/InitializationStatusProvider';
import useInput from '../hooks/useInput';
import { createUser } from '../lib/api/setup';
import classNames from 'classnames';
import PasswordField from '../components/PasswordField/PasswordField';
import { Logo } from '@/components/Logo/Logo';
import AppCenteredPage from '@/components/Layout/AppCenteredPage';
import AppCard from '@/components/Layout/AppCard';
import styles from './SetupUserCreation.module.scss';

function ErrorText({ show, message }: { show: boolean; message: string }) {
	return <span className={styles.errorText}>{show ? message : '\u00A0'}</span>;
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
		<AppCenteredPage className={styles.setupPage}>
			<AppCard className={styles.card}>
				<Logo className={styles.logo} />
				<h1>Welcome to Pi-hole Cluster Admin!</h1>
				<p>Please set up an admin user to begin</p>
				<form onSubmit={handleFormSubmission}>
					<label htmlFor='user-creation-username'>
						Username
						<input
							id='user-creation-username'
							className={classNames({
								[styles.inputError]:
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
					<PasswordField
						label='Password'
						className={classNames({
							[styles.inputError]:
								(submitted || (touched?.password && !!password?.value.length)) &&
								!!errors?.password,
						})}
						value={password.value}
						onChange={handlePasswordChange}
						onBlur={handlePasswordBlur}
						autoComplete='current-password'
					/>
					<ErrorText
						show={
							(submitted || (touched?.password && !!password?.value.length)) &&
							!!errors?.password
						}
						message={errors.password || ''}
					/>
					<PasswordField
						label='Verify Password'
						className={classNames({
							[styles.inputError]:
								(submitted ||
									(touched?.passwordVerify && !!passwordVerify?.value.length)) &&
								!!errors?.passwordVerify,
						})}
						value={passwordVerify.value}
						onChange={handlePasswordVerifyChange}
						onBlur={handlePasswordVerifyBlur}
						autoComplete='current-password'
					/>
					<ErrorText
						show={
							(submitted || (touched?.password && !!password?.value.length)) &&
							!!errors?.passwordVerify
						}
						message={errors.passwordVerify || ''}
					/>
					<button type='submit'>Create User</button>
				</form>
			</AppCard>
		</AppCenteredPage>
	);
}
