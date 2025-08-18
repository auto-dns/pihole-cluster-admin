import { useState, FormEvent } from 'react';
import { useAuth } from '../../providers/AuthProvider';
import useInput from '../../hooks/useInput';
import { HttpError } from '../../types';
import PasswordField from '../../components/PasswordField/PasswordField';
import styles from './Login.module.scss';
import { Logo } from '@/components/Logo/Logo';
import AppCenteredPage from '@/components/Layout/AppCenteredPage';
import AppCard from '@/components/Layout/AppCard';

export default function Login() {
	const { login } = useAuth();
	const username = useInput('');
	const password = useInput('');
	const [error, setError] = useState('');

	function handleFormSubmission(e: FormEvent<HTMLFormElement>) {
		e.preventDefault();
		submitForm();
	}

	async function submitForm() {
		setError('');
		try {
			await login(username.value, password.value);
			// TODO: update to accept redirect param and use if present
		} catch (err: unknown) {
			console.error(err);
			if (err instanceof Error) {
				const status = (err as HttpError).status;
				if (status === 401) {
					setError(err.message || 'Invalid username or password');
				} else {
					setError(err.message || 'An unexpected error occurred');
				}
			} else {
				setError('Unknown error occurred');
			}
		}
	}

	return (
		<AppCenteredPage className={styles.loginPage}>
			<AppCard className={styles.appCard}>
				<h1 className={styles.visuallyHideOnMobile}>Login</h1>
				<Logo className={styles.logo} />
				<h2>Pi-hole Cluster Admin</h2>
				<form onSubmit={handleFormSubmission}>
					<div className={styles.errorText}>{error || '\u00A0'}</div>
					<label htmlFor='login-username'>
						Username
						<input
							id='login-username'
							value={username.value}
							onChange={username.onChange}
						/>
					</label>
					<PasswordField
						label='Password'
						value={password.value}
						onChange={password.onChange}
						autoComplete='current-password'
					/>
					<button type='submit'>Log In</button>
				</form>
			</AppCard>
		</AppCenteredPage>
	);
}
