import { defineConfig } from 'vite';
// import { fileURLToPath } from 'url';
// import { dirname, resolve } from 'path';

// const __filename = fileURLToPath(import.meta.url);
// const __dirname = dirname(__filename);

export default defineConfig({
	build: {
		outDir: 'dist',
		emptyOutDir: true,
	},
	server: {
		host: '0.0.0.0',
		port: 5174,
		proxy: {
			'/api/events': {
				target: 'http://localhost:8081',
				changeOrigin: true,
				ws: false,
				proxyTimeout: 0,
				timeout: 0,
				configure: (proxy) => {
					proxy.on('proxyRes', (res) => {
						delete (res.headers as any)['content-length'];
						res.headers['cache-control'] = 'no-cache, no-transform';
						res.headers['connection'] = 'keep-alive';
					});
				},
			},
			'/api': 'http://localhost:8081',
		},
	},
});
