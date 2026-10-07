<script lang="ts">
	import { ArrowUpRight } from '@lucide/svelte';
</script>

<footer class="site-footer">
	<div class="container">
		<div class="footer-top">
			<div>
				<a href="/" class="wordmark">forma<span>®</span></a>
				<p>Móveis para viver.<br />Design para permanecer.</p>
			</div>
			<div>
				<h2>Explore</h2>
				<a href="/moveis">Todos os móveis</a><a href="/ambientes">Ambientes</a><a href="/sobre"
					>Nossa história</a
				>
			</div>
			<div>
				<h2>Podemos ajudar</h2>
				<a href="/ajuda">Entrega e montagem</a><a href="/ajuda">Trocas e cuidados</a><a
					href="/contato">Fale com a gente <ArrowUpRight size={14} /></a
				>
			</div>
			<div>
				<h2>Sua forma de morar</h2>
				<a href="/conta">Minha conta</a><a href="/favoritos">Meus favoritos</a><a
					href="/conta/pedidos">Meus pedidos</a
				>
			</div>
		</div>
		<div class="footer-bottom">
			<span>© {new Date().getFullYear()} Forma.</span>
			<div><a href="/privacidade">Privacidade</a><a href="/termos">Termos de uso</a></div>
		</div>
	</div>
</footer>
