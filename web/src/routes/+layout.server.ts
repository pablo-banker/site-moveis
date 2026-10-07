import { loadCatalog } from '$lib/server/catalog';
import type { LayoutServerLoad } from './$types';
export const load = (async ({ fetch, getClientAddress }) => ({
	catalog: await loadCatalog(fetch, getClientAddress()),
})) satisfies LayoutServerLoad;
