import { FullInitStatus, PiholeInitStatus } from '../types/initialization';
import * as api from '../lib/api/setup';
import { ReactNode, createContext, useEffect, useContext, useState } from 'react';
import { useAuth } from '../providers/AuthProvider';

export interface InitStatusContextType {
	publicStatus: boolean;
	fullStatus: FullInitStatus | undefined;
	publicLoading: boolean;
	fullLoading: boolean;
	updatingPiholeStatus: boolean;
	refreshPublic: () => Promise<void>;
	refreshFull: () => Promise<void>;
	updatePiholeInitStatus: (status: PiholeInitStatus, triggerRefresh: boolean) => Promise<void>;
}

const InitStatusContext = createContext<InitStatusContextType | undefined>(undefined);

export function InitStatusProvider({ children }: { children: ReactNode }) {
	const { user } = useAuth();
	const [publicStatus, setPublicStatus] = useState<boolean>(false);
	const [fullStatus, setFullStatus] = useState<FullInitStatus | undefined>(undefined);
	const [publicLoading, setPublicLoading] = useState<boolean>(true);
	const [fullLoading, setFullLoading] = useState<boolean>(false);
	const [updatingPiholeStatus, setUpdatingPiholeStatus] = useState<boolean>(false);

	async function refreshPublic() {
		setPublicLoading(true);
		try {
			const initialized = await api.getPublicInitStatus();
			setPublicStatus(initialized);
		} finally {
			setPublicLoading(false);
		}
	}

	async function refreshFull() {
		if (user) {
			setFullLoading(true);
			try {
				const initStatus = await api.getFullInitStatus();
				setFullStatus(initStatus);
			} finally {
				setFullLoading(false);
			}
		} else {
			setFullStatus(undefined);
		}
	}

	async function updatePiholeInitStatus(
		status: PiholeInitStatus,
		triggerRefresh: boolean = true,
	) {
		setUpdatingPiholeStatus(true);
		try {
			await api.updatePiholeInitStatus(status);
			if (triggerRefresh) {
				await refreshFull();
			}
		} finally {
			setUpdatingPiholeStatus(false);
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
		updatingPiholeStatus,
		refreshPublic,
		refreshFull,
		updatePiholeInitStatus,
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
