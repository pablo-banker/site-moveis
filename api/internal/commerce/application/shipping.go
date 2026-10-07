package application

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"sort"
	"strings"
	"time"
)

type ShippingInput struct {
	CEP   string `json:"cep"`
	Items []Item `json:"items"`
}
type ShippingQuote struct {
	ID          string    `json:"id"`
	Fingerprint string    `json:"-"`
	Provider    string    `json:"provider"`
	Service     string    `json:"service"`
	AmountCents int64     `json:"amountCents"`
	MinDays     int       `json:"minDays"`
	MaxDays     int       `json:"maxDays"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Simulated   bool      `json:"simulated"`
}

// Adapters own transport-specific data enrichment and carrier integration.
// The application persists an immutable quote bound to customer, CEP and cart.
type ShippingProvider interface {
	Quote(context.Context, ShippingInput) (ShippingQuote, error)
}

var QuoteInvalid = &Failure{409, "shipping_quote_invalid", "O frete expirou ou a seleção mudou. Calcule o frete novamente."}

func ShippingFingerprint(in ShippingInput) string {
	items := append([]Item(nil), in.Items...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].ID == items[j].ID {
			return items[i].Finish < items[j].Finish
		}
		return items[i].ID < items[j].ID
	})
	body, _ := json.Marshal(ShippingInput{CEP: strings.ReplaceAll(in.CEP, "-", ""), Items: items})
	return fmt.Sprintf("%x", sha256.Sum256(body))
}
func (s *Service) QuoteShipping(ctx context.Context, user string, in ShippingInput) (ShippingQuote, error) {
	if !cep.MatchString(in.CEP) {
		return ShippingQuote{}, Invalid("Informe um CEP válido.")
	}
	if err := validateItems(in.Items); err != nil {
		return ShippingQuote{}, err
	}
	for _, item := range in.Items {
		variants, err := s.orders.Variants(ctx, item.ID)
		if err != nil {
			return ShippingQuote{}, err
		}
		found := false
		for _, v := range variants {
			if v.Finish == item.Finish {
				found = true
			}
		}
		if !found {
			return ShippingQuote{}, StockUnavailable
		}
	}
	in.CEP = strings.ReplaceAll(in.CEP, "-", "")
	quote, err := s.shipping.Quote(ctx, in)
	if err != nil {
		return ShippingQuote{}, err
	}
	if quote.AmountCents < 0 || quote.AmountCents > 1000000000 || quote.MinDays < 1 || quote.MaxDays < quote.MinDays || quote.Provider == "" || quote.Service == "" {
		return ShippingQuote{}, fmt.Errorf("invalid shipping provider quote")
	}
	quote.ID = uuid.NewString()
	quote.Fingerprint = ShippingFingerprint(in)
	// Never outlive the carrier's validity; cap quotes to fifteen minutes.
	limit := time.Now().Add(15 * time.Minute)
	if quote.ExpiresAt.IsZero() || quote.ExpiresAt.After(limit) {
		quote.ExpiresAt = limit
	}
	if !quote.ExpiresAt.After(time.Now()) {
		return ShippingQuote{}, QuoteInvalid
	}
	if err := s.orders.SaveQuote(ctx, user, quote); err != nil {
		return ShippingQuote{}, err
	}
	return quote, nil
}
