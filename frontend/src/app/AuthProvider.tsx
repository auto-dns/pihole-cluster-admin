import { ReactNode, createContext, useContext, useEffect, useState } from 'react';
import * as apiAuth from '../lib/api/auth';
import { User } from '../types'

export interface AuthContextType {
    user: User | undefined;
    loading: boolean;
    login: (username: string, password: string) => Promise<void>
    logout: () => void;
    setUser: (user: User | undefined) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | undefined>(undefined);
    const [loading, setLoading] = useState<boolean>(true);

    async function login(username: string, password: string): Promise<void> {
        await apiAuth.login(username, password);
        const user = await apiAuth.getUser();
        setUser(user);
        setLoading(false);
    }

    async function logout(): Promise<void> {
        await apiAuth.logout();
        setUser(undefined);
    }

    useEffect(() => {
        (async() => {
            try {
                const sessionUser = await apiAuth.getUser();
                setUser(sessionUser);
            } catch {
                setUser(undefined);
            } finally {
                setLoading(false);
            }
        })();
    }, []);

    const authContext = {
        user,
        loading,
        login,
        logout,
        setUser,
    }

    return (
        <AuthContext.Provider value={authContext}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth(): AuthContextType {
    const authContext = useContext(AuthContext);
    if (!authContext) {
        throw new Error('useAuth must be used within AuthProvider')
    }
    return authContext;
}
