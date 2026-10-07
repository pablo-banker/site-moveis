<script lang="ts">
	import Catalog from '$lib/components/Catalog.svelte';
	import ProductDetail from '$lib/components/ProductDetail.svelte';
	import Cart from '$lib/components/Cart.svelte';
	import Checkout from '$lib/components/Checkout.svelte';
	import Rooms from '$lib/components/Rooms.svelte';
	import RoomDetail from '$lib/components/RoomDetail.svelte';
	import AccountAccess from '$lib/components/AccountAccess.svelte';
	import Account from '$lib/components/Account.svelte';
	import Information from '$lib/components/Information.svelte';
	import Confirmation from '$lib/components/Confirmation.svelte';
	import DesignSystem from '$lib/components/DesignSystem.svelte';
	let { data } = $props();
</script>

{#key data.path}{#if data.path === 'moveis'}<Catalog
		/>{:else if data.path.startsWith('moveis/categoria/')}<Catalog
			categorySlug={data.path.split('/')[2]}
		/>{:else if data.path.startsWith('moveis/')}<ProductDetail
			id={data.path.split('/')[1]}
		/>{:else if data.path === 'favoritos'}<Catalog
			favoritesOnly
		/>{:else if data.path === 'carrinho'}<Cart />{:else if data.path === 'checkout'}<Checkout
		/>{:else if data.path === 'pedido-confirmado'}<Confirmation
		/>{:else if data.path === 'ambientes'}<Rooms
		/>{:else if data.path.startsWith('ambientes/')}<RoomDetail
			room={data.path.split('/')[1]}
		/>{:else if ['conta/recuperar-senha', 'conta/redefinir-senha', 'conta/confirmar-email'].includes(data.path)}<AccountAccess
			kind={data.path}
		/>{:else if data.path === 'conta' || data.path === 'conta/pedidos'}<Account
			orders={data.path === 'conta/pedidos'}
		/>{:else if data.path === 'design-system'}<DesignSystem />{:else}<Information
			kind={data.path}
		/>{/if}{/key}
