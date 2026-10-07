package domain

import "errors"

var ErrNotFound = errors.New("product not found")

type Product struct {
	ID         string
	Name       string
	Category   string
	Room       string
	PriceCents int64
	Image      string
	Dimensions string
	Material   string
	Label      string
	Finishes   []string
}

type Taxon struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Filter struct {
	Query, Category, Room, Sort string
	Limit, Offset               int
}

type ProductPage struct {
	Products []Product
	Total    int
}
