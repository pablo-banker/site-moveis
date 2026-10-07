<script lang="ts">
	import Parallax from './Parallax.svelte';
	const colors = [
		['Branco', '--color-surface', '#FFFFFF'],
		['Grafite', '--color-primary', '#242424'],
		['Névoa', '--color-background', '#F7F7F5'],
		['Texto', '--color-foreground', '#242424'],
		['Cinza', '--color-muted', '#666666'],
		['Borda', '--color-border', '#E2E2DE'],
	];
</script>

<svelte:head><title>Sistema visual — Forma</title></svelte:head>
<div class="container page-shell">
	<div class="page-heading" data-cascade>
		<h1>Sistema visual.</h1>
		<p class="muted">Tokens globais, componentes e estudo do fundo em camadas.</p>
	</div>
	<section>
		<h2>Paleta minimalista</h2>
		<div class="swatches">
			{#each colors as [name, token, hex]}<div>
					<span style:background={`var(${token})`}></span><strong>{name}</strong>
					<p>{hex}</p>
					<code>{token}</code>
				</div>{/each}
		</div>
	</section>
	<section class="design-section">
		<h2>Tipografia e ações</h2>
		<h1>Espaço para viver.</h1>
		<p>Texto base de 16 px, contraste claro e hierarquia consistente.</p>
		<div class="form-actions">
			<button class="button">Ação principal</button><button class="button-outline"
				>Ação secundária</button
			><button class="button" disabled>Indisponível</button>
		</div>
	</section>
	<section class="parallax-study">
		<Parallax />
		<div>
			<h2>Fundo em camadas</h2>
			<p>Sala minimalista em cinco imagens PNG.<br />Role a página para testar a profundidade.</p>
		</div>
	</section>
</div>
