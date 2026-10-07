<script lang="ts">
	import { ArrowDown, ArrowUpRight } from '@lucide/svelte';
	import { roomStory } from '$lib/room-stories';
	import { editorial } from '$lib/actions/editorial';
	import Catalog from './Catalog.svelte';
	let { room }: { room: string } = $props();
	let story = $derived(roomStory(room));
</script>

<svelte:head><title>{story.name} — Forma</title></svelte:head>
<div class="container page-shell room-editorial">
	<nav class="breadcrumb" aria-label="Você está em">
		<a href="/">Início</a><span>/</span><a href="/ambientes">Ambientes</a><span>/</span><span
			>{story.name}</span
		>
	</nav>
	<div class="room-editorial-heading" data-cascade>
		<p class="muted">{story.name}</p>
		<h1>{story.title}</h1>
		<p>{story.intro}</p>
		<a class="text-link" href="#pecas">Ver peças para este ambiente <ArrowDown size={18} /></a>
	</div>
	<figure class="room-editorial-photo" use:editorial>
		<img src={story.image} alt={`Inspiração de ${story.name.toLowerCase()}`} fetchpriority="high" />
		<figcaption>
			Ambiente de inspiração. A seleção abaixo reúne peças do catálogo para este espaço.
		</figcaption>
	</figure>
	<section class="room-editorial-story" data-cascade use:editorial>
		<div>
			<p class="muted">Compor com intenção</p>
			<h2>Os detalhes fazem<br />o seu espaço.</h2>
			<p>{story.detail}</p>
		</div>
		<ul>
			{#each story.tips as tip}<li>
					<span aria-hidden="true">↗</span>
					<p>{tip}</p>
				</li>{/each}
		</ul>
	</section>
</div>
<section id="pecas" class="room-pieces">
	<div class="container section-heading" data-cascade>
		<div>
			<p class="muted">Da inspiração à escolha</p>
			<h2>Peças para {story.name.toLowerCase()}.</h2>
		</div>
		<a class="text-link" href="/moveis">Ver todos os móveis <ArrowUpRight size={17} /></a>
	</div>
	<Catalog {room} embedded />
</section>
