import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
export default {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter(),
		csp: {
			mode: 'auto',
			directives: {
				'default-src': ['self'],
				'script-src': ['self'],
				'style-src': ['self', 'unsafe-inline'],
				'img-src': ['self', 'data:', 'https://images.unsplash.com'],
				'font-src': ['self'],
				'connect-src':
					process.env.NODE_ENV === 'production'
						? ['self']
						: ['self', 'ws://127.0.0.1:*', 'ws://localhost:*'],
				'frame-ancestors': ['none'],
				'object-src': ['none'],
				'base-uri': ['self'],
				'form-action': ['self'],
			},
		},
	},
};
