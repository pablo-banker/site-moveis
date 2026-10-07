export type ParallaxLayer = {
	src: string;
	depth: number;
	lateral?: number;
	position?: string;
	scale?: number;
	x?: number;
	y?: number;
	clip?: string;
};
// All PNGs share a 1672 × 941 canvas. Registration compensates for generated cutout framing.
export const roomLayers: ParallaxLayer[] = [
	{ src: '/background/room-background.png', depth: 0.07, lateral: 0 },
	{
		src: '/background/room-rear.png',
		depth: 0.055,
		lateral: 0.025,
		scale: 0.66,
		x: 27.2,
		y: 11,
		clip: 'inset(0 0 0 60%)',
	},
	{
		src: '/background/room-rear.png',
		depth: 0.035,
		lateral: -0.12,
		scale: 0.65,
		x: 18.8,
		y: 12.6,
		clip: 'inset(0 60% 0 0)',
	},
	{ src: '/background/room-sofa.png', depth: 0.04, lateral: -0.07, scale: 0.714, x: 15.6, y: 13.3 },
	{
		src: '/background/room-rug-table.png',
		depth: 0.025,
		lateral: 0.045,
		scale: 0.82,
		x: 11.7,
		y: 13.1,
	},
	{ src: '/background/room-chair.png', depth: 0.015, lateral: -0.11, scale: 0.89, x: 10.9, y: 7 },
];
export const backgroundStudy = roomLayers;
