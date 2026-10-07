import { json } from '@sveltejs/kit';
import { apiURL, signedHeaders } from '$lib/server/api';
import type { RequestHandler } from './$types';
export const GET: RequestHandler = async ({ params, getClientAddress }) => {
	if (!/^\d{8}$/.test(params.cep))
		return json({ message: 'Informe um CEP com 8 dígitos.' }, { status: 400 });
	const path = `/api/v1/addresses/${params.cep}`;
	try {
		const response = await fetch(apiURL(path), {
			headers: signedHeaders('GET', path, getClientAddress()),
			signal: AbortSignal.timeout(7000),
			redirect: 'error',
		});
		const result = await response.json();
		if (!response.ok)
			return json(
				{
					message:
						result.error?.message ||
						'Não foi possível consultar o CEP. Preencha o endereço manualmente.',
				},
				{ status: response.status, headers: { 'Cache-Control': 'no-store' } },
			);
		return json(result, { headers: { 'Cache-Control': 'public, max-age=3600' } });
	} catch {
		return json(
			{ message: 'Não foi possível consultar o CEP. Preencha o endereço manualmente.' },
			{ status: 503, headers: { 'Cache-Control': 'no-store' } },
		);
	}
};
