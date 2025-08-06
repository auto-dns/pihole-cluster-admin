import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { PiholeNode } from '../types/pihole';
import { getPiholeNodes } from '../lib/api/pihole';
import { useAuth } from './AuthProvider';

export interface PiholeContextType {
	piholeNodes: PiholeNode[];
	loading: boolean;
	error: Error | undefined;
}

const PiholeContext = createContext<PiholeContextType | undefined>(undefined);

export function PiholeProvider({ children }: { children: ReactNode }) {
	const { user } = useAuth();
	const [piholeNodes, setPiholeNodes] = useState<Array<PiholeNode>>([]);
	const [loading, setLoading] = useState<boolean>(false);
	const [error, setError] = useState<Error | undefined>(undefined);

	async function fetchNodes() {
		if (user) {
			setLoading(true);
			setError(undefined);
			try {
				const nodes = await getPiholeNodes();
				setPiholeNodes(nodes);
			} catch (err: unknown) {
				console.error(err);
				setError(err as Error);
			} finally {
				setLoading(false);
			}
		}
	}

	useEffect(() => {
		(async () => {
			fetchNodes();
		})();
	}, [user]);

	const piholeContext = {
		piholeNodes,
		loading,
		error,
	};

	return <PiholeContext.Provider value={piholeContext}>{children}</PiholeContext.Provider>;
}

export function usePiholes() {
	const piholeContext = useContext(PiholeContext);
	if (!piholeContext) {
		throw new Error('usePiholes must be used within PiholeProvider');
	}
	return piholeContext;
}
