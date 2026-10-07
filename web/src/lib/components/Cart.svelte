<script lang="ts">
	let products = $derived(page.data.catalog?.products ?? []);
	import { page } from '$app/state';
	import { Minus, Plus, X, ArrowRight, ShoppingBag } from '@lucide/svelte';
	import { cart, subtotal, change } from '$lib/shop';
	import { money } from '$lib/catalog';
	let total = $derived(subtotal($cart, products));
</script>

<svelte:head><title>Sua sacola — Forma</title></svelte:head>
<div class="container page-shell cart-page">
	<nav class="breadcrumb" aria-label="Você está em">
		<a href="/">Início</a><span>/</span><span>Sacola</span>
	</nav>
	<div class="page-heading" data-cascade>
		<h1>Sua sacola.</h1>
		<p class="muted">Um novo capítulo para a sua casa.</p>
	</div>
	{#if $cart.length}<div class="cart-layout">
			<div>
				{#each $cart as item}{@const product = products.find((p) => p.id === item.id)!}
					<article class="cart-item" data-reveal>
						<a href={`/moveis/${item.id}`}><img src={product.image} alt={product.name} /></a>
						<div class="cart-item-info">
							<a class="product-name" href={`/moveis/${item.id}`}>{product.name}</a>
							<p class="small muted">{item.finish}</p>
							<p>{money(product.price)}</p>
							<div class="quantity">
								<button
									aria-label={`Diminuir quantidade de ${product.name}`}
									onclick={() => change(item.id, item.finish, item.quantity - 1)}
									><Minus size={14} /></button
								><span>{item.quantity}</span><button
									aria-label={`Aumentar quantidade de ${product.name}`}
									disabled={item.quantity >= 10}
									onclick={() => change(item.id, item.finish, item.quantity + 1)}
									><Plus size={14} /></button
								>
							</div>
						</div>
						<div class="cart-item-end">
							<button
								class="icon-button"
								aria-label={`Remover ${product.name}`}
								onclick={() => change(item.id, item.finish, 0)}><X size={18} /></button
							><strong>{money(product.price * item.quantity)}</strong>
						</div>
					</article>{/each}<a class="text-link continue-link" href="/moveis"
					>Continuar explorando <ArrowRight size={17} /></a
				>
			</div>
			<aside class="order-summary" data-reveal>
				<h2>Resumo da sacola</h2>
				<div><span>Subtotal</span><span>{money(total)}</span></div>
				<div><span>Entrega</span><span class="muted">No próximo passo</span></div>
				<div class="summary-total">
					<strong>Total dos produtos</strong><strong>{money(total)}</strong>
				</div>
				<p class="small muted">Até 10× de {money(total / 10)} sem juros</p>
				<a class="button full-width" href="/checkout"
					>Continuar para checkout <ArrowRight size={17} /></a
				>
				<p class="small muted">Frete e prazo de entrega serão calculados no checkout.</p>
			</aside>
		</div>{:else}<div class="empty-state">
			<ShoppingBag size={40} strokeWidth={1} />
			<h2>Há espaço para novas escolhas.</h2>
			<p>Sua sacola está vazia. Explore móveis para o seu dia a dia.</p>
			<a class="button" href="/moveis">Explorar móveis <ArrowRight size={17} /></a>
		</div>{/if}
</div>
