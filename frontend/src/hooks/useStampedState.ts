import { useCallback, useState } from 'react';

export type StampedStateSetOptions = { keepReceivedAt?: boolean };

export function useStampedState<T>(initial?: T) {
	const [value, setValue] = useState<T | undefined>(initial);
	const [receivedAt, setReceivedAt] = useState<number | undefined>(undefined);

	const set = useCallback((next: T, options?: StampedStateSetOptions) => {
		setValue(next);
		if (!options?.keepReceivedAt) {
			setReceivedAt(Date.now());
		}
	}, []);

	return { value, set, receivedAt };
}
