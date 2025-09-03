import { useCallback, useState } from 'react';

export function useStampedState<T>(initial?: T) {
	const [value, setValue] = useState<T | undefined>(initial);
	const [receivedAt, setReceivedAt] = useState<number | undefined>(undefined);

	const set = useCallback((next: T) => {
		setValue(next);
		setReceivedAt(Date.now());
	}, []);

	return { value, set, receivedAt };
}
