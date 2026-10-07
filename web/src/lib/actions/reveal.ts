import type { Action } from 'svelte/action';

/** One observer per page; opt in with data-reveal or data-cascade. */
export const reveal: Action<HTMLElement> = (root) => {
	const preference = matchMedia('(prefers-reduced-motion: reduce)');
	const targets = new Map<
		HTMLElement,
		{ delay: number; animation?: Animation; revealed: boolean }
	>();
	const tokens = getComputedStyle(root);
	const duration = Number.parseFloat(tokens.getPropertyValue('--motion-reveal-duration')) || 620;
	const stagger = Number.parseFloat(tokens.getPropertyValue('--motion-stagger')) || 70;
	const easing =
		tokens.getPropertyValue('--motion-ease-out').trim() || 'cubic-bezier(0.22, 1, 0.36, 1)';

	function show(element: HTMLElement, animate = true) {
		const state = targets.get(element);
		if (!state || state.revealed) return;
		state.revealed = true;
		observer?.unobserve(element);
		element.classList.remove('motion-pending');
		if (!animate || preference.matches || !element.animate) return;
		state.animation = element.animate(
			[
				{ opacity: 0, transform: 'translate3d(0, 24px, 0)' },
				{ opacity: 1, transform: 'none' },
			],
			{ duration, delay: state.delay, easing, fill: 'backwards' },
		);
		state.animation.onfinish = () => {
			state.animation = undefined;
		};
	}

	const observer =
		'IntersectionObserver' in window
			? new IntersectionObserver(
					(entries) => {
						for (const entry of entries)
							if (entry.isIntersecting) show(entry.target as HTMLElement);
					},
					{ threshold: 0, rootMargin: '0px 0px -16px 0px' },
				)
			: null;

	function register(element: HTMLElement, delay: number) {
		if (targets.has(element)) return;
		targets.set(element, { delay, revealed: false });
		const rect = element.getBoundingClientRect();
		// Initial content animates immediately; only below-the-fold content waits for scroll.
		// Relative coordinates also work before SvelteKit resets/restores scroll.
		const pageTop = root.getBoundingClientRect().top;
		const inFirstScreen = rect.top - pageTop < innerHeight && rect.bottom - pageTop > 0;
		if (preference.matches || !observer) {
			show(element, false);
			return;
		}
		if (inFirstScreen) {
			// Start now, without waiting for IntersectionObserver or a route snapshot.
			show(element);
			return;
		}
		// Content remains readable without JS or when observers are unsupported.
		element.classList.add('motion-pending');
		observer.observe(element);
	}

	function scan() {
		for (const [element, state] of targets) {
			if (!root.contains(element)) {
				observer?.unobserve(element);
				state.animation?.cancel();
				targets.delete(element);
			}
		}
		root.querySelectorAll<HTMLElement>('[data-reveal]').forEach((element) => register(element, 0));
		root.querySelectorAll<HTMLElement>('[data-cascade]').forEach((group) => {
			Array.from(group.children).forEach((child, index) => {
				if (child instanceof HTMLElement) register(child, Math.min(index * stagger, 280));
			});
		});
	}
	const mutations = new MutationObserver(scan);
	scan();
	mutations.observe(root, { childList: true, subtree: true });
	const reduce = () => {
		if (preference.matches)
			for (const [element, state] of targets) {
				state.animation?.cancel();
				show(element, false);
			}
	};
	const focus = (event: FocusEvent) => {
		if (!(event.target instanceof Node)) return;
		for (const [element, state] of targets)
			if (element.contains(event.target)) {
				state.animation?.cancel();
				show(element, false);
			}
	};
	preference.addEventListener('change', reduce);
	root.addEventListener('focusin', focus);
	return {
		destroy() {
			observer?.disconnect();
			mutations.disconnect();
			preference.removeEventListener('change', reduce);
			root.removeEventListener('focusin', focus);
			for (const [element, state] of targets) {
				state.animation?.cancel();
				element.classList.remove('motion-pending');
			}
			targets.clear();
		},
	};
};
