import { useEffect, useRef, useState } from 'react';

type Options<T> = {
	syncAcrossTabs?: boolean;
	serialize?: (v: T) => string;
	deserialize?: (raw: string) => T;
};

export function useLocalStorageState<T>(
	key: string,
	initialValue: T | (() => T),
	opts: Options<T> = {},
) {
	const { syncAcrossTabs, serialize = JSON.stringify, deserialize = JSON.parse } = opts;
	const getInitial = () => {
		if (typeof window === 'undefined') {
			if (typeof initialValue === 'function') {
				return (initialValue as () => T)();
			} else {
				return initialValue;
			}
		}

		const raw = window.localStorage.getItem(key);
		if (raw == null) {
			if (typeof initialValue === 'function') {
				return (initialValue as () => T)();
			} else {
				return initialValue;
			}
		}

		try {
			return deserialize(raw);
		} catch {
			if (typeof initialValue === 'function') {
				return (initialValue as () => T)();
			} else {
				return initialValue;
			}
		}
	};

	// Fetch from local storage if exists, set as initial value
	const [value, setValue] = useState<T>(getInitial);

	// Handle changes
	useEffect(() => {
		try {
			window.localStorage.setItem(key, serialize(value));
		} catch {
			// Ignore quota or serialization errors
		}
	}, [key, value, serialize]);

	// Handle value changes from other tabs
	const keyRef = useRef(key);
	useEffect(() => {
		if (!syncAcrossTabs) return;
		const handler = (e: StorageEvent) => {
			if (e.key !== keyRef.current) return; // Ignore value change events for other keys
			if (e.newValue == null) return; // ignore removals
			try {
				// Update our state to match the changed value in the local storage
				setValue(deserialize(e.newValue));
			} catch {
				// ignore parse errors
			}
		};
		window.addEventListener('storage', handler);
		return () => window.removeEventListener('storage', handler);
	}, [syncAcrossTabs, deserialize]);

	return [value, setValue] as const;
}
