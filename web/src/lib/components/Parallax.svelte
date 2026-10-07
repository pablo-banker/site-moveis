<script lang="ts">
	import { onMount } from 'svelte';
	import { roomLayers, type ParallaxLayer } from '$lib/parallax';
	let { layers = roomLayers }: { layers?: ParallaxLayer[] } = $props();
	let element: HTMLDivElement;
	let offset = $state(0);
	let width = $state(1672);
	let lateralFactor = $state(1);
	let lateralLimit = $state(64);
	let height = $derived((width * 941) / 1672);
	onMount(() => {
		const motion = window.matchMedia('(prefers-reduced-motion: reduce)');
		let frame = 0;
		let visible = false;
		let initialTop: number | undefined;
		const update = () => {
			frame = 0;
			const rect = element.getBoundingClientRect();
			width = Math.max(rect.width, (rect.height * 1672) / 941) * 1.08;
			lateralFactor = Math.min(1, rect.width / 900);
			lateralLimit = Math.min(64, rect.width * 0.08);
			// Initial viewport sections start assembled. Later sections begin moving
			// as their top crosses the viewport, keeping uncovered edges off screen.
			initialTop ??= rect.top + window.scrollY < window.innerHeight ? rect.top + window.scrollY : 0;
			offset = motion.matches ? 0 : Math.max(0, Math.min(rect.height, initialTop - rect.top));
		};
		const request = () => {
			if (visible && !frame) frame = requestAnimationFrame(update);
		};
		const observer = new IntersectionObserver((entries) => {
			visible = entries[0].isIntersecting;
			if (visible) request();
		});
		observer.observe(element);
		const resize = new ResizeObserver(() => {
			if (!frame) frame = requestAnimationFrame(update);
		});
		resize.observe(element);
		const changeMotion = () => {
			if (motion.matches) offset = 0;
			else request();
		};
		window.addEventListener('scroll', request, { passive: true });
		motion.addEventListener('change', changeMotion);
		return () => {
			observer.disconnect();
			resize.disconnect();
			cancelAnimationFrame(frame);
			window.removeEventListener('scroll', request);
			motion.removeEventListener('change', changeMotion);
		};
	});
</script>

<div bind:this={element} class="parallax-background" aria-hidden="true">
	<div class="parallax-scene" style:width={`${width}px`} style:height={`${height}px`}>
		{#each layers as layer, index}
			<div
				class="parallax-plane"
				data-depth={layer.depth}
				data-lateral={layer.lateral || 0}
				style:transform={`translate3d(${Math.max(-lateralLimit, Math.min(lateralLimit, offset * (layer.lateral || 0) * lateralFactor))}px, ${offset * layer.depth}px, 0)`}
			>
				<img
					src={layer.src}
					alt=""
					width="1672"
					height="941"
					fetchpriority={index === 0 ? 'high' : 'auto'}
					decoding="async"
					style:clip-path={layer.clip || 'none'}
					style:transform={`translate(${layer.x || 0}%, ${layer.y || 0}%) scale(${layer.scale || 1})`}
				/>
			</div>
		{/each}
	</div>
</div>
