package application_test

import (
	"context"
	"errors"
	"forma/api/internal/catalog/application"
	"forma/api/internal/catalog/domain"
	"testing"
)

// A second compatible family is provided only for isolated application tests.
type memoryFactory struct {
	productsCreated, taxonomyCreated bool
	filter                           domain.Filter
}

func (f *memoryFactory) Products() application.ProductRepository {
	f.productsCreated = true
	return &memoryProducts{f}
}
func (f *memoryFactory) Taxonomy() application.TaxonomyRepository {
	f.taxonomyCreated = true
	return memoryTaxonomy{}
}

type memoryProducts struct{ factory *memoryFactory }

func (p *memoryProducts) List(_ context.Context, f domain.Filter) (domain.ProductPage, error) {
	p.factory.filter = f
	return domain.ProductPage{Products: []domain.Product{}}, nil
}
func (*memoryProducts) Find(context.Context, string) (domain.Product, error) {
	return domain.Product{}, domain.ErrNotFound
}

type memoryTaxonomy struct{}

func (memoryTaxonomy) Categories(context.Context) ([]domain.Taxon, error) {
	return []domain.Taxon{{ID: "sofas", Name: "Sofás"}}, nil
}
func (memoryTaxonomy) Rooms(context.Context) ([]domain.Taxon, error) {
	return []domain.Taxon{{ID: "sala-de-estar", Name: "Sala de estar"}}, nil
}
func TestCompatibleFamilyAndValidation(t *testing.T) {
	factory := &memoryFactory{}
	service := application.NewService(factory)
	if !factory.productsCreated || !factory.taxonomyCreated {
		t.Fatal("service must create both repositories through the abstract factory")
	}
	_, err := service.List(context.Background(), domain.Filter{Query: "  Arco  "})
	if err != nil || factory.filter.Query != "Arco" || factory.filter.Limit != 24 || factory.filter.Sort != "relevance" {
		t.Fatalf("defaults: %+v %v", factory.filter, err)
	}
	for _, filter := range []domain.Filter{{Limit: -1}, {Limit: 101}, {Offset: -1}, {Sort: "sql"}} {
		if _, err := service.List(context.Background(), filter); !errors.Is(err, application.ErrInvalidFilter) {
			t.Fatalf("invalid filter accepted: %+v", filter)
		}
	}
	if _, err := service.Find(context.Background(), "missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("not found not propagated")
	}
}
