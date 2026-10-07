<script lang="ts">
	import { Search, Heart, UserRound, ShoppingBag, Menu, X, ArrowRight } from '@lucide/svelte';
	import { Dialog } from 'bits-ui';
	import { page } from '$app/state';
	import { count } from '$lib/shop';
	let searchOpen = $state(false);
	let menuOpen = $state(false);
	let query = $state('');
</script>

<div class="announcement">Uma nova forma de habitar. Conheça nossa coleção.</div>
<header class="site-header">
	<div class="container header-inner">
		<a class="wordmark" href="/" aria-label="Forma, início">forma<span>®</span></a>
		<nav class="desktop-nav" aria-label="Principal">
			<a class:current={page.url.pathname.startsWith('/moveis')} href="/moveis">Móveis</a><a
				class:current={page.url.pathname.startsWith('/ambientes')}
				href="/ambientes">Ambientes</a
			><a class:current={page.url.pathname === '/sobre'} href="/sobre">Nossa história</a>
		</nav>
		<form action="/moveis" class="header-search" role="search" aria-label="Buscar no site">
			<label class="sr-only" for="header-search">Buscar móveis</label>
			<input id="header-search" name="q" placeholder="Buscar móveis" bind:value={query} />
			<button type="submit" aria-label="Enviar busca"><Search size={18} /></button>
		</form>
		<div class="header-actions">
			<Dialog.Root bind:open={searchOpen}
				><Dialog.Trigger class="icon-button compact-search" aria-label="Buscar móveis"
					><Search size={21} /></Dialog.Trigger
				><Dialog.Portal
					><Dialog.Overlay class="dialog-overlay" /><Dialog.Content class="search-dialog"
						><div class="flex-between">
							<Dialog.Title class="dialog-title">Encontre seu próximo móvel</Dialog.Title
							><Dialog.Close class="icon-button" aria-label="Fechar busca"
								><X size={22} /></Dialog.Close
							>
						</div>
						<Dialog.Description class="muted"
							>Busque pelo nome, material ou categoria.</Dialog.Description
						>
						<form action="/moveis" onsubmit={() => (searchOpen = false)} class="search-form">
							<label class="sr-only" for="site-search">Buscar móveis</label><input
								id="site-search"
								name="q"
								bind:value={query}
								placeholder="O que você procura?"
							/><button class="button" aria-label="Buscar"><ArrowRight size={20} /></button>
						</form>
						<p class="small muted">Experimente: sofá, mesa ou carvalho.</p></Dialog.Content
					></Dialog.Portal
				></Dialog.Root
			>
			<a class="icon-button hide-small" href="/favoritos" aria-label="Favoritos"
				><Heart size={21} /></a
			><a class="icon-button hide-small" href="/conta" aria-label="Minha conta"
				><UserRound size={21} /></a
			><a class="icon-button bag" href="/carrinho" aria-label={`Sacola com ${$count} itens`}
				><ShoppingBag size={21} />{#if $count}<span class="bag-count">{$count}</span>{/if}</a
			>
			<Dialog.Root bind:open={menuOpen}
				><Dialog.Trigger class="icon-button mobile-menu" aria-label="Abrir menu"
					><Menu size={22} /></Dialog.Trigger
				><Dialog.Portal
					><Dialog.Overlay class="dialog-overlay" /><Dialog.Content class="menu-dialog"
						><div class="flex-between">
							<Dialog.Title class="wordmark">forma</Dialog.Title><Dialog.Close
								class="icon-button"
								aria-label="Fechar menu"><X /></Dialog.Close
							>
						</div>
						<Dialog.Description class="sr-only">Navegação da loja</Dialog.Description>
						<nav aria-label="Menu móvel">
							{#each [['Móveis', '/moveis'], ['Ambientes', '/ambientes'], ['Nossa história', '/sobre'], ['Favoritos', '/favoritos'], ['Minha conta', '/conta'], ['Ajuda', '/ajuda']] as link}<a
									href={link[1]}
									onclick={() => (menuOpen = false)}>{link[0]}<ArrowRight size={18} /></a
								>{/each}
						</nav></Dialog.Content
					></Dialog.Portal
				></Dialog.Root
			>
		</div>
	</div>
</header>
