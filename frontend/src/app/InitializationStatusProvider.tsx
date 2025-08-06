import { FullInitStatus } from '../types';
import { getPublicInitStatus, getFullInitStatus } from '../lib/api/setup';
import { ReactNode, createContext, useEffect, useContext, useState } from 'react';
import { useAuth } from './AuthProvider';

export interface InitStatusContextType {
	publicStatus: boolean;
	fullStatus: FullInitStatus | undefined;
	publicLoading: boolean;
	fullLoading: boolean;
	refreshPublic: () => Promise<void>;
	refreshFull: () => Promise<void>;
}

const InitStatusContext = createContext<InitStatusContextType | undefined>(undefined);

export function InitStatusProvider({ children }: { children: ReactNode }) {
	const { user } = useAuth();
	const [publicStatus, setPublicStatus] = useState<boolean>(false);
	const [fullStatus, setFullStatus] = useState<FullInitStatus | undefined>(undefined);
	const [publicLoading, setPublicLoading] = useState(true);
	const [fullLoading, setFullLoading] = useState(false);

	async function refreshPublic() {
		setPublicLoading(true);
		try {
			const initialized = await getPublicInitStatus();
			setPublicStatus(initialized);
		} finally {
			setPublicLoading(false);
		}
	}

	async function refreshFull() {
		if (user) {
			setFullLoading(true);
			try {
				const initStatus = await getFullInitStatus();
				setFullStatus(initStatus);
			} finally {
				setFullLoading(false);
			}
		} else {
			setFullStatus(undefined);
		}
	}

	useEffect(() => {
		refreshPublic();
	}, []);

	useEffect(() => {
		refreshFull();
	}, [user]);

	const initStateContext = {
		publicStatus,
		fullStatus,
		publicLoading,
		fullLoading,
		refreshPublic,
		refreshFull,
	};

	return (
		<InitStatusContext.Provider value={initStateContext}>{children}</InitStatusContext.Provider>
	);
}

export function useInitializationStatus() {
	const initStatusContext = useContext(InitStatusContext);
	if (!initStatusContext) {
		throw new Error('useInitializationStatus must be used within InitStatusProvider');
	}
	return initStatusContext;
}
