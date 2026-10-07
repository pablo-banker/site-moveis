<script lang="ts">
	import { onMount } from 'svelte';
	import { beforeNavigate, afterNavigate } from '$app/navigation';
	import Lenis from 'lenis';
	import 'lenis/dist/lenis.css';

	let scrolling: Lenis | undefined;
	let frame = 0;

	// Cancel momentum before routing so it cannot move the next page after
	// SvelteKit has reset or restored the native scroll position.
	beforeNavigate(() => scrolling?.scrollTo(window.scrollY, { immediate: true, force: true }));
	afterNavigate(() => {
		cancelAnimationFrame(frame);
		frame = requestAnimationFrame(() => {
			scrolling?.resize();
			scrolling?.scrollTo(window.scrollY, { immediate: true, force: true });
		});
	});

	onMount(() => {
		const reduced = matchMedia('(prefers-reduced-motion: reduce)');
		const finePointer = matchMedia('(pointer: fine)');
		function configure() {
			scrolling?.destroy();
			scrolling = undefined;
			if (reduced.matches || !finePointer.matches) return;
			scrolling = new Lenis({
				autoRaf: true,
				lerp: 0.09,
				smoothWheel: true,
				syncTouch: false,
				anchors: { offset: -24 },
				autoToggle: false,
				allowNestedScroll: true,
				stopInertiaOnNavigate: true,
				respectReducedMotion: true,
			});

			synchronizeLock();
		}
		function synchronizeLock() {
			if (getComputedStyle(document.body).overflow === 'hidden') scrolling?.stop();
			else scrolling?.start();
		}
		const locks = new MutationObserver(synchronizeLock);
		locks.observe(document.body, {
			attributes: true,
			attributeFilter: ['style', 'data-scroll-locked'],
		});
		configure();
		reduced.addEventListener('change', configure);
		finePointer.addEventListener('change', configure);
		return () => {
			cancelAnimationFrame(frame);
			locks.disconnect();
			reduced.removeEventListener('change', configure);
			finePointer.removeEventListener('change', configure);
			scrolling?.destroy();
			scrolling = undefined;
		};
	});
</script>
