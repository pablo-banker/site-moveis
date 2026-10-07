<script lang="ts">
	import { Heart, ArrowUpRight } from '@lucide/svelte';
	import type { Product } from '$lib/catalog';
	import { money } from '$lib/catalog';
	import { favorites, favorite } from '$lib/shop';
	let { product }: { product: Product } = $props();
</script>

<article class="product-card">
	<div class="product-image">
		<a href={`/moveis/${product.id}`} aria-label={`Ver ${product.name}`}
			><img src={product.image} alt={product.name} loading="lazy" /></a
		>{#if product.label}<span class="product-label">{product.label}</span>{/if}<button
			class:active={$favorites.includes(product.id)}
			class="favorite-button"
			aria-label={`${$favorites.includes(product.id) ? 'Remover' : 'Salvar'} ${product.name} nos favoritos`}
			aria-pressed={$favorites.includes(product.id)}
			onclick={() => favorite(product.id)}
			><Heart size={18} fill={$favorites.includes(product.id) ? 'currentColor' : 'none'} /></button
		>
	</div>
	<div class="product-meta">
		<div>
			<p class="muted small">{product.category}</p>
			<a class="product-name" href={`/moveis/${product.id}`}>{product.name}</a>
		</div>
		<ArrowUpRight size={18} />
	</div>
	<p class="product-price">
		{money(product.price)} <span class="muted small">ou 10× de {money(product.price / 10)}</span>
	</p>
</article>
