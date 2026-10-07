package postgres

import (
	"forma/api/internal/catalog/application"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Factory struct{ pool *pgxpool.Pool }

func NewFactory(pool *pgxpool.Pool) application.RepositoryFactory { return &Factory{pool: pool} }
func (f *Factory) Products() application.ProductRepository        { return &products{pool: f.pool} }
func (f *Factory) Taxonomy() application.TaxonomyRepository       { return &taxonomy{pool: f.pool} }

var _ application.RepositoryFactory = (*Factory)(nil)
