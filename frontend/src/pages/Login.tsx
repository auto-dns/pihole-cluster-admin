import { useState, FormEvent } from 'react';
import { useAuth } from '../providers/AuthProvider';
import useInput from '../hooks/useInput';
import '../styles/pages/login.scss';
import { HttpError } from '../types';

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
		<div className='app-centered-page'>
			<div className='app-card'>
				<h1>Login</h1>
				<form onSubmit={handleFormSubmission}>
					<div className='error-text'>{error || '\u00A0'}</div>
					<label htmlFor='login-username'>
						Username
						<input
							id='login-username'
							value={username.value}
							onChange={username.onChange}
						/>
					</label>
					<label htmlFor='login-password'>
						Password
						<input
							id='login-password'
							type='password'
							value={password.value}
							onChange={password.onChange}
						/>
					</label>
					<button type='submit'>Log In</button>
				</form>
			</div>
		</div>
	);
}
