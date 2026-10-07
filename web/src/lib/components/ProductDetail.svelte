<script lang="ts">
	let products = $derived(page.data.catalog?.products ?? []);
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { commerce, type ShippingQuote } from '$lib/commerce';
	let variants = $state<{ finish: string; stock: number }[]>([]),
		stockMessage = $state('Consultando estoque...');
	let available = $derived(variants.find((v) => v.finish === finish)?.stock ?? 0);
	onMount(() => {
		void commerce<{ data: { finish: string; stock: number }[] }>(`products/${id}/variants`)
			.then((r) => {
				variants = r.data;
				stockMessage = 'Acabamento indisponível no estoque.';
			})
			.catch(() => {
				stockMessage = 'Não foi possível consultar a disponibilidade.';
			});
	});
	import { ArrowRight, Heart, Minus, Plus, Truck, Check, ChevronDown } from '@lucide/svelte';
	import { money, slug } from '$lib/catalog';
	import { productStories } from '$lib/product-stories';
	import { add, favorite, favorites } from '$lib/shop';
	import { editorial } from '$lib/actions/editorial';
	import ProductCard from './ProductCard.svelte';
	let { id }: { id: string } = $props();
	let product = $derived(products.find((p) => p.id === id)!);
	let finish = $state('');
	let quantity = $state(1);
	let cep = $state('');
	let delivery = $state('');
	$effect(() => {
		finish = product.finishes[0];
		quantity = 1;
		delivery = '';
	});

	let deliveryLoading = $state(false);
	async function shipping(event: SubmitEvent) {
		event.preventDefault();
		if (deliveryLoading) return;
		const code = cep.replace(/-/g, '');
		if (!/^\d{8}$/.test(code)) {
			delivery = 'Digite um CEP válido com 8 números.';
			return;
		}
		deliveryLoading = true;
		delivery = 'Calculando frete...';
		try {
			const quote = await commerce<ShippingQuote>('shipping/quotes', {
				cep: code,
				items: [{ id, finish, quantity }],
			});
			delivery = `${quote.service}: ${money(quote.amountCents / 100)} · ${quote.minDays} a ${quote.maxDays} dias úteis.`;
		} catch (err) {
			delivery = err instanceof Error ? err.message : 'Não foi possível calcular o frete.';
		} finally {
			deliveryLoading = false;
		}
	}
</script>

<svelte:head><title>{product.name} — Forma</title></svelte:head>
<div class="container page-shell product-page">
	<nav class="breadcrumb" aria-label="Você está em">
		<a href="/">Início</a><span>/</span><a href="/moveis">Móveis</a><span>/</span><span
			>{product.name}</span
		>
	</nav>
	<div class="product-detail">
		<div class="detail-gallery" data-reveal use:editorial>
			<img src={product.image} alt={product.name} fetchpriority="high" />
			<div class="gallery-note">Imagem de referência · Confira os detalhes da peça</div>
		</div>
		<div class="detail-info" data-reveal>
			<p class="muted">{product.category}</p>
			<h1>{product.name}</h1>
			<p class="detail-description">
				{productStories[id].description}
			</p>
			<p class="detail-price">{money(product.price)}</p>
			<p class="muted small">Em até 10× de {money(product.price / 10)} sem juros</p>
			<div class="detail-divider"></div>
			<fieldset class="finish-options">
				<legend>Acabamento <strong>{finish}</strong></legend
				>{#each product.finishes as option}<button
						class:chosen={finish === option}
						type="button"
						onclick={() => (finish = option)}
						aria-pressed={finish === option}>{option}</button
					>{/each}
			</fieldset>
			<div class="purchase-row">
				<div class="quantity">
					<button
						aria-label="Diminuir quantidade"
						disabled={quantity <= 1}
						onclick={() => quantity--}><Minus size={15} /></button
					><span>{quantity}</span><button
						aria-label="Aumentar quantidade"
						disabled={quantity >= Math.min(10, available)}
						onclick={() => quantity++}><Plus size={15} /></button
					>
				</div>
				<button
					class="button"
					disabled={available < quantity}
					onclick={() => add(product.id, finish, quantity)}
					>Adicionar à sacola <ArrowRight size={18} /></button
				><button
					class="icon-button favorite-detail"
					aria-label="Salvar nos favoritos"
					aria-pressed={$favorites.includes(id)}
					onclick={() => favorite(id)}
					><Heart size={21} fill={$favorites.includes(id) ? 'currentColor' : 'none'} /></button
				>
			</div>
			<p class="small muted">
				{available > 0 ? `${available} unidades disponíveis · estoque` : stockMessage}
			</p>
			<form class="shipping-form" onsubmit={shipping}>
				<label for="cep"><Truck size={18} /> Calcule a entrega</label>
				<div>
					<input
						id="cep"
						inputmode="numeric"
						maxlength="9"
						placeholder="Seu CEP"
						bind:value={cep}
					/><button class="button-outline">Calcular</button>
				</div>
				<p class="small" role="status">{delivery}</p>
			</form>
			<details open>
				<summary>Dimensões e materiais <ChevronDown size={17} /></summary>
				<dl>
					<div>
						<dt>Largura × profundidade × altura</dt>
						<dd>{product.dimensions}</dd>
					</div>
					<div>
						<dt>Material</dt>
						<dd>{product.material}</dd>
					</div>
				</dl>
			</details>
			<details>
				<summary>Cuidados com a peça <ChevronDown size={17} /></summary>
				<p class="muted">
					Use pano macio e seco. Evite produtos abrasivos, umidade excessiva e exposição direta ao
					sol. Consulte as instruções de cuidado da peça.
				</p>
			</details>
			<details>
				<summary>Entrega e montagem <ChevronDown size={17} /></summary>
				<p class="muted">
					As condições reais de entrega e montagem serão exibidas após a integração com a operação
					da loja.
				</p>
				<a href="/ajuda" class="text-link">Ver dúvidas frequentes</a>
			</details>
		</div>
	</div>
	<section class="product-story collection-story" data-cascade>
		<div>
			<p class="muted">A peça no seu cotidiano</p>
			<h2>{productStories[id].title}</h2>
		</div>
		<div>
			<p>{productStories[id].detail}</p>
			<a class="text-link" href={`/ambientes/${slug(product.room)}`}
				>Explore o ambiente <ArrowRight size={17} /></a
			>
		</div>
	</section>
	{#if products.some((p) => p.id !== id && p.room === product.room)}
		<section class="related">
			<div class="section-heading" data-cascade>
				<h2>Ficam bem juntos.</h2>
				<a href={`/ambientes/${slug(product.room)}#pecas`} class="text-link"
					>Ver peças do ambiente <ArrowRight size={18} /></a
				>
			</div>
			<div class="product-grid" data-cascade>
				{#each products
					.filter((p) => p.id !== id && p.room === product.room)
					.slice(0, 4) as item}<ProductCard product={item} />{/each}
			</div>
		</section>
	{/if}
</div>
