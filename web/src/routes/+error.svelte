<script lang="ts">
	import { page } from '$app/state';
</script>

<svelte:head><title>Página não encontrada — Forma</title></svelte:head>
<div class="container empty-state">
	<p class="muted">{page.status}</p>
	<h1>Não encontramos esse espaço.</h1>
	<p>Volte ao catálogo para continuar sua visita.</p>
	<a href="/moveis" class="button">Explorar móveis</a>
</div>
