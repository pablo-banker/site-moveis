import type { Action } from 'svelte/action';

/** Scroll progress shared by editorial image and typography effects. */
export const editorial: Action<HTMLElement> = (node) => {
	const reduced = matchMedia('(prefers-reduced-motion: reduce)');
	let frame = 0;
	let visible = false;
	const update = () => {
		frame = 0;
		const rect = node.getBoundingClientRect();
		const progress = reduced.matches
			? 0.5
			: Math.max(0, Math.min(1, (innerHeight - rect.top) / (innerHeight + rect.height)));
		node.style.setProperty('--story-progress', String(progress));
	};
	const request = () => {
		if (visible && !frame) frame = requestAnimationFrame(update);
	};
	const observer = new IntersectionObserver(([entry]) => {
		visible = entry.isIntersecting;
		request();
	});
	observer.observe(node);
	update();
	window.addEventListener('scroll', request, { passive: true });
	window.addEventListener('resize', request);
	reduced.addEventListener('change', update);
	return {
		destroy() {
			observer.disconnect();
			cancelAnimationFrame(frame);
			window.removeEventListener('scroll', request);
			window.removeEventListener('resize', request);
			reduced.removeEventListener('change', update);
		},
	};
};
