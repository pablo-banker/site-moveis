<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { Check, ArrowRight } from '@lucide/svelte';
	import { money } from '$lib/catalog';
	import { commerce, orderStatus, type Order } from '$lib/commerce';
	let order = $state<Order | null>(null),
		message = $state(''),
		busy = $state(false),
		loading = $state(true);
	onMount(async () => {
		const id = page.url.searchParams.get('id');
		try {
			if (id) order = await commerce<Order>(`orders/${id}`);
		} catch (err) {
			message = err instanceof Error ? err.message : 'Tente novamente.';
		} finally {
			loading = false;
		}
	});
	async function pay() {
		if (!order || busy) return;
		busy = true;
		try {
			order = await commerce<Order>(`orders/${order.id}/payment`, {});
			message =
				order.status === 'paid'
					? 'Pagamento aprovado. Seu pedido está confirmado.'
					: 'Pagamento não aprovado. Tente novamente ou cancele o pedido na sua conta.';
		} catch (err) {
			message = err instanceof Error ? err.message : 'Tente novamente.';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>Seu pedido — Forma</title></svelte:head>
<div class="container page-shell confirmation-page">
	<div class="confirmation" data-cascade>
		<span class="confirmation-icon"><Check size={34} /></span>
		<p class="muted">Seu pedido</p>
		<h1>
			{loading
				? 'Carregando seu pedido.'
				: order
					? 'Uma nova forma de morar.'
					: 'Sua próxima escolha começa aqui.'}
		</h1>
		{#if order}<p>Pedido {order.id.slice(0, 8)} · {orderStatus(order.status)}</p>
			<div class="confirmation-box">
				<p class="small muted">
					Subtotal: {money(order.subtotalCents / 100)} · Frete: {money(order.shippingCents / 100)}
				</p>
				{#if order.shipping}<p class="small muted">
						{order.shipping.service} · {order.shipping.minDays} a {order.shipping.maxDays} dias úteis
					</p>{/if}
				<span>Total do pedido</span><strong>{money(order.totalCents / 100)}</strong
				>{#each order.items as item}<p class="small muted">
						{item.quantity} × {item.name} · {item.finish}
					</p>{/each}
			</div>
			{#if order.status === 'awaiting_payment'}<div class="form-actions">
					<button class="button button-white" disabled={busy} onclick={pay}
						>{busy ? 'Processando pagamento...' : 'Pagar pedido'}<ArrowRight size={18} /></button
					>
				</div>{/if}<a class="text-link" href="/conta/pedidos"
				>Acompanhar meus pedidos <ArrowRight size={18} /></a
			>{:else if !loading}<a class="button" href="/moveis"
				>Explorar móveis <ArrowRight size={18} /></a
			>{/if}
		<p role="status">{message}</p>
	</div>
</div>
