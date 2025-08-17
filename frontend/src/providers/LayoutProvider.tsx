import { createContext, useContext, useEffect, useState } from 'react';
import { useLocalStorageState } from '../hooks/useLocalStorageState';
import { useMediaQuery } from '../hooks/useMediaQuery';

type LayoutCtx = {
	isMobile: boolean;
	sidebarOpen: boolean;
	setSidebarOpen: (v: boolean | ((v: boolean) => boolean)) => void;
};

const LayoutContext = createContext<LayoutCtx | null>(null);
export const useLayout = () => {
	const ctx = useContext(LayoutContext);
	if (!ctx) throw new Error('useLayout must be used within LayoutProvider');
	return ctx;
};

export function LayoutProvider({ children }: { children: React.ReactNode }) {
	const isMobile = useMediaQuery('(max-width: 768px)');

	const [sidebarOpenDesktop, setSidebarOpenDesktop] = useLocalStorageState<boolean>(
		'pihole-cluster-admin.sidebarOpen',
		true,
		{ syncAcrossTabs: true },
	);
	const [sidebarOpenMobile, setSidebarOpenMobile] = useState<boolean>(false);

	const sidebarOpen = isMobile ? sidebarOpenMobile : sidebarOpenDesktop;
	const setSidebarOpen = (v: boolean | ((v: boolean) => boolean)) => {
		if (isMobile) setSidebarOpenMobile(v);
		else setSidebarOpenDesktop(v);
	};

	useEffect(() => {
		if (isMobile) setSidebarOpenMobile(false);
	}, [isMobile]);

	return (
		<LayoutContext.Provider value={{ isMobile, sidebarOpen, setSidebarOpen }}>
			{children}
		</LayoutContext.Provider>
	);
}
