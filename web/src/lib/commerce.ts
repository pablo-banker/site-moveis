export type ShippingQuote = {
	id: string;
	provider: string;
	service: string;
	amountCents: number;
	minDays: number;
	maxDays: number;
	expiresAt: string;
	simulated: boolean;
};
export type Customer = { id: string; name: string; email: string; emailVerified: boolean };
export type Order = {
	id: string;
	status: 'awaiting_payment' | 'cancelled' | 'paid';
	subtotalCents: number;
	shippingCents: number;
	shipping?: ShippingQuote;
	totalCents: number;
	createdAt: string;
	address?: { cep: string; street: string; number: string; city: string; state: string };
	items: { id: string; finish: string; quantity: number; name: string; unitPriceCents: number }[];
};
export async function commerce<T>(path: string, body?: unknown): Promise<T> {
	const response = await fetch(`/backend/${path}`, {
		method: body === undefined ? 'GET' : 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: body === undefined ? undefined : JSON.stringify(body),
	});
	const result = await response.json();
	if (!response.ok)
		throw new Error(result.error?.message || 'Não foi possível concluir a solicitação.');
	return result;
}
export const orderStatus = (status: Order['status']) =>
	({
		awaiting_payment: 'Aguardando pagamento',
		cancelled: 'Cancelado',
		paid: 'Pagamento aprovado',
	})[status];
