import { createHmac } from 'node:crypto';
import { env } from '$env/dynamic/private';
import { error } from '@sveltejs/kit';
export function signedHeaders(
	method: string,
	path: string,
	client: string,
): Record<string, string> {
	if (!env.BFF_SHARED_SECRET) {
		if (env.APP_ENV === 'production') error(503, 'Serviço temporariamente indisponível.');
		return {};
	}
	const stamp = String(Math.floor(Date.now() / 1000));
	return {
		'X-Forma-Client': client,
		'X-Forma-Timestamp': stamp,
		'X-Forma-Signature': createHmac('sha256', env.BFF_SHARED_SECRET)
			.update(`${stamp}\n${method}\n${path}\n${client}`)
			.digest('hex'),
	};
}
export function apiURL(path: string): string {
	return `${(env.API_URL || 'http://127.0.0.1:8081').replace(/\/$/, '')}${path}`;
}
