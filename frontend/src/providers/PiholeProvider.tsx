import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { PiholeNode } from '../types/pihole';
import { PiholeCreateBody, PiholePatchBody } from '../lib/api/pihole';
import {
	getPiholeNodes,
	createPiholeNode,
	deletePiholeNode,
	editPiholeNode,
} from '../lib/api/pihole';
import { useAuth } from './AuthProvider';

export interface PiholeContextType {
	piholeNodes: PiholeNode[];
	fetchNodes: () => Promise<void>;
	addNode: (node: PiholeCreateBody) => Promise<void>;
	deleteNode: (id: number) => Promise<void>;
	editNode: (id: number, node: PiholePatchBody) => Promise<void>;
	fetchingNode: boolean;
	addingNode: boolean;
	deletingNode: boolean;
	editingNode: boolean;
	error: Error | undefined;
}

const PiholeContext = createContext<PiholeContextType | undefined>(undefined);

export function PiholeProvider({ children }: { children: ReactNode }) {
	const { user } = useAuth();
	const [piholeNodes, setPiholeNodes] = useState<Array<PiholeNode>>([]);
	const [fetchingNode, setFetchingNode] = useState<boolean>(false);
	const [addingNode, setAddingNode] = useState<boolean>(false);
	const [deletingNode, setDeletingNode] = useState<boolean>(false);
	const [editingNode, setEditingNode] = useState<boolean>(false);
	const [error, setError] = useState<Error | undefined>(undefined);

	async function fetchNodes() {
		if (user) {
			setFetchingNode(true);
			setError(undefined);
			try {
				const nodes = await getPiholeNodes();
				setPiholeNodes(nodes);
			} catch (err: unknown) {
				console.error(err);
				setError(err as Error);
			} finally {
				setFetchingNode(false);
			}
		}
	}

	async function addNode(node: PiholeCreateBody) {
		setAddingNode(true);
		try {
			const created = await createPiholeNode(node);
			setPiholeNodes((prev) => [...prev, created]);
		} finally {
			setAddingNode(false);
		}
	}

	async function deleteNode(id: number) {
		setDeletingNode(true);
		try {
			await deletePiholeNode(id);
			setPiholeNodes((prev) => prev.filter((node) => node.id != id));
		} finally {
			setDeletingNode(false);
		}
	}

	async function editNode(id: number, updatedNode: PiholePatchBody) {
		setEditingNode(true);
		try {
			const edited = await editPiholeNode(id, updatedNode);
			setPiholeNodes((prev) => [...prev.map((n) => (n.id === id ? edited : n))]);
			setEditingNode(false);
		} finally {
			setEditingNode(false);
		}
	}

	useEffect(() => {
		(async () => {
			fetchNodes();
		})();
	}, [user]);

	const piholeContext = {
		piholeNodes,
		fetchNodes,
		addNode,
		deleteNode,
		editNode,
		fetchingNode,
		addingNode,
		deletingNode,
		editingNode,
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
