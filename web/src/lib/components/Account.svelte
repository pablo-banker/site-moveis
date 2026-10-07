<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		UserRound,
		ArrowRight,
		Package,
		Heart,
		MapPin,
		LogOut,
		Pencil,
		Check,
		X,
	} from '@lucide/svelte';
	import { commerce, orderStatus, type Customer, type Order } from '$lib/commerce';
	import { money } from '$lib/catalog';
	import { favorites } from '$lib/shop';
	let sessions = $state<{ current: boolean; expiresAt: string }[]>([]);
	let editingName = $state(false),
		editedName = $state(''),
		profileBusy = $state(false),
		profileMessage = $state('');
	function focusName(node: HTMLInputElement) {
		node.focus();
		node.select();
	}
	function cancelNameEdit() {
		editingName = false;
		profileMessage = '';
	}
	function editName() {
		editedName = customer?.name ?? '';
		profileMessage = '';
		editingName = true;
	}
	async function saveName(event: SubmitEvent) {
		event.preventDefault();
		if (profileBusy) return;
		profileBusy = true;
		profileMessage = '';
		try {
			customer = await commerce<Customer>('auth/profile', { name: editedName.trim() });
			editingName = false;
			profileMessage = 'Nome atualizado.';
		} catch (err) {
			profileMessage =
				err instanceof Error ? err.message : 'Não foi possível salvar seu nome. Tente novamente.';
		} finally {
			profileBusy = false;
		}
	}
	let status = $state('all');
	let ordersError = $state('');
	let { orders = false }: { orders?: boolean } = $props();
	let mode = $state('login'),
		name = $state(''),
		email = $state(''),
		password = $state(''),
		message = $state('');
	let customer = $state<Customer | null>(null),
		list = $state<Order[]>([]),
		busy = $state(false),
		loading = $state(true);
	let filteredOrders = $derived(
		list.filter((order) => status === 'all' || order.status === status),
	);
	let pending = $derived(list.filter((order) => order.status === 'awaiting_payment').length);
	let latestAddress = $derived(list[0]?.address);
	async function load() {
		try {
			customer = await commerce<Customer>('auth/me');
			sessions = (await commerce<{ data: typeof sessions }>('auth/sessions')).data;
			ordersError = '';
			try {
				list = (await commerce<{ data: Order[] }>('orders')).data;
			} catch {
				ordersError = 'Não foi possível carregar seus pedidos. Tente novamente.';
			}
		} catch {
			customer = null;
		} finally {
			loading = false;
		}
	}
	onMount(load);
	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (busy) return;
		busy = true;
		message = '';
		try {
			if (mode === 'signup') {
				const result = await commerce<{ message: string }>('auth/register', {
					name,
					email,
					password,
				});
				password = '';
				mode = 'login';
				message = result.message;
				return;
			}
			const result = await commerce<{ customer: Customer }>(
				`auth/${mode === 'signup' ? 'register' : 'login'}`,
				mode === 'signup' ? { name, email, password } : { email, password },
			);
			customer = result.customer;
			password = '';
			await load();
			if (page.url.searchParams.get('next') === '/checkout') await goto('/checkout');
		} catch (err) {
			message = err instanceof Error ? err.message : 'Tente novamente.';
		} finally {
			busy = false;
		}
	}
	async function logout() {
		try {
			await commerce('auth/logout', {});
			customer = null;
			list = [];
		} catch (err) {
			message = String(err);
		}
	}
	async function cancel(id: string) {
		if (busy) return;
		busy = true;
		try {
			await commerce(`orders/${id}/cancel`, {});
			await load();
			message = 'Pedido cancelado. A reserva de estoque foi liberada.';
		} catch (err) {
			message = err instanceof Error ? err.message : 'Tente novamente.';
		} finally {
			busy = false;
		}
	}
	async function resend() {
		if (!customer || busy) return;
		busy = true;
		try {
			const result = await commerce<{ message: string }>('auth/verify/request', {
				email: customer.email,
			});
			message = result.message;
		} catch (err) {
			message = err instanceof Error ? err.message : 'Tente novamente.';
		} finally {
			busy = false;
		}
	}
	async function revoke() {
		if (busy) return;
		busy = true;
		try {
			await commerce('auth/sessions/revoke', {});
			await load();
			message = 'As outras sessões foram encerradas.';
		} catch (err) {
			message = err instanceof Error ? err.message : 'Tente novamente.';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>{orders ? 'Meus pedidos' : 'Minha conta'} — Forma</title></svelte:head>
<div class="container page-shell account-page" class:account-dashboard={customer}>
	<nav class="breadcrumb" aria-label="Você está em">
		<a href="/">Início</a><span>/</span>{#if orders}<a href="/conta">Minha conta</a><span>/</span
			><span>Meus pedidos</span>{:else}<span>Minha conta</span>{/if}
	</nav>
	{#if loading}<p role="status">Carregando sua conta...</p>
	{:else if customer}
		<div class="page-heading" data-cascade>
			<p class="account-eyebrow">Área do cliente <span>Minha Forma</span></p>
			<h1>{orders ? 'Meus pedidos.' : 'Minha conta.'}</h1>
			<p class="muted">
				{orders
					? 'Acompanhe o status e os detalhes das suas compras.'
					: `Olá, ${customer.name}. Que bom ter você por aqui.`}
			</p>
		</div>
		<div class="account-grid" data-reveal>
			<aside class="account-nav" aria-label="Área do cliente">
				<a href="/conta" aria-current={!orders ? 'page' : undefined}
					><UserRound size={19} /> Minha conta</a
				><a href="/conta/pedidos" aria-current={orders ? 'page' : undefined}
					><Package size={19} /> Meus pedidos</a
				><a href="/favoritos"><Heart size={19} /> Favoritos</a><button
					class="account-logout"
					onclick={logout}><LogOut size={18} /> Sair da conta</button
				>
			</aside>
			<div class="account-content">
				{#if !customer.emailVerified}<section
						class="account-verification"
						aria-labelledby="verify-email-title"
					>
						<h2 id="verify-email-title">Confirme seu e-mail</h2>
						<p>
							Confirme o endereço <span class="account-verification-email">{customer.email}</span>
							antes de finalizar sua compra.
						</p>
						<div class="account-verification-actions">
							<button class="button-outline" onclick={resend} disabled={busy}
								>Reenviar confirmação</button
							><button class="text-link" onclick={load}>Já confirmei meu e-mail</button>
						</div>
					</section>{/if}
				{#if orders}
					<div class="account-section-heading">
						<div>
							<h2>Histórico de pedidos</h2>
							<p class="muted small">Até 100 pedidos mais recentes.</p>
						</div>
						{#if !ordersError}<label class="order-filter"
								>Mostrar<select bind:value={status}
									><option value="all">Todos os pedidos</option><option value="awaiting_payment"
										>Aguardando pagamento</option
									><option value="paid">Pagamento aprovado</option><option value="cancelled"
										>Cancelados</option
									></select
								></label
							>{/if}
					</div>
					{#if ordersError}<div class="notice-panel" role="alert">
							<p>{ordersError}</p>
							<button class="text-link" onclick={load}>Tentar novamente</button>
						</div>
					{:else if filteredOrders.length}{#each filteredOrders as order}<article
								class="customer-order"
							>
								<p class="muted small">
									Pedido {order.id.slice(0, 8)} · {new Date(order.createdAt).toLocaleDateString(
										'pt-BR',
									)}
								</p>
								<h3>{orderStatus(order.status)}</h3>
								{#each order.items as item}<p>
										{item.quantity} × {item.name} · {item.finish}
									</p>{/each}
								<strong>{money(order.totalCents / 100)}</strong>
								<div class="form-actions">
									<a class="button-outline" href={`/pedido-confirmado?id=${order.id}`}
										>{order.status === 'awaiting_payment'
											? 'Ver pedido e pagar'
											: 'Ver detalhes do pedido'}</a
									>{#if order.status === 'awaiting_payment'}<button
											class="text-link"
											disabled={busy}
											onclick={() => cancel(order.id)}>Cancelar pedido</button
										>{/if}
								</div>
							</article>{/each}
					{:else}<div class="empty-state">
							<Package size={34} strokeWidth={1} />
							<h2>{list.length ? 'Nenhum pedido neste status.' : 'Seu histórico começa aqui.'}</h2>
							<p>
								{list.length
									? 'Escolha outro status para encontrar seus pedidos.'
									: 'Os pedidos criados no checkout aparecerão nesta página.'}
							</p>
							{#if list.length}<button class="button-outline" onclick={() => (status = 'all')}
									>Mostrar todos os pedidos</button
								>{:else}<a href="/moveis" class="button">Explorar móveis</a>{/if}
						</div>{/if}
				{:else}
					<div class="account-info-grid">
						<section class="account-profile">
							<div class="account-section-heading">
								<div>
									<p class="account-card-label">01 / Perfil</p>
									<h2>Seus dados</h2>
								</div>
								<UserRound size={24} strokeWidth={1.4} />
							</div>
							<dl class="account-details">
								<div>
									<dt id="profile-name-label">Nome</dt>
									<dd>
										{#if editingName}
											<form class="profile-name-editor" onsubmit={saveName}>
												<input
													id="profile-name"
													onkeydown={(event) => {
														if (event.key === 'Escape' && !profileBusy) {
															event.preventDefault();
															cancelNameEdit();
														}
													}}
													aria-label="Nome completo"
													use:focusName
													bind:value={editedName}
													required
													minlength="2"
													maxlength="100"
													autocomplete="name"
													disabled={profileBusy}
												/>
												<div class="profile-name-actions">
													<button
														type="submit"
														class="profile-name-save"
														disabled={profileBusy || editedName.trim() === customer.name}
														aria-label={profileBusy ? 'Salvando nome' : 'Salvar nome'}
														title="Salvar nome"><Check size={18} strokeWidth={1.8} /></button
													>
													<button
														type="button"
														class="profile-name-cancel"
														disabled={profileBusy}
														onclick={cancelNameEdit}
														aria-label="Cancelar edição"
														title="Cancelar edição"><X size={18} strokeWidth={1.6} /></button
													>
												</div>
											</form>
										{:else}
											<div class="profile-name-display">
												<span>{customer.name}</span>
												<button
													type="button"
													class="profile-edit-name"
													onclick={editName}
													aria-label="Editar nome"
													title="Editar nome"><Pencil size={16} strokeWidth={1.6} /></button
												>
											</div>
										{/if}
									</dd>
								</div>
								<div>
									<dt>E-mail</dt>
									<dd>{customer.email}</dd>
								</div>
							</dl>
							{#if profileMessage}<p class="small profile-name-message" role="status">
									{profileMessage}
								</p>{/if}
						</section>
						<section class="account-address">
							<div class="account-section-heading">
								<div>
									<p class="account-card-label">02 / Entrega</p>
									<h2>Seu endereço</h2>
								</div>
								<MapPin size={24} strokeWidth={1.4} />
							</div>
							{#if latestAddress}<div>
									<p class="small muted">Utilizado no último pedido</p>
									<p>
										{latestAddress.street}, {latestAddress.number}<br />{latestAddress.city} · {latestAddress.state}<br
										/>CEP {latestAddress.cep}
									</p>
									<p class="muted small">Você informa o endereço de entrega a cada checkout.</p>
								</div>{:else}<p class="muted">
									Seu endereço aparecerá aqui após o primeiro pedido.
								</p>{/if}
						</section>
					</div>
					<div class="account-shortcuts">
						<a href="/conta/pedidos"
							><Package size={23} strokeWidth={1.4} />
							<div>
								<h3>Meus pedidos</h3>
								<p class="small muted">
									{ordersError
										? 'Consulte seu histórico'
										: `${list.length} ${list.length === 1 ? 'pedido recente' : 'pedidos recentes'} · ${pending} aguardando pagamento`}
								</p>
							</div>
							<ArrowRight size={18} /></a
						>
						<a href="/favoritos"
							><Heart size={23} strokeWidth={1.4} />
							<div>
								<h3>Minha seleção</h3>
								<p class="small muted">
									{$favorites.length}
									{$favorites.length === 1 ? 'peça salva' : 'peças salvas'} neste navegador
								</p>
							</div>
							<ArrowRight size={18} /></a
						>
					</div>
					<section class="account-profile">
						<div class="account-section-heading">
							<div>
								<p class="account-card-label">03 / Segurança</p>
								<h2>Acesso à sua conta</h2>
							</div>
						</div>
						<p class="muted small">
							{sessions.length}
							{sessions.length === 1 ? 'sessão ativa' : 'sessões ativas'}. Esta sessão permanece
							aberta ao encerrar os outros acessos.
						</p>
						<div class="form-actions">
							<button class="button-outline" disabled={busy || sessions.length < 2} onclick={revoke}
								>Encerrar outras sessões</button
							><a class="text-link" href="/conta/recuperar-senha">Alterar minha senha</a>
						</div>
					</section>
					<section class="account-recent">
						<div class="account-section-heading">
							<h2>Últimos pedidos</h2>
							<a class="text-link small" href="/conta/pedidos"
								>Ver histórico <ArrowRight size={16} /></a
							>
						</div>
						{#if ordersError}<p class="muted" role="alert">{ordersError}</p>
							<button class="text-link" onclick={load}>Tentar novamente</button>
						{:else if list.length}{#each list.slice(0, 2) as order}<a
									class="recent-order"
									href={`/pedido-confirmado?id=${order.id}`}
									><div>
										<strong>Pedido {order.id.slice(0, 8)}</strong>
										<p class="muted small">
											{new Date(order.createdAt).toLocaleDateString('pt-BR')} · {orderStatus(
												order.status,
											)}
										</p>
									</div>
									<span>{money(order.totalCents / 100)} <ArrowRight size={16} /></span></a
								>{/each}
						{:else}<p class="muted">
								Você ainda não criou pedidos. Explore as peças para começar sua seleção.
							</p>
							<a class="text-link" href="/moveis">Explorar móveis <ArrowRight size={16} /></a>{/if}
					</section>
				{/if}
				<p role="status">{message}</p>
			</div>
		</div>
	{:else}
		<div class="account-intro" data-cascade>
			<p class="account-mark" aria-hidden="true">forma</p>
			<h1>Seu espaço.<br />Suas escolhas.</h1>
			<p class="muted">Crie sua conta para acompanhar seus pedidos.</p>
		</div>
		<form class="account-form" data-reveal onsubmit={submit}>
			<div class="account-tabs">
				<button class:current={mode === 'login'} type="button" onclick={() => (mode = 'login')}
					>Entrar</button
				><button class:current={mode === 'signup'} type="button" onclick={() => (mode = 'signup')}
					>Criar conta</button
				>
			</div>
			{#if mode === 'signup'}<label
					>Nome completo<input
						required
						minlength="2"
						maxlength="100"
						bind:value={name}
						autocomplete="name"
						placeholder="Seu nome completo"
					/></label
				>{/if}
			<label
				>E-mail<input
					required
					type="email"
					maxlength="254"
					bind:value={email}
					autocomplete="username"
					placeholder="voce@exemplo.com"
				/></label
			>
			<label
				>Senha<input
					required
					type="password"
					minlength={mode === 'signup' ? 15 : 1}
					maxlength="128"
					bind:value={password}
					autocomplete={mode === 'signup' ? 'new-password' : 'current-password'}
					placeholder="Mínimo de 15 caracteres"
				/></label
			>
			<button class="button full-width" disabled={busy}
				>{busy
					? 'Aguarde...'
					: mode === 'login'
						? 'Entrar na minha conta'
						: 'Criar minha conta'}<ArrowRight size={18} /></button
			>
			{#if mode === 'login'}<a href="/conta/recuperar-senha" class="text-link small"
					>Esqueci minha senha</a
				>{/if}
			<p class="small muted">Sua conta permite acompanhar pedidos e consultar seus dados.</p>
			<p role="status">{message}</p>
		</form>
	{/if}
</div>
