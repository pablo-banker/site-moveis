package simulated

import (
	"context"
	"fmt"
	a "forma/api/internal/commerce/application"
	"os"
)

type Gateway struct{ Outcome string }

func NewGateway() (a.PaymentGateway, error) {
	outcome := os.Getenv("PAYMENT_SIMULATED_OUTCOME")
	if outcome == "" {
		outcome = "approved"
	}
	if outcome != "approved" && outcome != "declined" {
		return nil, fmt.Errorf("invalid PAYMENT_SIMULATED_OUTCOME")
	}
	return Gateway{Outcome: outcome}, nil
}
func (g Gateway) Process(_ context.Context, _ a.Order) (string, error) { return g.Outcome, nil }
