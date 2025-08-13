import { useEffect, useState } from 'react';

export function useFreshness(lastUpdatedMs: number | undefined, windowMs: number) {
	const [fresh, setFresh] = useState(!!lastUpdatedMs && Date.now() - lastUpdatedMs <= windowMs);

	useEffect(() => {
		if (!lastUpdatedMs) {
			setFresh(false);
			return;
		}
		setFresh(true); // just got an update => fresh
		const remaining = Math.max(0, lastUpdatedMs + windowMs - Date.now());
		const id = window.setTimeout(() => setFresh(false), remaining);
		return () => clearTimeout(id);
	}, [lastUpdatedMs, windowMs]);

	return fresh;
}
