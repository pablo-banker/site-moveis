<script lang="ts">
	import { slug } from '$lib/catalog';
	import { ArrowUpRight, ArrowDown } from '@lucide/svelte';
	import { roomStories } from '$lib/room-stories';
	import { editorial } from '$lib/actions/editorial';
</script>

<svelte:head><title>Ambientes — Forma</title></svelte:head>
<div class="container page-shell rooms-page">
	<nav class="breadcrumb" aria-label="Você está em">
		<a href="/">Início</a><span>/</span><span>Ambientes</span>
	</nav>
	<section class="rooms-intro" use:editorial>
		<div data-cascade>
			<p class="muted">Inspiração para morar</p>
			<h1>Cada espaço,<br />sua forma.</h1>
			<p>
				Uma casa não acontece de uma vez. Comece pelo lugar onde você quer estar e descubra as peças
				que acompanham esse momento.
			</p>
			<a class="text-link" href="#ambientes">Escolha seu ambiente <ArrowDown size={18} /></a>
		</div>
		<figure>
			<img
				src={roomStories[0].image}
				alt="Sala iluminada com uma composição de móveis"
				fetchpriority="high"
			/>
			<figcaption>Da inspiração ao seu cotidiano.</figcaption>
		</figure>
	</section>
	<section id="ambientes" class="rooms-grid rooms-editorial-grid">
		{#each roomStories as room}<a
				class="room-card"
				href={`/ambientes/${slug(room.name)}`}
				use:editorial
				data-reveal
			>
				<div>
					<img src={room.image} alt={`Inspiração para ${room.name.toLowerCase()}`} loading="lazy" />
				</div>
				<span
					><h2>{room.name}</h2>
					<span class="room-link-arrow"><ArrowUpRight size={25} /></span></span
				>
				<h3>{room.title}</h3>
				<p class="muted">{room.intro}</p>
			</a>{/each}
	</section>
	<section class="rooms-outro" data-cascade>
		<h2>Já sabe o que procura?</h2>
		<p class="muted">Escolha por categoria, material ou nome no catálogo.</p>
		<a class="button" href="/moveis">Ver todos os móveis <ArrowUpRight size={18} /></a>
	</section>
</div>
