<script lang="ts">
	let products = $derived(page.data.catalog?.products ?? []);
	let categories = $derived(page.data.catalog?.categories ?? ['Todos']);
	import { page } from '$app/state';
	import { ArrowRight, ArrowUpRight, Truck, ShieldCheck, Armchair } from '@lucide/svelte';
	import { slug } from '$lib/catalog';
	import { editorial } from '$lib/actions/editorial';
	import Parallax from '$lib/components/Parallax.svelte';
	import { roomStories } from '$lib/room-stories';
	import ProductCard from '$lib/components/ProductCard.svelte';
</script>

<svelte:head><title>Forma — Móveis para viver</title></svelte:head>
<section class="home-hero">
	<Parallax />
	<div class="hero-shade"></div>
	<div class="container hero-content">
		<div class="hero-copy" data-cascade>
			<p class="hero-kicker">Coleção Essencial</p>
			<h1>Mais espaço<br />para viver.</h1>
			<p>Formas simples, materiais que acolhem.<br />Móveis que fazem parte da sua história.</p>
			<a class="button" href="#colecao">Explore a coleção <ArrowUpRight size={18} /></a>
		</div>
		<span class="hero-caption">Design que encontra o seu dia a dia.</span>
	</div>
</section>
<section id="colecao" class="collection-chapter" use:editorial>
	<div class="collection-chapter-stage">
		<div class="collection-chapter-copy" data-cascade>
			<p>Coleção Essencial</p>
			<h2>A casa muda.<br />O que importa<br />fica.</h2>
			<p>
				O livro por perto. A pausa no sofá. A conversa sem pressa. Antes de escolher um móvel,
				imagine o que você quer viver ao redor dele.
			</p>
			<a class="text-link" href="/ambientes">Encontre seu ambiente <ArrowRight size={19} /></a>
		</div>
		<figure class="collection-chapter-image">
			<img
				src={roomStories[0].image}
				alt="Uma sala acolhedora, com luz natural e peças reunidas em torno da mesa"
				loading="lazy"
			/>
			<figcaption>Formas simples. Espaço para a sua vida.</figcaption>
		</figure>
		<span class="chapter-wordmark" aria-hidden="true">forma</span>
	</div>
</section>
<section class="room-feature editorial-cover" use:editorial>
	<div class="room-photo editorial-cover-media">
		<img
			src={roomStories[0].image}
			alt="Ambiente de estar com tons neutros e iluminação natural"
			loading="lazy"
		/>
	</div>
	<div class="room-copy" data-cascade>
		<p>Um ambiente, muitas histórias</p>
		<h2>O seu lugar<br />de estar.</h2>
		<p>
			Um sofá para desacelerar. Uma mesa para reunir. Peças que conversam entre si, com espaço para
			aquilo que é só seu.
		</p>
		<a class="button button-white" href="/ambientes/sala-de-estar"
			>Explore a sala de estar <ArrowUpRight size={19} /></a
		>
	</div>
</section>
<section class="container selection-section">
	<div class="section-heading" data-cascade>
		<div>
			<p class="muted">Escolhas que ficam</p>
			<h2>Peças para o seu estar.</h2>
		</div>
		<a class="text-link" href="/ambientes/sala-de-estar#pecas"
			>Conheça a seleção <ArrowRight size={17} /></a
		>
	</div>
	<div class="product-grid" data-cascade>
		{#each products.filter((p) => p.room === 'Sala de estar') as product}<ProductCard
				{product}
			/>{/each}
	</div>
</section>
<div class="signature-line" use:editorial><p>O seu jeito de morar.</p></div>
<section class="container category-section">
	<div class="section-heading" data-cascade>
		<h2>Encontre sua forma.</h2>
		<a class="text-link" href="/moveis">Ver todos os móveis <ArrowRight size={17} /></a>
	</div>
	<div class="category-list" data-cascade>
		{#each categories.slice(1) as category}<a
				href={`/moveis?categoria=${encodeURIComponent(category)}`}
				>{category}<ArrowUpRight size={17} /></a
			>{/each}
	</div>
</section>
<section class="container other-rooms">
	<div class="section-heading" data-cascade>
		<div>
			<p class="muted">Outras formas de viver</p>
			<h2>Uma casa, muitos momentos.</h2>
		</div>
		<a class="text-link" href="/ambientes">Ver ambientes <ArrowRight size={17} /></a>
	</div>
	<div class="other-rooms-grid" data-cascade>
		{#each roomStories.slice(1) as room}<a class="room-card" href={`/ambientes/${slug(room.name)}`}
				><div>
					<img src={room.image} alt={`Inspiração de ${room.name.toLowerCase()}`} loading="lazy" />
				</div>
				<span
					><h3>{room.name}</h3>
					<ArrowUpRight size={20} /></span
				>
				<p class="muted small">{room.title}</p></a
			>{/each}
	</div>
</section>
<section class="container values" data-cascade>
	<div>
		<Armchair size={28} strokeWidth={1.3} />
		<h3>Design com propósito</h3>
		<p>Dimensões, materiais e acabamentos para escolher com atenção.</p>
	</div>
	<div>
		<Truck size={28} strokeWidth={1.3} />
		<h3>Da escolha à sua casa</h3>
		<p>Consulte as informações de entrega e montagem antes de comprar.</p>
	</div>
	<div>
		<ShieldCheck size={28} strokeWidth={1.3} />
		<h3>Detalhes que importam</h3>
		<p>Conheça os cuidados para manter cada peça por mais tempo.</p>
	</div>
</section>
