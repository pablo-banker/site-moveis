package simulated

import (
	"context"
	a "forma/api/internal/commerce/application"
	"testing"
)

func TestShippingDestinationAndQuantity(t *testing.T) {
	for _, tt := range []struct {
		cep      string
		quantity int
		amount   int64
	}{{"01001000", 1, 14900}, {"88301401", 1, 18900}, {"88301401", 2, 22800}, {"50000000", 1, 22900}} {
		q, err := (Shipping{}).Quote(context.Background(), a.ShippingInput{CEP: tt.cep, Items: []a.Item{{Quantity: tt.quantity}}})
		if err != nil || q.AmountCents != tt.amount || q.MinDays < 1 || q.MaxDays < q.MinDays {
			t.Fatalf("%+v: %+v %v", tt, q, err)
		}
	}
}
