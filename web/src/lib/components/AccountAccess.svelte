<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { ArrowRight, Mail, ShieldCheck } from '@lucide/svelte';
	import { commerce } from '$lib/commerce';
	let { kind }: { kind: string } = $props();
	let email = $state(''),
		password = $state(''),
		confirmation = $state(''),
		token = $state(''),
		message = $state(''),
		busy = $state(false),
		done = $state(false);
	const reset = $derived(kind === 'conta/redefinir-senha');
	const verify = $derived(kind === 'conta/confirmar-email');
	onMount(() => {
		token = new URLSearchParams(page.url.hash.slice(1)).get('token') || '';
	});
	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (busy) return;
		if (reset && password !== confirmation) {
			message = 'As senhas precisam ser iguais.';
			return;
		}
		busy = true;
		message = '';
		try {
			if (reset || verify) {
				await commerce(`auth/${verify ? 'verify' : 'reset'}/confirm`, {
					token,
					...(reset ? { password } : {}),
				});
				done = true;
				password = '';
				confirmation = '';
				message = verify
					? 'E-mail confirmado. Sua conta está pronta para comprar.'
					: 'Senha atualizada. Entre novamente com sua nova senha.';
			} else {
				const result = await commerce<{ message: string }>('auth/reset/request', { email });
				message = result.message;
				done = true;
			}
		} catch (err) {
			message = err instanceof Error ? err.message : 'Tente novamente.';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head
	><title
		>{verify ? 'Confirmar e-mail' : reset ? 'Redefinir senha' : 'Recuperar acesso'} — Forma</title
	><meta name="robots" content="noindex,nofollow" /><meta
		name="referrer"
		content="no-referrer"
	/></svelte:head
>
<div class="container page-shell account-page">
	<nav class="breadcrumb" aria-label="Você está em">
		<a href="/">Início</a><span>/</span><a href="/conta">Minha conta</a><span>/</span><span
			>Segurança</span
		>
	</nav>
	<div class="account-intro" data-cascade>
		<p class="account-mark" aria-hidden="true">forma</p>
		<h1>{verify ? 'Um último passo.' : reset ? 'Um novo começo.' : 'Seu acesso de volta.'}</h1>
		<p class="muted">
			{verify
				? 'Confirme seu e-mail para finalizar suas compras.'
				: reset
					? 'Escolha uma nova senha para proteger sua conta.'
					: 'Enviaremos um link para você redefinir sua senha.'}
		</p>
	</div>
	<form class="account-form" data-reveal onsubmit={submit}>
		{#if done}<ShieldCheck size={32} />
			<p role="status">{message}</p>
			<a class="button full-width" href="/conta">Ir para minha conta <ArrowRight size={18} /></a>
		{:else if (reset || verify) && !token}<p role="alert">
				Este link está incompleto. Solicite um novo e-mail.
			</p>
			<a href={verify ? '/conta' : '/conta/recuperar-senha'} class="button-outline"
				>{verify ? 'Ir para minha conta' : 'Solicitar novo link'}</a
			>
		{:else}
			{#if !reset && !verify}<Mail size={28} /><label
					>E-mail<input
						type="email"
						required
						maxlength="254"
						autocomplete="username"
						bind:value={email}
						placeholder="voce@exemplo.com"
					/></label
				>{/if}
			{#if reset}<label
					>Nova senha<input
						type="password"
						required
						minlength="15"
						maxlength="128"
						autocomplete="new-password"
						bind:value={password}
					/></label
				><label
					>Confirme a nova senha<input
						type="password"
						required
						minlength="15"
						maxlength="128"
						autocomplete="new-password"
						bind:value={confirmation}
					/></label
				>
				<p class="small muted">
					Use pelo menos 15 caracteres. Suas outras sessões serão encerradas.
				</p>{/if}
			<button class="button full-width" disabled={busy}
				>{busy
					? 'Aguarde...'
					: verify
						? 'Confirmar meu e-mail'
						: reset
							? 'Salvar nova senha'
							: 'Enviar link de recuperação'}<ArrowRight size={18} /></button
			>
			<p role="status">{message}</p>
			<a class="text-link" href="/conta">Voltar para minha conta</a>
		{/if}
	</form>
</div>
