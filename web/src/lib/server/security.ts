import { env } from '$env/dynamic/private';
import { error } from '@sveltejs/kit';

export function trustedOrigin(url: URL): string {
	const mode = env.APP_ENV || 'development';
	if (!['development', 'test', 'production'].includes(mode))
		error(503, 'Serviço temporariamente indisponível.');
	if (mode !== 'production') return url.origin;
	let configured: URL;
	try {
		configured = new URL(env.SITE_ORIGIN || '');
	} catch {
		error(503, 'Serviço temporariamente indisponível.');
	}
	if (
		configured.protocol !== 'https:' ||
		configured.username ||
		configured.password ||
		configured.pathname !== '/' ||
		configured.search ||
		configured.hash ||
		url.origin !== configured.origin
	) {
		error(503, 'Serviço temporariamente indisponível.');
	}
	if (!env.API_URL || (env.BFF_SHARED_SECRET?.length || 0) < 32 || env.CATALOG_SOURCE === 'demo')
		error(503, 'Serviço temporariamente indisponível.');
	return configured.origin;
}

/** Bound bytes and time before buffering a request, including chunked requests. */
export async function limitedBody(request: Request, maxBytes = 32768): Promise<string> {
	const length = request.headers.get('content-length');
	if (length && (!/^\d+$/.test(length) || Number(length) > maxBytes))
		error(413, 'Solicitação muito grande.');
	if (!request.body) return '';
	const reader = request.body.getReader();
	let timer: ReturnType<typeof setTimeout>;
	const timeout = new Promise<never>((_, reject) => {
		timer = setTimeout(() => {
			try {
				error(408, 'Tempo de envio excedido.');
			} catch (reason) {
				reject(reason);
			}
			void reader.cancel().catch(() => {});
		}, 8000);
	});
	let bytes = 0;
	const chunks: Uint8Array[] = [];
	try {
		for (;;) {
			const part = await Promise.race([reader.read(), timeout]);
			if (part.done) break;
			bytes += part.value.byteLength;
			if (bytes > maxBytes) {
				void reader.cancel().catch(() => {});
				error(413, 'Solicitação muito grande.');
			}
			chunks.push(part.value);
		}
		const body = new Uint8Array(bytes);
		let offset = 0;
		for (const chunk of chunks) {
			body.set(chunk, offset);
			offset += chunk.length;
		}
		try {
			return new TextDecoder('utf-8', { fatal: true }).decode(body);
		} catch {
			error(400, 'Corpo da solicitação inválido.');
		}
	} finally {
		clearTimeout(timer!);
		reader.releaseLock();
	}
}
