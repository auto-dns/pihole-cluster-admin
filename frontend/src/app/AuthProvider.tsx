import { ReactNode, createContext, useContext, useState } from 'react';
import * as apiAuth from '../lib/api/auth';
import { User } from '../types'

export interface AuthContextType {
    user: User | undefined;
    loading: boolean;
    login: (username: string, password: string) => Promise<void>
    logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | undefined>(undefined);
    const [loading, setLoading] = useState<boolean>(true);

    async function login(username: string, password: string) {
        const user = await apiAuth.login(username, password);
        setUser(user);
        setLoading(false);
    }

    async function logout(): Promise<void> {
        await apiAuth.logout();
        setUser(undefined);
    }

    const authContext = {
        user,
        loading,
        login,
        logout,
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
