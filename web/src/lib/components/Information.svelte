<script lang="ts">
	import { ArrowUpRight, Check, ChevronDown } from '@lucide/svelte';
	import { editorial } from '$lib/actions/editorial';
	import Parallax from './Parallax.svelte';
	import { commerce } from '$lib/commerce';
	import { photo } from '$lib/catalog';
	let { kind }: { kind: string } = $props();
	let sent = $state(false);
	let busy = $state(false),
		message = $state('');
	let contactName = $state(''),
		contactEmail = $state(''),
		subject = $state('Dúvida sobre um produto'),
		body = $state('');
	async function send(event: SubmitEvent) {
		event.preventDefault();
		if (busy) return;
		busy = true;
		sent = false;
		message = '';
		try {
			await commerce('contact', { name: contactName, email: contactEmail, subject, body });
			sent = true;
			body = '';
		} catch (err) {
			message = err instanceof Error ? err.message : 'Não foi possível enviar a mensagem.';
		} finally {
			busy = false;
		}
	}

	const faqs = [
		[
			'Como funciona a entrega?',
			'O frete é calculado no servidor pelo CEP e pelos itens da sacola. O checkout exibe o valor e o prazo da cotação.',
		],
		[
			'Os móveis precisam de montagem?',
			'Essa informação será definida para cada produto e exibida na página da peça antes da compra.',
		],
		[
			'Posso escolher outro acabamento?',
			'Sim, você pode explorar as opções disponíveis na página do produto. Selecione o acabamento antes de adicionar à sacola.',
		],
		[
			'Como cuidar dos móveis?',
			'Utilize pano macio e evite abrasivos. As instruções específicas de cada material serão fornecidas com os produtos reais.',
		],
		[
			'Como funcionam trocas e devoluções?',
			'Consulte nossa equipe para solicitar uma troca ou devolução e informar o número do pedido.',
		],
	];
	let title = $derived(
		kind === 'sobre'
			? 'Nossa história'
			: kind === 'contato'
				? 'Fale com a gente'
				: kind === 'ajuda'
					? 'Podemos ajudar'
					: kind === 'privacidade'
						? 'Privacidade'
						: 'Termos de uso',
	);
</script>

<svelte:head><title>{title} — Forma</title></svelte:head>
{#if kind === 'sobre'}<section class="manifesto">
		<Parallax />
		<div class="container manifesto-content" data-cascade>
			<p>Uma ideia de morar</p>
			<h1>O essencial<br />tem lugar.</h1>
			<p class="manifesto-text">
				Acreditamos em espaços que deixam a vida acontecer. Em formas simples e escolhas feitas com
				tempo.
			</p>
			<p class="small muted">Móveis para viver. Espaço para permanecer.</p>
		</div>
	</section>
	<section class="container story-section brand-story" data-cascade use:editorial>
		<div>
			<p class="muted">Nossa forma de pensar</p>
			<h2>Morar é dar forma<br />ao que importa.</h2>
		</div>
		<div>
			<p>
				A Forma reúne linhas contemporâneas, materiais naturais e atenção aos detalhes. Cada peça
				faz parte de um jeito de morar mais simples e acolhedor.
			</p>
			<p class="muted">Escolhas que atravessam o tempo e encontram seu lugar na rotina.</p>
			<a href="/moveis" class="text-link">Conheça as peças <ArrowUpRight size={18} /></a>
		</div>
	</section>
{:else}<div class="container page-shell information-page">
		<nav class="breadcrumb" aria-label="Você está em">
			<a href="/">Início</a><span>/</span><span>{title}</span>
		</nav>
		<div class="page-heading" data-cascade>
			<h1>{title}.</h1>
			<p class="muted">
				{kind === 'contato'
					? 'Vamos conversar sobre a sua próxima escolha.'
					: kind === 'ajuda'
						? 'Informações para escolher com tranquilidade.'
						: 'Informações sobre sua conta e o uso da loja.'}
			</p>
		</div>
		{#if kind === 'ajuda'}<div class="faq-list" data-cascade>
				{#each faqs as [question, answer]}<details>
						<summary>{question}<ChevronDown size={18} /></summary>
						<p class="muted">{answer}</p>
					</details>{/each}<a href="/contato" class="text-link"
					>Ainda precisa de ajuda? Fale com a gente <ArrowUpRight size={17} /></a
				>
			</div>{:else if kind === 'contato'}<div class="contact-layout">
				<div>
					<h2>Estamos por aqui.</h2>
					<p class="muted">Envie sua dúvida sobre produtos, entrega ou pedidos.</p>
					<div class="notice-panel">Sua mensagem será registrada para atendimento.</div>
					<a href="/ajuda" class="text-link"
						>Veja as dúvidas frequentes <ArrowUpRight size={18} /></a
					>
				</div>
				<form class="checkout-form" data-reveal onsubmit={send}>
					<label
						>Seu nome<input
							required
							minlength="2"
							maxlength="100"
							bind:value={contactName}
							autocomplete="name"
							placeholder="Seu nome"
						/></label
					><label
						>E-mail<input
							type="email"
							bind:value={contactEmail}
							required
							autocomplete="off"
							placeholder="voce@exemplo.com"
						/></label
					><label
						>Assunto<select bind:value={subject}
							><option>Dúvida sobre um produto</option><option>Entrega e montagem</option><option
								>Outro assunto</option
							></select
						></label
					><label
						>Mensagem<textarea
							bind:value={body}
							required
							minlength="5"
							maxlength="5000"
							rows="5"
							placeholder="Como podemos ajudar?"
						></textarea></label
					><button class="button" disabled={busy}
						>{busy ? 'Enviando...' : 'Enviar mensagem'} <ArrowUpRight size={17} /></button
					>{#if sent}<p class="notice-panel" role="status">
							<Check size={20} />Mensagem recebida. Obrigado por entrar em contato.
						</p>{/if}
					<p role="status">{message}</p>
				</form>
			</div>{:else}<div class="legal-copy" data-reveal>
				<h2>{kind === 'privacidade' ? 'Dados e serviços' : 'Uso da loja'}</h2>
				<p>
					O site possui cadastro, autenticação, catálogo, estoque e pedidos persistidos. O checkout
					apresenta o total e a cotação de frete antes da confirmação do pedido.
				</p>
				{#if kind === 'privacidade'}<h2>Dados no navegador</h2>
					<p>
						A sacola e os favoritos são salvos no armazenamento local do navegador. O cadastro, os
						endereços informados no checkout e os pedidos são enviados à API e persistidos no banco.
						A sessão utiliza um cookie HttpOnly. O CEP é consultado no ViaCEP.
					</p>
					<h2>Recursos externos</h2>
					<p>
						Imagens de referência são carregadas do Unsplash. Ao carregá-las, o navegador faz
						requisições a esse serviço externo.
					</p>{:else}<h2>Antes do lançamento</h2>
					<p>
						As condições comerciais, entrega, montagem, trocas e devoluções precisam ser definidas
						para a operação real. Este texto provisório não substitui os termos finais da loja.
					</p>{/if}<a href="/contato" class="text-link"
					>Fale com a gente <ArrowUpRight size={18} /></a
				>
			</div>{/if}
	</div>{/if}
