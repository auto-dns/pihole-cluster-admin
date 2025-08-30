import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import svgr from 'vite-plugin-svgr';
import { fileURLToPath, URL } from 'node:url';

export default defineConfig({
	plugins: [svgr(), react()],
	resolve: {
		alias: {
			'@': fileURLToPath(new URL('./src', import.meta.url)),
		},
	},
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
