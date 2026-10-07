import { signedHeaders, apiURL } from '$lib/server/api';
import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { limitedBody, trustedOrigin } from '$lib/server/security';
import type { RequestHandler } from './$types';
const proxy: RequestHandler = async ({
	request,
	params,
	cookies,
	url,
	fetch,
	getClientAddress,
}) => {
	const path = params.path;
	const allowed =
		request.method === 'GET'
			? /^(products\/[a-z0-9-]{1,100}\/variants|auth\/(me|sessions)|orders(?:\/[a-f0-9-]{36})?)$/
			: /^(auth\/(register|login|logout|profile|(?:verify|reset)\/(?:request|confirm)|sessions\/revoke)|contact|shipping\/quotes|orders(?:\/[a-f0-9-]{36}\/(cancel|payment))?)$/;
	if (!allowed.test(path))
		return json({ error: { message: 'Rota não encontrada.' } }, { status: 404 });
	if (
		request.method === 'POST' &&
		(request.headers.get('origin') !== trustedOrigin(url) ||
			request.headers.get('sec-fetch-site') === 'cross-site')
	)
		return json({ error: { message: 'Origem não permitida.' } }, { status: 403 });

	if (
		request.method === 'POST' &&
		request.headers.get('content-type')?.split(';')[0].trim() !== 'application/json'
	) {
		return json({ error: { message: 'Envie os dados como JSON.' } }, { status: 415 });
	}
	const body = request.method === 'POST' ? await limitedBody(request) : undefined;
	const sessionCookie = env.APP_ENV === 'production' ? '__Host-forma-session' : 'forma-session';
	const session = cookies.get(sessionCookie);

	try {
		const upstreamPath = `/api/v1/${path}`;
		const headers: Record<string, string> = {
			'Content-Type': 'application/json',
			...signedHeaders(request.method, upstreamPath, getClientAddress()),
		};
		if (session) headers.Authorization = `Bearer ${session}`;
		const response = await fetch(apiURL(upstreamPath), {
			method: request.method,
			headers,
			body,
			signal: AbortSignal.timeout(8000),
			redirect: 'error',
		});
		const result = await response.json();
		if (response.ok && path === 'auth/login') {
			if (typeof result.token !== 'string' || result.token.length !== 43)
				throw new Error('Invalid session');
			cookies.set(sessionCookie, result.token, {
				path: '/',
				httpOnly: true,
				sameSite: 'lax',
				secure: env.APP_ENV === 'production' || url.protocol === 'https:',
				maxAge: 7 * 24 * 60 * 60,
			});
			delete result.token;
		}
		if (
			(response.ok && ['auth/logout', 'auth/reset/confirm'].includes(path)) ||
			(response.status === 401 && !['auth/login', 'auth/register'].includes(path))
		)
			cookies.delete(sessionCookie, { path: '/' });
		return json(result, { status: response.status, headers: { 'Cache-Control': 'no-store' } });
	} catch {
		return json(
			{ error: { message: 'O serviço está indisponível. Tente novamente.' } },
			{ status: 503 },
		);
	}
};
export const GET = proxy;
export const POST = proxy;
