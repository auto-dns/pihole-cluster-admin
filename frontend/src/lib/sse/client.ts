type Handler<T = any> = (data: T) => void;

class SSEClient {
	private es: EventSource | null = null;
	private handlers = new Map<string, Set<Handler>>();
	private attachedTopics = new Set<string>();
	private reconnectTimer: number | null = null;

	private url(): string {
		const topics = Array.from(this.handlers.keys());
		const qs = topics.length ? `topics=${encodeURIComponent(topics.join(','))}` : '';
		return `/api/events?${qs}`;
	}

	private connect() {
		if (this.es) this.es.close();
		this.es = new EventSource(this.url(), { withCredentials: true });
		this.attachedTopics.clear();

		for (const topic of this.handlers.keys()) this.attachTopicListener(topic);

		this.es.onerror = () => {
			if (this.reconnectTimer == null) {
				this.reconnectTimer = window.setTimeout(() => {
					this.reconnectTimer = null;
					this.connect();
				}, 4000);
			}
		};
	}

	private attachTopicListener(topic: string) {
		if (!this.es || this.attachedTopics.has(topic)) return;
		this.es.addEventListener(topic, (event: MessageEvent) => {
			const set = this.handlers.get(topic);
			if (!set || set.size === 0) return;
			let data: any = event.data;
			try {
				data = JSON.parse(event.data);
			} catch {
				//Nothing
			}
			for (const handler of set) handler(data);
		});
		this.attachedTopics.add(topic);
	}

	subscribe<T = any>(topic: string, handler: Handler<T>): () => void {
		let set = this.handlers.get(topic);
		if (!set) {
			set = new Set();
			this.handlers.set(topic, set);
			this.connect();
		}
		set.add(handler as Handler);

		this.attachTopicListener(topic);

		return () => {
			const s = this.handlers.get(topic);
			if (!s) return;
			s.delete(handler as Handler);
			if (s.size === 0) {
				this.handlers.delete(topic);
				if (this.handlers.size === 0) {
					this.es?.close();
					this.es = null;
					this.attachedTopics.clear();
				} else {
					this.connect();
				}
			}
		};
	}
}

const key = '__pca_sse_client__';
export const sseClient: SSEClient =
	(globalThis as any)[key] ?? ((globalThis as any)[key] = new SSEClient());
