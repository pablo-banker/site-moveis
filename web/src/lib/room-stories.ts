import { photo, slug } from './catalog';

export const roomStories = [
	{
		name: 'Sala de estar',
		title: 'Um lugar para desacelerar.',
		image: photo('photo-1600210492486-724fe5c67fb0', 1600),
		intro:
			'Conversas que se estendem, um livro no fim da tarde, uma pausa só sua. A sala começa com o jeito que você quer viver nela.',
		detail:
			'Comece pelo sofá, aproxime uma poltrona e deixe uma mesa ao alcance. Madeira e tecidos claros ajudam a compor uma base tranquila, que ganha personalidade com os seus objetos.',
		tips: [
			'Deixe espaço para circular entre as peças.',
			'Use uma poltrona para criar um canto de leitura.',
			'Compare as medidas do sofá com a área disponível.',
		],
	},
	{
		name: 'Sala de jantar',
		title: 'Espaço para reunir.',
		image: photo('photo-1617806118233-18e1de247200', 1600),
		intro:
			'Do café rápido aos encontros sem hora para terminar. A mesa é o ponto de partida para os momentos compartilhados.',
		detail:
			'Escolha a mesa pensando na rotina e no espaço ao redor. As cadeiras completam a composição, enquanto a madeira aproxima o ambiente de uma atmosfera acolhedora.',
		tips: [
			'Reserve espaço para afastar as cadeiras.',
			'Considere quantas pessoas usam a mesa no dia a dia.',
			'Combine acabamentos sem precisar repetir tudo.',
		],
	},
	{
		name: 'Quarto',
		title: 'O seu tempo de descansar.',
		image: photo('photo-1505693416388-ac5ce068fe85', 1600),
		intro:
			'Um ambiente para diminuir o ritmo. Menos elementos, escolhas cuidadosas e espaço para descansar.',
		detail:
			'A cama organiza o quarto. Pense nas proporções antes de escolher e mantenha o caminho ao redor livre. Madeira e tecido ajudam a construir uma composição simples.',
		tips: [
			'Confira a medida da cama e do colchão.',
			'Mantenha os caminhos de circulação livres.',
			'Escolha acabamentos que conversem com a roupa de cama.',
		],
	},
	{
		name: 'Escritório',
		title: 'Um lugar para suas ideias.',
		image: photo('photo-1497366754035-f200968a6e72', 1600),
		intro:
			'Uma superfície livre, luz por perto e espaço para se concentrar. O trabalho também pode encontrar sua forma dentro de casa.',
		detail:
			'Comece pela área de trabalho e pelos objetos que precisam estar ao alcance. Uma escrivaninha de proporções adequadas ajuda a organizar o ambiente sem ocupar mais espaço do que o necessário.',
		tips: [
			'Planeje a posição da mesa em relação à luz.',
			'Confira a profundidade para seus equipamentos.',
			'Reserve espaço para a cadeira e para circulação.',
		],
	},
];
export const roomStory = (id: string) => roomStories.find((room) => slug(room.name) === id)!;
