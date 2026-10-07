package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	a "forma/api/internal/commerce/application"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"sort"
	"time"
)

type Factory struct{ pool *pgxpool.Pool }

func NewFactory(pool *pgxpool.Pool) a.RepositoryFactory { return &Factory{pool} }
func (f *Factory) Identity() a.IdentityRepository       { return &identity{f.pool} }
func (f *Factory) Orders() a.OrderRepository            { return &orders{f.pool} }

type identity struct{ pool *pgxpool.Pool }
type orders struct{ pool *pgxpool.Pool }

func (r *identity) Create(ctx context.Context, c a.Customer, hash string, delivery *a.AccountDelivery) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO customers(id,name,email,password_hash)VALUES($1,$2,$3,$4)`, c.ID, c.Name, c.Email, hash)
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && pgerr.Code == "23505" {
		return a.Conflict
	}
	if err != nil {
		return err
	}
	if delivery != nil {
		if _, err = tx.Exec(ctx, `INSERT INTO account_challenges(token_hash,customer_id,purpose,expires_at)VALUES($1,$2,'verify',$3)`, delivery.TokenHash, c.ID, delivery.ExpiresAt); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO email_outbox(id,recipient,subject,body)VALUES($1,$2,$3,$4)`, uuid.NewString(), delivery.Email, delivery.Subject, delivery.Body); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (r *identity) ByEmail(ctx context.Context, email string) (a.Credentials, error) {
	var c a.Credentials
	err := r.pool.QueryRow(ctx, `SELECT id::text,name,email,password_hash,email_verified_at IS NOT NULL FROM customers WHERE email=$1`, email).Scan(&c.Customer.ID, &c.Customer.Name, &c.Customer.Email, &c.Hash, &c.Customer.EmailVerified)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, a.Unauthorized
	}
	return c, err
}
func (r *identity) Session(ctx context.Context, hash, user string, expires time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT id FROM customers WHERE id=$1 FOR UPDATE`, user); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO customer_sessions(token_hash,customer_id,expires_at)VALUES($1,$2,$3)`, hash, user, expires); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM customer_sessions WHERE customer_id=$1 AND token_hash IN (SELECT token_hash FROM customer_sessions WHERE customer_id=$1 ORDER BY created_at DESC,token_hash DESC OFFSET 20)`, user); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *identity) Current(ctx context.Context, hash string) (a.Customer, error) {
	var c a.Customer
	err := r.pool.QueryRow(ctx, `SELECT c.id::text,c.name,c.email,c.email_verified_at IS NOT NULL FROM customers c JOIN customer_sessions s ON s.customer_id=c.id WHERE s.token_hash=$1 AND s.expires_at>now()`, hash).Scan(&c.ID, &c.Name, &c.Email, &c.EmailVerified)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, a.Unauthorized
	}
	return c, err
}
func (r *identity) Logout(ctx context.Context, hash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM customer_sessions WHERE token_hash=$1`, hash)
	return err
}

func (r *identity) UpdateName(ctx context.Context, hash, name string) (a.Customer, error) {
	var c a.Customer
	err := r.pool.QueryRow(ctx, `UPDATE customers c SET name=$2 FROM customer_sessions s WHERE s.customer_id=c.id AND s.token_hash=$1 AND s.expires_at>clock_timestamp() RETURNING c.id::text,c.name,c.email,c.email_verified_at IS NOT NULL`, hash, name).Scan(&c.ID, &c.Name, &c.Email, &c.EmailVerified)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, a.Unauthorized
	}
	return c, err
}

type querier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func read(ctx context.Context, q querier, user, id string) (a.Order, error) {
	var o a.Order
	var address []byte
	err := q.QueryRow(ctx, `SELECT id::text,status,subtotal_cents,shipping_cents,total_cents,address,created_at,shipping,reservation_expires_at FROM orders WHERE customer_id=$1 AND id=$2`, user, id).Scan(&o.ID, &o.Status, &o.SubtotalCents, &o.ShippingCents, &o.TotalCents, &address, &o.CreatedAt, &o.Shipping, &o.ReservationExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return o, a.NotFound
	}
	if err != nil {
		return o, err
	}
	if err = json.Unmarshal(address, &o.Address); err != nil {
		return o, err
	}
	rows, err := q.Query(ctx, `SELECT product_id,finish,quantity,name,unit_price_cents FROM order_items WHERE order_id=$1 ORDER BY product_id,finish`, id)
	if err != nil {
		return o, err
	}
	defer rows.Close()
	o.Items = []a.OrderItem{}
	for rows.Next() {
		var i a.OrderItem
		if err := rows.Scan(&i.ID, &i.Finish, &i.Quantity, &i.Name, &i.UnitPriceCents); err != nil {
			return o, err
		}
		o.Items = append(o.Items, i)
	}
	return o, rows.Err()
}
func (r *orders) Find(ctx context.Context, user, id string) (a.Order, error) {
	return read(ctx, r.pool, user, id)
}
func (r *orders) List(ctx context.Context, user string) ([]a.Order, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text FROM orders WHERE customer_id=$1 ORDER BY created_at DESC,id LIMIT 100`, user)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result := []a.Order{}
	for _, id := range ids {
		o, err := r.Find(ctx, user, id)
		if err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, nil
}
func (r *orders) Create(ctx context.Context, user string, in a.OrderInput) (a.Order, error) {
	sort.Slice(in.Items, func(i, j int) bool {
		if in.Items[i].ID == in.Items[j].ID {
			return in.Items[i].Finish < in.Items[j].Finish
		}
		return in.Items[i].ID < in.Items[j].ID
	})
	body, err := json.Marshal(in)
	if err != nil {
		return a.Order{}, err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(body))
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return a.Order{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, user+in.Key); err != nil {
		return a.Order{}, err
	}
	var id, existingHash string
	err = tx.QueryRow(ctx, `SELECT id::text,request_hash FROM orders WHERE customer_id=$1 AND idempotency_key=$2`, user, in.Key).Scan(&id, &existingHash)
	if err == nil {
		if existingHash != hash {
			return a.Order{}, a.Conflict
		}
		return read(ctx, tx, user, id)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return a.Order{}, err
	}
	var verified bool
	if err = tx.QueryRow(ctx, `SELECT email_verified_at IS NOT NULL FROM customers WHERE id=$1 FOR SHARE`, user).Scan(&verified); err != nil {
		return a.Order{}, err
	}
	if !verified {
		return a.Order{}, &a.Failure{Status: 403, Code: "email_unverified", Message: "Confirme seu e-mail antes de finalizar o pedido."}
	}
	var quote a.ShippingQuote
	var fingerprint string
	err = tx.QueryRow(ctx, `SELECT quote,fingerprint FROM shipping_quotes WHERE id=$1 AND customer_id=$2 AND expires_at>now() FOR SHARE`, in.ShippingQuoteID, user).Scan(&quote, &fingerprint)
	if errors.Is(err, pgx.ErrNoRows) {
		return a.Order{}, a.QuoteInvalid
	}
	if err != nil {
		return a.Order{}, err
	}
	if fingerprint != a.ShippingFingerprint(a.ShippingInput{CEP: in.Address.CEP, Items: in.Items}) {
		return a.Order{}, a.QuoteInvalid
	}
	items := []a.OrderItem{}
	var subtotal int64
	for _, i := range in.Items {
		var price int64
		var stock int
		var name string
		err = tx.QueryRow(ctx, `SELECT p.price_cents,p.name,v.stock FROM product_variants v JOIN products p ON p.id=v.product_id WHERE v.product_id=$1 AND v.finish=$2 AND p.active FOR UPDATE OF v FOR SHARE OF p`, i.ID, i.Finish).Scan(&price, &name, &stock)
		if errors.Is(err, pgx.ErrNoRows) || err == nil && stock < i.Quantity {
			return a.Order{}, a.StockUnavailable
		}
		if err != nil {
			return a.Order{}, err
		}
		if price > 1000000000 {
			return a.Order{}, a.Invalid("Preço fora do limite permitido.")
		}
		if _, err = tx.Exec(ctx, `UPDATE product_variants SET stock=stock-$3 WHERE product_id=$1 AND finish=$2`, i.ID, i.Finish, i.Quantity); err != nil {
			return a.Order{}, err
		}
		subtotal += price * int64(i.Quantity)
		items = append(items, a.OrderItem{Item: i, Name: name, UnitPriceCents: price})
	}
	address, err := json.Marshal(in.Address)
	if err != nil {
		return a.Order{}, err
	}
	id = uuid.NewString()
	if _, err = tx.Exec(ctx, `INSERT INTO orders(id,customer_id,idempotency_key,request_hash,status,subtotal_cents,shipping_cents,total_cents,address,shipping)VALUES($1,$2,$3,$4,'awaiting_payment',$5,$6,$7,$8,$9)`, id, user, in.Key, hash, subtotal, quote.AmountCents, subtotal+quote.AmountCents, address, quote); err != nil {
		return a.Order{}, err
	}
	for _, i := range items {
		if _, err = tx.Exec(ctx, `INSERT INTO order_items(order_id,product_id,finish,name,quantity,unit_price_cents)VALUES($1,$2,$3,$4,$5,$6)`, id, i.ID, i.Finish, i.Name, i.Quantity, i.UnitPriceCents); err != nil {
			return a.Order{}, err
		}
	}
	o, err := read(ctx, tx, user, id)
	if err != nil {
		return o, err
	}
	return o, tx.Commit(ctx)
}
func (r *orders) Cancel(ctx context.Context, user, id string) (a.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return a.Order{}, err
	}
	defer tx.Rollback(ctx)
	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM orders WHERE id=$1 AND customer_id=$2 FOR UPDATE`, id, user).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return a.Order{}, a.NotFound
	}
	if err != nil {
		return a.Order{}, err
	}
	o, err := read(ctx, tx, user, id)
	if err != nil {
		return o, err
	}
	if status == "paid" {
		return o, a.Conflict
	}
	if status == "awaiting_payment" {
		for _, i := range o.Items {
			if _, err = tx.Exec(ctx, `UPDATE product_variants SET stock=stock+$3 WHERE product_id=$1 AND finish=$2`, i.ID, i.Finish, i.Quantity); err != nil {
				return o, err
			}
		}
		if _, err = tx.Exec(ctx, `UPDATE orders SET status='cancelled' WHERE id=$1`, id); err != nil {
			return o, err
		}
		o.Status = "cancelled"
	}
	return o, tx.Commit(ctx)
}

func (r *orders) Payment(ctx context.Context, user, id string, gateway a.PaymentGateway) (a.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return a.Order{}, err
	}
	defer tx.Rollback(ctx)
	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM orders WHERE id=$1 AND customer_id=$2 FOR UPDATE`, id, user).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return a.Order{}, a.NotFound
	}
	if err != nil {
		return a.Order{}, err
	}
	if status == "paid" {
		return read(ctx, tx, user, id)
	}
	var expired bool
	if err := tx.QueryRow(ctx, `SELECT reservation_expires_at<=clock_timestamp() FROM orders WHERE id=$1`, id).Scan(&expired); err != nil {
		return a.Order{}, err
	}
	if status == "cancelled" || expired {
		return a.Order{}, a.Conflict
	}
	if status == "awaiting_payment" {
		order, err := read(ctx, tx, user, id)
		if err != nil {
			return a.Order{}, err
		}
		outcome, err := gateway.Process(ctx, order)
		if err != nil {
			return a.Order{}, err
		}
		if outcome != "approved" && outcome != "declined" {
			return a.Order{}, fmt.Errorf("invalid payment provider result")
		}
		if _, err = tx.Exec(ctx, `INSERT INTO order_payments(order_id,outcome)VALUES($1,$2) ON CONFLICT(order_id) DO UPDATE SET outcome=excluded.outcome,updated_at=now()`, id, outcome); err != nil {
			return a.Order{}, err
		}
		if outcome == "approved" {
			if _, err = tx.Exec(ctx, `UPDATE orders SET status='paid' WHERE id=$1`, id); err != nil {
				return a.Order{}, err
			}
		}
	}
	o, err := read(ctx, tx, user, id)
	if err != nil {
		return o, err
	}
	return o, tx.Commit(ctx)
}

func (r *orders) Variants(ctx context.Context, id string) ([]a.Variant, error) {
	rows, err := r.pool.Query(ctx, `SELECT v.finish,v.stock FROM product_variants v JOIN products p ON p.id=v.product_id WHERE p.active AND p.id=$1 ORDER BY v.finish`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []a.Variant{}
	for rows.Next() {
		var v a.Variant
		if err := rows.Scan(&v.Finish, &v.Stock); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (r *orders) SaveQuote(ctx context.Context, user string, q a.ShippingQuote) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO shipping_quotes(id,customer_id,fingerprint,quote,expires_at) VALUES($1,$2,$3,$4,$5)`, q.ID, user, q.Fingerprint, q, q.ExpiresAt)
	return err
}

func (r *identity) AllowAuthentication(ctx context.Context, hash string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `INSERT INTO auth_attempts(key_hash,attempts,resets_at)VALUES($1,1,now()+interval '1 minute')
 ON CONFLICT(key_hash) DO UPDATE SET attempts=CASE WHEN auth_attempts.resets_at<=now() THEN 1 ELSE auth_attempts.attempts+1 END,
 resets_at=CASE WHEN auth_attempts.resets_at<=now() THEN now()+interval '1 minute' ELSE auth_attempts.resets_at END
 WHERE auth_attempts.resets_at<=now() OR auth_attempts.attempts<8 RETURNING attempts`, hash).Scan(&count)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
