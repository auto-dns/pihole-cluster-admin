import { useAuth } from '../app/AuthProvider';
import { useInitializationStatus } from '../app/InitializationStatusProvider';
import useInput from '../lib/hooks/useInput';
import {createUser} from '../lib/api/setup';

export default function Login() {
    const auth = useAuth();
    const init = useInitializationStatus();
    const username = useInput('');
    const password = useInput('');
    const passwordVerify = useInput('');

    async function submitForm() {
        // TODO: Improve the UX paradigm - better input form validation UX (red text beneath inputs with bad fields? Iconography? Continue to only render on submit - clear out on for input - no UI jitter / resize when error message is added)
        const trimmedUsername = username.value.trim();
        const trimmedPassword = password.value.trim();

        if (!trimmedUsername) {
            alert('Username cannot be empty or whitespace only');
            return;
        }

        if (!trimmedPassword || trimmedPassword.trim().length < 8) {
            alert('Password must be at least 8 characters and not only whitespace');
            return;
        }

        if (password.value != passwordVerify.value) {
            alert('Passwords do not match');
            return;
        }

        try {
            const newUser = await createUser(username.value, password.value);
            await auth.setUser(newUser);
            await init.refreshPublic();
            await init.refreshFull();
        } catch (e: unknown) {
            if (e instanceof Error) {
                alert(e.message);
                console.error(e);
            } else {
                alert('Unknown error (see console)');
                console.error(e);
            }
        }
    }

    return (
        <div>
            <h1>Welcome to Pihole Cluster Admin!</h1>
            <p>Please set up an admin user to begin</p>
            <div>
                <label>
                    Username
                    <input name='username' value={username.value} onChange={username.onChange}/>
                </label>
                <label>
                    Password
                    <input name='password' type='password' value={password.value} onChange={password.onChange}/>
                </label>
                <label>
                    Verify Password
                    <input name='password-verification' type='password' value={passwordVerify.value} onChange={passwordVerify.onChange}/>
                </label>
                <button onClick={submitForm}>Create User</button>
            </div>
        </div>
    );
}