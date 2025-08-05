import { useState, FormEvent } from 'react';
import { useAuth } from '../app/AuthProvider';
import useInput from '../lib/hooks/useInput';
import '../styles/pages/login.scss';

export default function Login() {
    const {login} = useAuth();
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
            console.log(1);
            await login(username.value, password.value);
            console.log(2);
            // TODO: update to accept redirect param and use if present
        } catch (err: unknown) {
            console.log(3);
            console.error(err);
            if (err instanceof Error) {
                console.log(4);
                const status = (err as any).status;
                if (status === 401) {
                    console.log(5);
                    setError(err.message || 'Invalid username or password');
                } else {
                    console.log(6);
                    setError(err.message || 'An unexpected error occurred');
                }
                console.log(7);
                console.error(err);
            } else {
                console.log(8);
                setError('Unknown error occurred');
                console.error(err);
            }
            console.log(9);
        }
        console.log(10);
    }

    return (
        <div className='login-page'>
            <div className='login-card'>
                <h1>Login</h1>
                <form onSubmit={handleFormSubmission}>
                    <div className="error-text">{error || '\u00A0'}</div>
                    <label
                        htmlFor='login-username'
                    >
                        Username
                        <input
                            id='login-username'
                            value={username.value}
                            onChange={username.onChange}
                        />
                    </label>
                    <label
                        htmlFor='login-password'
                    >
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