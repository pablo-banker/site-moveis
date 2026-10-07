<script lang="ts">
	let products = $derived(page.data.catalog?.products ?? []);
	import { page } from '$app/state';
	import { Search, SlidersHorizontal, X } from '@lucide/svelte';
	import { slug } from '$lib/catalog';
	let categories = $derived(page.data.catalog?.categories ?? ['Todos']);
	let rooms = $derived(page.data.catalog?.rooms ?? []);
	import { favorites } from '$lib/shop';
	import ProductCard from './ProductCard.svelte';
	let {
		favoritesOnly = false,
		embedded = false,
		room = '',
		categorySlug = '',
	}: {
		favoritesOnly?: boolean;
		embedded?: boolean;
		room?: string;
		categorySlug?: string;
	} = $props();
	let search = $state('');
	let category = $state('Todos');
	let selectedRoom = $state('Todos');
	let sort = $state('relevancia');
	let filtersOpen = $state(false);
	$effect(() => {
		search = page.url.searchParams.get('q') || '';
		selectedRoom = rooms.find((r) => slug(r) === page.url.searchParams.get('ambiente')) || 'Todos';
		category =
			categories.find(
				(c) => c === page.url.searchParams.get('categoria') || slug(c) === categorySlug,
			) || 'Todos';
	});
	let filtered = $derived(
		products
			.filter(
				(p) =>
					(!favoritesOnly || $favorites.includes(p.id)) &&
					(category === 'Todos' || p.category === category) &&
					(!room || slug(p.room) === room) &&
					(selectedRoom === 'Todos' || p.room === selectedRoom) &&
					`${p.name} ${p.material} ${p.category}`
						.toLocaleLowerCase('pt-BR')
						.includes(search.toLocaleLowerCase('pt-BR')),
			)
			.sort((a, b) =>
				sort === 'menor' ? a.price - b.price : sort === 'maior' ? b.price - a.price : 0,
			),
	);
	function reset() {
		search = '';
		category = 'Todos';
		selectedRoom = 'Todos';
	}
</script>

<svelte:head
	><title
		>{favoritesOnly ? 'Favoritos' : room ? rooms.find((r) => slug(r) === room) : 'Todos os móveis'} —
		Forma</title
	></svelte:head
>
<div class="container page-shell catalog-page" class:embedded-catalog={embedded}>
	{#if !embedded}
		<nav class="breadcrumb" aria-label="Você está em">
			<a href="/">Início</a><span>/</span><span>{favoritesOnly ? 'Favoritos' : 'Móveis'}</span>
		</nav>
		<div class="page-heading catalog-heading" data-cascade>
			<p class="muted">{favoritesOnly ? 'Sua seleção pessoal' : 'Design para o cotidiano'}</p>
			<h1>
				{favoritesOnly
					? 'Seus favoritos.'
					: room
						? `${rooms.find((r) => slug(r) === room) || 'Ambiente'}.`
						: 'Móveis para viver.'}
			</h1>
			<p class="muted">
				{favoritesOnly
					? 'Guarde as peças que combinam com a sua casa.'
					: 'Escolha com calma. Encontre o que faz sentido para você.'}
			</p>
		</div>
	{/if}
	<div class="catalog-toolbar">
		<div class="catalog-search">
			<Search size={18} /><label class="sr-only" for="catalog-search">Buscar no catálogo</label
			><input id="catalog-search" placeholder="Buscar um móvel" bind:value={search} />
		</div>
		<button
			class="button-outline filter-toggle"
			onclick={() => (filtersOpen = !filtersOpen)}
			aria-expanded={filtersOpen}><SlidersHorizontal size={16} />Filtros</button
		><label class="sort-label"
			>Ordenar por <select bind:value={sort}
				><option value="relevancia">Relevância</option><option value="menor">Menor preço</option
				><option value="maior">Maior preço</option></select
			></label
		>
	</div>
	<div class="catalog-layout">
		<aside class:filters-open={filtersOpen} class="catalog-filters">
			<div class="flex-between">
				<h2>Filtros</h2>
				<button class="small muted" onclick={reset}>Limpar</button>
			</div>
			<fieldset>
				<legend>Categoria</legend>{#each categories as option}<label class="radio-label"
						><input
							type="radio"
							name="category"
							value={option}
							bind:group={category}
						/>{option}</label
					>{/each}
			</fieldset>
			{#if !room}<fieldset>
					<legend>Ambiente</legend><label class="radio-label"
						><input type="radio" name="room" value="Todos" bind:group={selectedRoom} />Todos</label
					>{#each rooms as option}<label class="radio-label"
							><input
								type="radio"
								name="room"
								value={option}
								bind:group={selectedRoom}
							/>{option}</label
						>{/each}
				</fieldset>{/if}
		</aside>
		<div>
			<p class="results-count muted small">
				{filtered.length}
				{filtered.length === 1 ? 'móvel encontrado' : 'móveis encontrados'}
			</p>
			{#if filtered.length}<div class="product-grid catalog-grid" data-cascade>
					{#each filtered as product}<ProductCard {product} />{/each}
				</div>{:else}<div class="empty-state">
					<h2>{favoritesOnly ? 'Sua seleção começa aqui.' : 'Nenhum móvel encontrado.'}</h2>
					<p>
						{favoritesOnly
							? 'Toque no coração de um produto para salvá-lo.'
							: 'Tente outro termo ou remova os filtros.'}
					</p>
					{#if favoritesOnly}<a class="button" href="/moveis">Explorar móveis</a>{:else}<button
							class="button"
							onclick={reset}>Limpar filtros</button
						>{/if}
				</div>{/if}
		</div>
	</div>
</div>
