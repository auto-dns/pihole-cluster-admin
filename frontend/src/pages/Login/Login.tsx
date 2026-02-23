import { useState, FormEvent } from 'react';
import { useLocation, useNavigate } from 'react-router';
import { useAuth } from '@/providers/AuthProvider';
import { useInput } from '@/hooks/useInput';
import { HttpError } from '@/types';
import { PasswordField } from '@/components/PasswordField';
import { Logo } from '@/components/Logo';
import { AppCenteredPage } from '@/components/Layout/AppCenteredPage';
import { AppCard } from '@/components/Layout/AppCard';
import styles from './Login.module.scss';

type LocationState = { from?: Location } | undefined;

function isSafePath(p?: string | null) {
	// Only allow same-origin paths like "/x", not full URLs or protocol-relative
	return !!p && p.startsWith('/') && !p.startsWith('//');
}

function coerceSafeRedirect(locationState: LocationState, searchParam?: string): string {
	// 1) Best: precise Location from router state
	if (locationState?.from) {
		const { pathname, search, hash } = locationState.from;
		return `${pathname || '/'}${search || ''}${hash || ''}`;
	}

	// 2) Fallback: ?redirect=/some/path (if safe)
	if (isSafePath(searchParam)) return searchParam!;

	// 3) Default
	return '/';
}

export function Login() {
	const { login } = useAuth();
	const username = useInput('');
	const password = useInput('');
	const [error, setError] = useState('');
	const location = useLocation();
	const navigate = useNavigate();

	const params = new URLSearchParams(location.search);
	const redirectQuery = params.get('redirect') ?? undefined;

	function handleFormSubmission(e: FormEvent<HTMLFormElement>) {
		e.preventDefault();
		submitForm();
	}

	async function submitForm() {
		setError('');
		try {
			await login(username.value, password.value);
			const to = coerceSafeRedirect(location.state as LocationState, redirectQuery);
			navigate(to, { replace: true });
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
