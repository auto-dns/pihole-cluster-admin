import { useState } from 'react';
import { useNavigate } from 'react-router';
import { useAuth } from '../app/AuthProvider';
import useInput from '../lib/hooks/useInput';

export default function Login() {
    const {login} = useAuth();
    const username = useInput('');
    const password = useInput('');
    const navigate = useNavigate();

    async function submitForm() {
        try {
            await login(username.value, password.value);
            // TODO: update to accept redirect param and use if present
            navigate('/');
        } catch (e: unknown) {
            if (e instanceof Error) {
                alert(e.message);
                console.error(e);
            } else {
                alert('Error: see console log for full error output')
                console.error(e);
            }
        }
    }

    return (
        <div>
            <h1>Login</h1>
            <div>
                <label>
                    Username
                    <input name='username' value={username.value} onChange={username.onChange}/>
                </label>
                <label>
                    Password
                    <input name='password' type='password' value={password.value} onChange={password.onChange}/>
                </label>
                <button onClick={submitForm}>Log In</button>
            </div>
        </div>
    );
}