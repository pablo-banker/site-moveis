package postgres

import (
	"context"
	"errors"
	"forma/api/internal/catalog/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type products struct{ pool *pgxpool.Pool }
type taxonomy struct{ pool *pgxpool.Pool }

const columns = `p.id,p.name,c.name,r.name,p.price_cents,p.image,p.dimensions,p.material,p.label,p.finishes`
const joins = ` FROM products p JOIN categories c ON c.id=p.category_id JOIN rooms r ON r.id=p.room_id `
const where = ` WHERE p.active AND ($1='' OR strpos(lower(p.name || ' ' || p.material || ' ' || c.name),lower($1))>0) AND ($2='' OR c.id=$2) AND ($3='' OR r.id=$3) `

func scan(row pgx.Row) (domain.Product, error) {
	var p domain.Product
	err := row.Scan(&p.ID, &p.Name, &p.Category, &p.Room, &p.PriceCents, &p.Image, &p.Dimensions, &p.Material, &p.Label, &p.Finishes)
	return p, err
}
func (repo *products) List(ctx context.Context, f domain.Filter) (domain.ProductPage, error) {
	result := domain.ProductPage{Products: []domain.Product{}}
	// One read transaction keeps count and rows on the same snapshot.
	tx, err := repo.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)
	if err = tx.QueryRow(ctx, `SELECT count(*)`+joins+where, f.Query, f.Category, f.Room).Scan(&result.Total); err != nil {
		return result, err
	}
	order := `p.position,p.id`
	switch f.Sort {
	case "price_asc":
		order = `p.price_cents,p.id`
	case "price_desc":
		order = `p.price_cents DESC,p.id`
	}
	rows, err := tx.Query(ctx, `SELECT `+columns+joins+where+` ORDER BY `+order+` LIMIT $4 OFFSET $5`, f.Query, f.Category, f.Room, f.Limit, f.Offset)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		p, e := scan(rows)
		if e != nil {
			return result, e
		}
		result.Products = append(result.Products, p)
	}
	if err = rows.Err(); err != nil {
		return result, err
	}
	return result, tx.Commit(ctx)
}
func (repo *products) Find(ctx context.Context, id string) (domain.Product, error) {
	p, err := scan(repo.pool.QueryRow(ctx, `SELECT `+columns+joins+` WHERE p.active AND p.id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return p, domain.ErrNotFound
	}
	return p, err
}
func (repo *taxonomy) list(ctx context.Context, table string) ([]domain.Taxon, error) {
	rows, err := repo.pool.Query(ctx, `SELECT id,name FROM `+table+` ORDER BY position,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Taxon{}
	for rows.Next() {
		var t domain.Taxon
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}
func (repo *taxonomy) Categories(ctx context.Context) ([]domain.Taxon, error) {
	return repo.list(ctx, "categories")
}
func (repo *taxonomy) Rooms(ctx context.Context) ([]domain.Taxon, error) {
	return repo.list(ctx, "rooms")
}
