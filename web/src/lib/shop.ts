import { writable, derived } from 'svelte/store';
import type { Product } from './catalog';
export type CartItem = { id: string; finish: string; quantity: number };
export const cart = writable<CartItem[]>([]);
export const favorites = writable<string[]>([]);
export const notice = writable('');
export const count = derived(cart, ($cart) => $cart.reduce((sum, item) => sum + item.quantity, 0));
export function subtotal(items: CartItem[], products: Product[]): number {
	return items.reduce(
		(sum, item) => sum + (products.find((p) => p.id === item.id)?.price ?? 0) * item.quantity,
		0,
	);
}
let timer: ReturnType<typeof setTimeout>;
export function notify(message: string) {
	notice.set(message);
	clearTimeout(timer);
	timer = setTimeout(() => notice.set(''), 4000);
}
export function add(id: string, finish: string, quantity = 1) {
	cart.update((items) => {
		const current = items.find((item) => item.id === id && item.finish === finish);
		return current
			? items.map((item) =>
					item === current ? { ...item, quantity: Math.min(10, item.quantity + quantity) } : item,
				)
			: [...items, { id, finish, quantity }];
	});
	notify('Móvel adicionado à sacola');
}
export function change(id: string, finish: string, quantity: number) {
	cart.update((items) =>
		quantity < 1
			? items.filter((item) => !(item.id === id && item.finish === finish))
			: items.map((item) =>
					item.id === id && item.finish === finish
						? { ...item, quantity: Math.min(10, quantity) }
						: item,
				),
	);
}
export function favorite(id: string) {
	favorites.update((items) =>
		items.includes(id) ? items.filter((item) => item !== id) : [...items, id],
	);
}
export function hydrate(products: Product[]) {
	try {
		const saved = JSON.parse(localStorage.getItem('forma-cart') || '[]');
		if (Array.isArray(saved))
			cart.set(
				saved.filter(
					(i) =>
						i &&
						typeof i === 'object' &&
						products.some((p) => p.id === i.id && p.finishes.includes(i.finish)) &&
						Number.isInteger(i.quantity) &&
						i.quantity > 0 &&
						i.quantity <= 10,
				),
			);
		const savedFavorites = JSON.parse(localStorage.getItem('forma-favorites') || '[]');
		if (Array.isArray(savedFavorites))
			favorites.set(savedFavorites.filter((id) => products.some((p) => p.id === id)));
	} catch {
		cart.set([]);
		favorites.set([]);
	}
	const saveCart = cart.subscribe((value) => {
		try {
			localStorage.setItem('forma-cart', JSON.stringify(value));
		} catch {}
	});
	const saveFavorites = favorites.subscribe((value) => {
		try {
			localStorage.setItem('forma-favorites', JSON.stringify(value));
		} catch {}
	});
	return () => {
		saveCart();
		saveFavorites();
	};
}
