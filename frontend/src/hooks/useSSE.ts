import { useEffect, useRef } from 'react';
import { sseClient } from '../lib/sse/client';

type Handler<T = any> = (data: T) => void;

export function useSSE<T = any>(topic: string, handler: Handler<T>) {
	const handlerRef = useRef(handler);
	useEffect(() => {
		handlerRef.current = handler;
	}, [handler]);

	useEffect(() => {
		const unsubscribe = sseClient.subscribe<T>(topic, (data) => handlerRef.current(data));
		return () => unsubscribe();
	}, [topic]);
}
