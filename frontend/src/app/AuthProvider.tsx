import { ReactNode, createContext, useContext, useEffect, useState } from 'react';
import * as apiAuth from '../lib/api/auth';
import { User } from '../types';

export interface AuthContextType {
	user: User | undefined;
	initializing: boolean;
	authenticating: boolean;
	login: (username: string, password: string) => Promise<void>;
	logout: () => void;
	setUser: (user: User | undefined) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
	const [user, setUser] = useState<User | undefined>(undefined);
	const [initializing, setInitializing] = useState<boolean>(true);
	const [authenticating, setAuthenticating] = useState<boolean>(true);

	async function login(username: string, password: string): Promise<void> {
		setAuthenticating(true);
		try {
			const user = await apiAuth.login(username, password);
			setUser(user);
		} finally {
			setAuthenticating(false);
		}
	}

	async function logout(): Promise<void> {
		await apiAuth.logout();
		setUser(undefined);
	}

	// Called on page load
	// If we don't have a session, user will remain undefined when we mark "loading" as done
	useEffect(() => {
		(async () => {
			setInitializing(true);
			try {
				const sessionUser = await apiAuth.getUser();
				setUser(sessionUser);
			} catch {
				setUser(undefined);
			} finally {
				setInitializing(false);
			}
		})();
	}, []);

	const authContext = {
		user,
		initializing,
		authenticating,
		login,
		logout,
		setUser,
	};

	return <AuthContext.Provider value={authContext}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
	const authContext = useContext(AuthContext);
	if (!authContext) {
		throw new Error('useAuth must be used within AuthProvider');
	}
	return authContext;
}
