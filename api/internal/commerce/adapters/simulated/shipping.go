package simulated

import (
	"context"
	a "forma/api/internal/commerce/application"
)

type Shipping struct{}

func NewShipping() a.ShippingProvider { return Shipping{} }
func (Shipping) Quote(_ context.Context, in a.ShippingInput) (a.ShippingQuote, error) {
	// Illustrative rates by destination band and number of pieces, not carrier tariffs.
	amount := int64(14900)
	min, max := 8, 12
	if in.CEP[0] >= '8' {
		amount = 18900
		min, max = 10, 15
	} else if in.CEP[0] >= '4' {
		amount = 22900
		min, max = 12, 18
	}
	count := 0
	for _, item := range in.Items {
		count += item.Quantity
	}
	amount += int64(count-1) * 3900
	return a.ShippingQuote{Provider: "simulated", Service: "Entrega padrão", AmountCents: amount, MinDays: min, MaxDays: max, Simulated: true}, nil
}
