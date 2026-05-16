const CACHE = 'pca-shell-v1';

self.addEventListener('install', (e) => {
	e.waitUntil(caches.open(CACHE).then((c) => c.add('/')));
	self.skipWaiting();
});

self.addEventListener('activate', (e) => {
	e.waitUntil(
		caches
			.keys()
			.then((keys) =>
				Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))),
			),
	);
	self.clients.claim();
});

// Network-first for navigation requests; serve cached shell when offline.
// API calls are never intercepted so live data is always fresh.
self.addEventListener('fetch', (e) => {
	if (e.request.url.includes('/api/')) return;
	if (e.request.mode !== 'navigate') return;
	e.respondWith(fetch(e.request).catch(() => caches.match('/')));
});
