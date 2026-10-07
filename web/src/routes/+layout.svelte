<script lang="ts">
	import '../app.css';
	import Header from '$lib/components/Header.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import SmoothScroll from '$lib/components/SmoothScroll.svelte';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { reveal } from '$lib/actions/reveal';
	import { hydrate, notice } from '$lib/shop';
	import { Check } from '@lucide/svelte';
	let { children } = $props();
	onMount(() => hydrate(page.data.catalog?.products ?? []));
</script>

<svelte:head
	><meta name="theme-color" content="#ffffff" /><meta
		name="description"
		content="Forma: móveis contemporâneos para sua casa. Explore materiais, dimensões e ambientes."
	/></svelte:head
>
<SmoothScroll />
<a class="skip-link" href="#conteudo">Pular para o conteúdo</a><Header />
<main id="conteudo">
	{#key page.url.pathname}<div class="page-stage" use:reveal>{@render children()}</div>{/key}
</main>
<Footer />
<div class="toast-region" role="status" aria-live="polite">
	{#if $notice}<div class="toast">
			<Check size={18} />{$notice}<a href="/carrinho">Ver sacola</a>
		</div>{/if}
</div>
