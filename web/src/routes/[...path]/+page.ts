import { error } from '@sveltejs/kit';
import { slug } from '$lib/catalog';
import type { PageLoad } from './$types';
export const load: PageLoad = async ({ params, parent }) => {
	const { catalog } = await parent();
	const { products, rooms, categories } = catalog;
	const path = params.path;
	const allowed = [
		'moveis',
		'ambientes',
		'favoritos',
		'carrinho',
		'checkout',
		'conta',
		'conta/pedidos',
		'conta/recuperar-senha',
		'conta/redefinir-senha',
		'conta/confirmar-email',
		'sobre',
		'contato',
		'ajuda',
		'privacidade',
		'termos',
		'pedido-confirmado',
		'design-system',
	];
	if (
		!allowed.includes(path) &&
		!products.some((p) => path === `moveis/${p.id}`) &&
		!rooms.some((r) => path === `ambientes/${slug(r)}`) &&
		!categories.some((c) => path === `moveis/categoria/${slug(c)}`)
	)
		error(404, 'Página não encontrada');
	return { path };
};
