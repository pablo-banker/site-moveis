import { signedHeaders } from '$lib/server/api';
import { env } from '$env/dynamic/private';
import { error } from '@sveltejs/kit';
import {
	products as demoProducts,
	categories as demoCategories,
	rooms as demoRooms,
	type Product,
} from '$lib/catalog';

type Taxon = { id: string; name: string };
function isProduct(value: unknown): value is Product {
	if (!value || typeof value !== 'object') return false;
	const p = value as Record<string, unknown>;
	return (
		['id', 'name', 'category', 'room', 'image', 'dimensions', 'material'].every(
			(k) => typeof p[k] === 'string',
		) &&
		typeof p.price === 'number' &&
		Number.isFinite(p.price) &&
		p.price >= 0 &&
		Array.isArray(p.finishes) &&
		p.finishes.length > 0 &&
		p.finishes.every((f) => typeof f === 'string')
	);
}
function taxa(value: unknown): Taxon[] {
	const body = value as { data?: unknown };
	if (
		!Array.isArray(body?.data) ||
		!body.data.every((t) => t && typeof t.id === 'string' && typeof t.name === 'string')
	)
		throw new Error('Invalid taxonomy response');
	return body.data;
}
export async function loadCatalog(fetcher: typeof fetch, client = '127.0.0.1') {
	if (env.CATALOG_SOURCE === 'demo')
		return { products: demoProducts, categories: demoCategories, rooms: demoRooms, source: 'demo' };
	const base = (env.API_URL || 'http://127.0.0.1:8081').replace(/\/$/, '');
	const signal = AbortSignal.timeout(6000);
	async function request(path: string): Promise<unknown> {
		const response = await fetcher(`${base}/api/v1/${path}`, {
			signal,
			headers: signedHeaders('GET', `/api/v1/${path.split('?')[0]}`, client),
			redirect: 'error',
		});
		if (!response.ok) throw new Error(`Catalog API returned ${response.status}`);
		return response.json();
	}
	try {
		const [categories, rooms] = await Promise.all([
			request('categories').then(taxa),
			request('rooms').then(taxa),
		]);
		const products: Product[] = [];
		let total = 0;
		do {
			const body = (await request(`products?limit=100&offset=${products.length}`)) as {
				data?: unknown;
				total?: unknown;
			};
			if (
				!Array.isArray(body.data) ||
				!body.data.every(isProduct) ||
				!Number.isInteger(body.total) ||
				Number(body.total) < 0
			)
				throw new Error('Invalid product response');
			if (body.data.length === 0 && products.length < Number(body.total))
				throw new Error('Incomplete catalog response');
			products.push(...body.data);
			total = Number(body.total);
		} while (products.length < total);
		return {
			products,
			categories: ['Todos', ...categories.map((t) => t.name)],
			rooms: rooms.map((t) => t.name),
			source: 'api',
		};
	} catch {
		error(503, 'O catálogo está temporariamente indisponível. Tente novamente em instantes.');
	}
}
