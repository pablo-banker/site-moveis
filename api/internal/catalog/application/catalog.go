package application

import (
	"context"
	"errors"
	"forma/api/internal/catalog/domain"
	"strings"
	"unicode/utf8"
)

type ProductRepository interface {
	List(context.Context, domain.Filter) (domain.ProductPage, error)
	Find(context.Context, string) (domain.Product, error)
}

type TaxonomyRepository interface {
	Categories(context.Context) ([]domain.Taxon, error)
	Rooms(context.Context) ([]domain.Taxon, error)
}

// RepositoryFactory creates a compatible family of catalog repositories.
// The application depends on this abstraction, never on pgx or SQL.
type RepositoryFactory interface {
	Products() ProductRepository
	Taxonomy() TaxonomyRepository
}

var ErrInvalidFilter = errors.New("invalid catalog filter")

type Service struct {
	products ProductRepository
	taxonomy TaxonomyRepository
}

func NewService(factory RepositoryFactory) *Service {
	return &Service{products: factory.Products(), taxonomy: factory.Taxonomy()}
}

func (s *Service) List(ctx context.Context, filter domain.Filter) (domain.ProductPage, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	if filter.Limit == 0 {
		filter.Limit = 24
	}
	if filter.Sort == "" {
		filter.Sort = "relevance"
	}
	if filter.Limit < 1 || filter.Limit > 100 || filter.Offset < 0 || utf8.RuneCountInString(filter.Query) > 120 {
		return domain.ProductPage{}, ErrInvalidFilter
	}
	switch filter.Sort {
	case "relevance", "price_asc", "price_desc":
	default:
		return domain.ProductPage{}, ErrInvalidFilter
	}
	return s.products.List(ctx, filter)
}
func (s *Service) Find(ctx context.Context, id string) (domain.Product, error) {
	return s.products.Find(ctx, id)
}
func (s *Service) Categories(ctx context.Context) ([]domain.Taxon, error) {
	return s.taxonomy.Categories(ctx)
}
func (s *Service) Rooms(ctx context.Context) ([]domain.Taxon, error) { return s.taxonomy.Rooms(ctx) }
