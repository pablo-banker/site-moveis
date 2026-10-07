import type { Product } from '$lib/catalog';
declare global {
	namespace App {
		interface PageData {
			catalog?: { products: Product[]; categories: string[]; rooms: string[]; source: string };
		}
	}
}
export {};
