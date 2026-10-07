export type Product = {
	id: string;
	name: string;
	category: string;
	room: string;
	price: number;
	image: string;
	dimensions: string;
	material: string;
	label?: string;
	finishes: string[];
};
export const photo = (id: string, width = 1000) =>
	`https://images.unsplash.com/${id}?auto=format&fit=crop&w=${width}&q=85`;
export const hero = photo('photo-1600210492486-724fe5c67fb0', 1800);
export const products: Product[] = [
	{
		id: 'sofa-arco',
		name: 'Sofá Arco',
		category: 'Sofás',
		room: 'Sala de estar',
		price: 4290,
		image: photo('photo-1555041469-a586c61ea9bc'),
		dimensions: '220 × 92 × 78 cm',
		material: 'Linho e madeira maciça',
		label: 'Mais vendido',
		finishes: ['Linho natural', 'Cinza claro', 'Grafite'],
	},
	{
		id: 'poltrona-linha',
		name: 'Poltrona Linha',
		category: 'Poltronas',
		room: 'Sala de estar',
		price: 1890,
		image: photo('photo-1567538096630-e0c55bd6374c'),
		dimensions: '76 × 82 × 80 cm',
		material: 'Tecido bouclé e carvalho',
		label: 'Novo',
		finishes: ['Bouclé claro', 'Cinza claro', 'Grafite'],
	},
	{
		id: 'mesa-encontro',
		name: 'Mesa Encontro',
		category: 'Mesas',
		room: 'Sala de jantar',
		price: 2690,
		image: photo('photo-1533090161767-e6ffed986c88'),
		dimensions: '160 × 90 × 75 cm',
		material: 'Carvalho natural',
		finishes: ['Carvalho natural', 'Nogueira'],
	},
	{
		id: 'cadeira-elo',
		name: 'Cadeira Elo',
		category: 'Cadeiras',
		room: 'Sala de jantar',
		price: 790,
		image: photo('photo-1598300042247-d088f8ab3a91'),
		dimensions: '48 × 52 × 80 cm',
		material: 'Madeira e tecido',
		finishes: ['Carvalho natural', 'Nogueira'],
	},
	{
		id: 'aparador-plano',
		name: 'Aparador Plano',
		category: 'Estantes e aparadores',
		room: 'Sala de estar',
		price: 2190,
		image: photo('photo-1538688423619-a81d3f23454b'),
		dimensions: '140 × 40 × 75 cm',
		material: 'Madeira natural',
		finishes: ['Carvalho natural', 'Nogueira'],
	},
	{
		id: 'mesa-origem',
		name: 'Mesa de centro Origem',
		category: 'Mesas',
		room: 'Sala de estar',
		price: 1190,
		image: photo('photo-1499933374294-4584851497cc'),
		dimensions: '90 × 60 × 35 cm',
		material: 'Carvalho natural',
		label: 'Novo',
		finishes: ['Carvalho natural', 'Nogueira'],
	},
	{
		id: 'cama-refugio',
		name: 'Cama Refúgio',
		category: 'Camas',
		room: 'Quarto',
		price: 3290,
		image: photo('photo-1505693416388-ac5ce068fe85'),
		dimensions: '158 × 198 × 100 cm',
		material: 'Madeira e tecido',
		finishes: ['Linho natural', 'Grafite'],
	},
	{
		id: 'escrivaninha-traco',
		name: 'Escrivaninha Traço',
		category: 'Mesas',
		room: 'Escritório',
		price: 1490,
		image: photo('photo-1518455027359-f3f8164ba6bd'),
		dimensions: '120 × 60 × 75 cm',
		material: 'Madeira e aço',
		finishes: ['Carvalho natural', 'Nogueira'],
	},
];
export const categories = [
	'Todos',
	'Sofás',
	'Poltronas',
	'Mesas',
	'Cadeiras',
	'Estantes e aparadores',
	'Camas',
];
export const rooms = ['Sala de estar', 'Sala de jantar', 'Quarto', 'Escritório'];
export const money = (value: number) =>
	new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value);
export const slug = (value: string) =>
	value
		.normalize('NFD')
		.replace(/[\u0300-\u036f]/g, '')
		.toLowerCase()
		.replaceAll(' ', '-');
