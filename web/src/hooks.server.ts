import { apiURL, signedHeaders } from '$lib/server/api';
import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import type { Handle } from '@sveltejs/kit';
import { trustedOrigin } from '$lib/server/security';

export const handle: Handle = async ({ event, resolve }) => {
	trustedOrigin(event.url);
	if (
		env.APP_ENV === 'production' &&
		!event.url.pathname.startsWith('/_app/') &&
		!/\.(?:png|svg|webp|jpg|ico|woff2)$/.test(event.url.pathname)
	) {
		let status = 503;
		try {
			const check = await fetch(apiURL('/internal/traffic'), {
				method: 'POST',
				headers: signedHeaders('POST', '/internal/traffic', event.getClientAddress()),
				signal: AbortSignal.timeout(3000),
				redirect: 'error',
			});
			status = check.status;
		} catch {}
		if (status === 429) error(429, 'Muitas solicitações. Aguarde um minuto.');
		if (status !== 200) error(503, 'Serviço temporariamente indisponível.');
	}
	const response = await resolve(event);
	response.headers.set('X-Content-Type-Options', 'nosniff');
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set(
		'Referrer-Policy',
		event.url.pathname.startsWith('/conta/') ? 'no-referrer' : 'strict-origin-when-cross-origin',
	);
	response.headers.set(
		'Permissions-Policy',
		'camera=(), microphone=(), geolocation=(), payment=()',
	);
	if (
		event.url.pathname.startsWith('/backend/') ||
		/^\/(conta|checkout|pedido-confirmado)(\/|$)/.test(event.url.pathname)
	) {
		response.headers.set('Cache-Control', 'private, no-store');
	}
	if (env.APP_ENV === 'production')
		response.headers.set('Strict-Transport-Security', 'max-age=31536000');
	return response;
};
