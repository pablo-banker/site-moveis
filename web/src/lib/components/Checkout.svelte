<script lang="ts">
	let products = $derived(page.data.catalog?.products ?? []);
	import { page } from '$app/state';
	import { cart, subtotal } from '$lib/shop';
	import { onMount } from 'svelte';
	import { commerce, type Customer, type Order, type ShippingQuote } from '$lib/commerce';
	let customer = $state<Customer | null>(null);
	let loading = $state(true),
		busy = $state(false),
		message = $state(''),
		key = $state('');
	onMount(async () => {
		key = crypto.randomUUID();
		try {
			customer = await commerce<Customer>('auth/me');
			name = customer.name;
			email = customer.email;
		} catch {
		} finally {
			loading = false;
		}
	});
	import { money } from '$lib/catalog';
	import { goto } from '$app/navigation';
	import { ArrowRight, LockKeyhole } from '@lucide/svelte';
	let step = $state(1);
	let name = $state('');
	let email = $state('');
	let cep = $state('');
	let street = $state('');
	let number = $state('');
	let city = $state('');
	let uf = $state('');
	let cepLoading = $state(false);
	let cepMessage = $state('');
	let cepFailed = $state(false);

	async function lookupCep(code: string, signal: AbortSignal) {
		const before = { street, city, uf };
		cepLoading = true;
		cepFailed = false;
		cepMessage = 'Buscando endereço...';
		try {
			const response = await fetch(`/endereco/${code}`, { signal });
			const result = await response.json();
			if (signal.aborted) return;
			if (!response.ok) throw new Error(result.message);
			// Preserve fields edited while the request was in progress.
			if (street === before.street) street = result.street;
			if (city === before.city) city = result.city;
			if (uf === before.uf) uf = result.state;
			cepMessage = result.street
				? 'Endereço encontrado. Confira os dados e informe o número.'
				: 'Cidade e estado encontrados. Informe a rua e o número.';
		} catch (error) {
			if (signal.aborted) return;
			cepFailed = true;
			cepMessage =
				error instanceof Error && error.message !== 'Failed to fetch'
					? error.message
					: 'Não foi possível consultar o CEP. Preencha o endereço manualmente.';
		} finally {
			if (!signal.aborted) cepLoading = false;
		}
	}
	$effect(() => {
		const code = cep.replace(/-/g, '');
		cepLoading = false;
		cepMessage = '';
		cepFailed = false;
		if (!/^\d{8}$/.test(code)) return;
		const controller = new AbortController();
		const timer = setTimeout(() => lookupCep(code, controller.signal), 350);
		return () => {
			clearTimeout(timer);
			controller.abort();
		};
	});

	let shipping = $state<ShippingQuote | null>(null);
	let shippingLoading = $state(false);
	let shippingMessage = $state('');
	let shippingAttempt = $state(0);
	$effect(() => {
		const code = cep.replace(/-/g, '');
		const items = $cart.map((item) => ({ ...item }));
		const user = customer?.id;
		shippingAttempt;
		shipping = null;
		shippingMessage = '';
		shippingLoading = false;
		if (!user || !/^\d{8}$/.test(code) || !items.length) return;
		let active = true;
		let expiryTimer: ReturnType<typeof setTimeout>;
		shippingLoading = true;
		const timer = setTimeout(async () => {
			try {
				const quote = await commerce<ShippingQuote>('shipping/quotes', { cep: code, items });
				if (!active) return;
				shipping = quote;
				expiryTimer = setTimeout(
					() => {
						shipping = null;
						shippingMessage = 'A cotação expirou. Calcule o frete novamente.';
					},
					Math.max(0, Date.parse(quote.expiresAt) - Date.now()),
				);
			} catch (err) {
				if (active)
					shippingMessage =
						err instanceof Error ? err.message : 'Não foi possível calcular o frete.';
			} finally {
				if (active) shippingLoading = false;
			}
		}, 350);
		return () => {
			active = false;
			clearTimeout(timer);
			clearTimeout(expiryTimer);
		};
	});

	$effect(() => {
		// A retry keeps its key; a changed checkout payload starts a new submission.
		JSON.stringify({
			items: $cart,
			address: { cep, street, number, city, state: uf },
			quote: shipping?.id,
		});
		key = crypto.randomUUID();
	});
	async function next(event: SubmitEvent) {
		event.preventDefault();
		if (busy || (step >= 2 && (cepLoading || shippingLoading))) return;
		if (step >= 2 && (!shipping || Date.parse(shipping.expiresAt) <= Date.now())) {
			shippingMessage = 'Calcule o frete antes de continuar.';
			step = 2;
			return;
		}
		if (step < 3) {
			step++;
			return;
		}
		busy = true;
		message = '';
		try {
			const order = await commerce<Order>('orders', {
				idempotencyKey: key,
				shippingQuoteId: shipping!.id,
				items: $cart,
				address: { cep, street, number, city, state: uf },
			});
			cart.set([]);
			await goto(`/pedido-confirmado?id=${order.id}`);
		} catch (err) {
			message = err instanceof Error ? err.message : 'Tente novamente.';
			if (message.includes('frete expirou')) {
				shipping = null;
				step = 2;
				shippingMessage = message;
			}
		} finally {
			busy = false;
		}
	}

	let total = $derived(subtotal($cart, products));
</script>

<svelte:head><title>Checkout — Forma</title></svelte:head>
<div class="container page-shell checkout-page">
	<nav class="breadcrumb" aria-label="Você está em">
		<a href="/carrinho">Sacola</a><span>/</span><span>Checkout</span>
	</nav>
	<div class="page-heading" data-cascade>
		<h1>Quase em casa.</h1>
		<p class="muted">Confira seus dados e escolha onde receber sua seleção.</p>
	</div>
	{#if loading}<p role="status">Carregando sua conta...</p>{:else if !customer}<div
			class="empty-state"
		>
			<h2>Seu pedido começa na sua conta.</h2>
			<p>Entre ou crie uma conta para continuar.</p>
			<a class="button" href="/conta?next=/checkout">Entrar ou criar conta</a>
		</div>{:else if $cart.length}<div class="cart-layout">
			<div>
				<ol class="checkout-steps">
					{#each ['Seus dados', 'Entrega', 'Pagamento'] as label, i}<li
							class:current={step === i + 1}
							class:done={step > i + 1}
						>
							<span>{i + 1}</span>{label}
						</li>{/each}
				</ol>
				<form class="checkout-form" data-reveal onsubmit={next}>
					{#if step === 1}<h2>Como podemos chamar você?</h2>
						<label
							>Nome completo<input
								required
								autocomplete="off"
								readonly
								bind:value={name}
								placeholder="Ex.: Ana Silva"
							/></label
						><label
							>E-mail<input
								type="email"
								required
								autocomplete="off"
								readonly
								bind:value={email}
								placeholder="ana@exemplo.com"
							/></label
						>{:else if step === 2}<h2>Onde o móvel vai morar?</h2>
						<div class="form-two">
							<label
								>CEP<input
									required
									pattern={'[0-9]{5}-?[0-9]{3}'}
									inputmode="numeric"
									maxlength="9"
									autocomplete="postal-code"
									aria-describedby="cep-feedback"
									bind:value={cep}
									placeholder="00000-000"
								/></label
							><label
								>Estado<select required bind:value={uf}
									><option value="" disabled>Selecione</option
									>{#each ['AC', 'AL', 'AP', 'AM', 'BA', 'CE', 'DF', 'ES', 'GO', 'MA', 'MT', 'MS', 'MG', 'PA', 'PB', 'PR', 'PE', 'PI', 'RJ', 'RN', 'RS', 'RO', 'RR', 'SC', 'SP', 'SE', 'TO'] as state}<option
											>{state}</option
										>{/each}</select
								></label
							>
						</div>
						<p
							id="cep-feedback"
							class="cep-feedback small"
							class:cep-error={cepFailed}
							role="status"
							aria-live="polite"
						>
							{cepMessage || 'Digite os 8 dígitos do CEP para buscar o endereço.'}
						</p>
						<label
							>Rua ou avenida<input
								required
								autocomplete="address-line1"
								bind:value={street}
							/></label
						>
						<div class="form-two">
							<label>Número<input required bind:value={number} /></label><label
								>Cidade<input required bind:value={city} /></label
							>
						</div>
						<div class="shipping-choice" aria-live="polite">
							{#if shippingLoading}<p role="status">Calculando frete...</p>
							{:else if shipping}<strong
									>{shipping.service} · {money(shipping.amountCents / 100)}</strong
								>
								<p class="small muted">
									{shipping.minDays} a {shipping.maxDays} dias úteis
								</p>
								<p class="small muted">Cotação válida por até 15 minutos.</p>
							{:else}<p>{shippingMessage || 'Informe o CEP para calcular o frete.'}</p>
								{#if shippingMessage}<button
										class="text-link"
										type="button"
										onclick={() => shippingAttempt++}>Calcular novamente</button
									>{/if}{/if}
						</div>{:else}<h2>Pronto para criar seu pedido?</h2>
						<p>
							Os preços, o estoque e a validade da cotação de frete serão conferidos no servidor.
						</p>
						<p class="notice-panel">
							<LockKeyhole size={19} /> Pagamento processado com segurança.
						</p>
						<p class="muted small">
							Depois de criar o pedido, siga para o pagamento e acompanhe o status na sua conta.
						</p>{/if}
					<div class="form-actions">
						{#if step > 1}<button class="button-outline" type="button" onclick={() => step--}
								>Voltar</button
							>{/if}<button
							class="button"
							disabled={busy || (step >= 2 && (cepLoading || shippingLoading || !shipping))}
							>{busy ? 'Criando pedido...' : step === 3 ? 'Criar pedido' : 'Continuar'}<ArrowRight
								size={18}
							/></button
						>
					</div>
					<p role="status">{message}</p>
				</form>
			</div>
			<aside class="order-summary" data-reveal>
				<h2>Sua seleção</h2>
				{#each $cart as item}{@const product = products.find((p) => p.id === item.id)!}
					<div class="checkout-item">
						<img src={product.image} alt={product.name} /><span
							>{product.name}<small>{item.quantity} × {item.finish}</small></span
						><span>{money(product.price * item.quantity)}</span>
					</div>{/each}
				<div><span>Subtotal</span><span>{money(total)}</span></div>
				<div>
					<span>Frete</span><span
						>{shipping
							? money(shipping.amountCents / 100)
							: shippingLoading
								? 'Calculando...'
								: 'A calcular'}</span
					>
				</div>
				<div class="summary-total">
					<strong>Total</strong><strong
						>{shipping ? money(total + shipping.amountCents / 100) : 'A calcular'}</strong
					>
				</div>
			</aside>
		</div>{:else}<div class="empty-state">
			<h2>Sua sacola está vazia.</h2>
			<p>Escolha um móvel antes de iniciar o checkout.</p>
			<a class="button" href="/moveis">Explorar móveis</a>
		</div>{/if}
</div>
