package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type Failure struct {
	Status        int
	Code, Message string
}

func (e *Failure) Error() string   { return e.Message }
func Invalid(message string) error { return &Failure{400, "invalid_input", message} }

var Unauthorized = &Failure{401, "unauthorized", "Entre na sua conta para continuar."}
var Conflict = &Failure{409, "conflict", "Não foi possível concluir com estes dados."}
var NotFound = &Failure{404, "not_found", "Pedido não encontrado."}
var StockUnavailable = &Failure{409, "stock_unavailable", "Quantidade indisponível para este acabamento."}

type Customer struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
}
type Credentials struct {
	Customer Customer
	Hash     string
}
type IdentityRepository interface {
	Create(context.Context, Customer, string, *AccountDelivery) error
	ByEmail(context.Context, string) (Credentials, error)
	Session(context.Context, string, string, time.Time) error
	Current(context.Context, string) (Customer, error)
	UpdateName(context.Context, string, string) (Customer, error)
	Logout(context.Context, string) error
	AllowAuthentication(context.Context, string) (bool, error)
}
type OrderRepository interface {
	SaveQuote(context.Context, string, ShippingQuote) error

	Create(context.Context, string, OrderInput) (Order, error)
	List(context.Context, string) ([]Order, error)
	Find(context.Context, string, string) (Order, error)
	Cancel(context.Context, string, string) (Order, error)
	Payment(context.Context, string, string, PaymentGateway) (Order, error)
	Variants(context.Context, string) ([]Variant, error)
}
type RepositoryFactory interface {
	Identity() IdentityRepository
	Orders() OrderRepository
}
type Service struct {
	identity   IdentityRepository
	orders     OrderRepository
	gateway    PaymentGateway
	shipping   ShippingProvider
	hashSlots  chan struct{}
	registered func(Customer) (*AccountDelivery, error)
}

func NewService(f RepositoryFactory, gateway PaymentGateway, shipping ShippingProvider) *Service {
	return &Service{identity: f.Identity(), orders: f.Orders(), gateway: gateway, shipping: shipping, hashSlots: PasswordHashSlots()}
}

type AuthInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type AuthResult struct {
	Customer Customer `json:"customer"`
	Token    string   `json:"token"`
}

func tokenHash(token string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(token))) }
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	return "argon2id$v1$" + base64.RawStdEncoding.EncodeToString(append(salt, hash...)), nil
}
func passwordMatches(password, encoded string) bool {
	if !strings.HasPrefix(encoded, "argon2id$v1$") {
		return false
	}
	raw, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(encoded, "argon2id$v1$"))
	if err != nil || len(raw) != 48 {
		return false
	}
	hash := argon2.IDKey([]byte(password), raw[:16], 3, 64*1024, 2, 32)
	return subtle.ConstantTimeCompare(hash, raw[16:]) == 1
}
func (s *Service) authenticate(ctx context.Context, c Customer) (AuthResult, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return AuthResult{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if err := s.identity.Session(ctx, tokenHash(token), c.ID, time.Now().Add(7*24*time.Hour)); err != nil {
		return AuthResult{}, err
	}
	return AuthResult{c, token}, nil
}
func (s *Service) Register(ctx context.Context, input AuthInput) (AuthResult, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	parsed, err := mail.ParseAddress(input.Email)
	if err != nil || parsed.Address != input.Email || len(input.Email) > 254 || utf8.RuneCountInString(input.Name) < 2 || utf8.RuneCountInString(input.Name) > 100 || utf8.RuneCountInString(input.Password) < 15 || len(input.Password) > 128 {
		return AuthResult{}, Invalid("Informe nome, e-mail válido e senha de pelo menos 15 caracteres (máximo 128 bytes).")
	}
	if err := s.authenticationBudget(ctx, input.Email); err != nil {
		return AuthResult{}, err
	}
	defer func() { <-s.hashSlots }()
	hash, err := HashPassword(input.Password)
	if err != nil {
		return AuthResult{}, err
	}
	c := Customer{ID: uuid.NewString(), Name: input.Name, Email: input.Email}
	var delivery *AccountDelivery
	if s.registered != nil {
		delivery, err = s.registered(c)
		if err != nil {
			return AuthResult{}, err
		}
	}
	if err = s.identity.Create(ctx, c, hash, delivery); err != nil {
		return AuthResult{}, err
	}
	return AuthResult{Customer: c}, nil
}
func (s *Service) Login(ctx context.Context, input AuthInput) (AuthResult, error) {
	if len(input.Password) > 128 || len(input.Email) > 254 {
		return AuthResult{}, Unauthorized
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if err := s.authenticationBudget(ctx, input.Email); err != nil {
		return AuthResult{}, err
	}
	defer func() { <-s.hashSlots }()
	cred, err := s.identity.ByEmail(ctx, input.Email)
	if err != nil {
		if err == Unauthorized {
			passwordMatches(input.Password, "argon2id$v1$"+strings.Repeat("A", 64))
		}
		return AuthResult{}, err
	}
	if !passwordMatches(input.Password, cred.Hash) {
		return AuthResult{}, Unauthorized
	}
	return s.authenticate(ctx, cred.Customer)
}
func (s *Service) Current(ctx context.Context, token string) (Customer, error) {
	if len(token) != 43 {
		return Customer{}, Unauthorized
	}
	return s.identity.Current(ctx, tokenHash(token))
}
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.identity.Logout(ctx, tokenHash(token))
}

func (s *Service) UpdateName(ctx context.Context, token, name string) (Customer, error) {
	if len(token) != 43 {
		return Customer{}, Unauthorized
	}
	name = strings.TrimSpace(name)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 100 || strings.ContainsFunc(name, unicode.IsControl) {
		return Customer{}, Invalid("Informe um nome com 2 a 100 caracteres.")
	}
	return s.identity.UpdateName(ctx, tokenHash(token), name)
}

type Item struct {
	ID       string `json:"id"`
	Finish   string `json:"finish"`
	Quantity int    `json:"quantity"`
}
type Address struct {
	CEP    string `json:"cep"`
	Street string `json:"street"`
	Number string `json:"number"`
	City   string `json:"city"`
	State  string `json:"state"`
}
type OrderInput struct {
	Key             string  `json:"idempotencyKey"`
	ShippingQuoteID string  `json:"shippingQuoteId"`
	Items           []Item  `json:"items"`
	Address         Address `json:"address"`
}
type OrderItem struct {
	Item
	Name           string `json:"name"`
	UnitPriceCents int64  `json:"unitPriceCents"`
}
type Order struct {
	ID                   string         `json:"id"`
	Status               string         `json:"status"`
	SubtotalCents        int64          `json:"subtotalCents"`
	ShippingCents        int64          `json:"shippingCents"`
	TotalCents           int64          `json:"totalCents"`
	Address              Address        `json:"address"`
	Items                []OrderItem    `json:"items"`
	Shipping             *ShippingQuote `json:"shipping,omitempty"`
	ReservationExpiresAt time.Time      `json:"reservationExpiresAt"`
	CreatedAt            time.Time      `json:"createdAt"`
}

var cep = regexp.MustCompile(`^\d{5}-?\d{3}$`)

func (s *Service) CreateOrder(ctx context.Context, user string, in OrderInput) (Order, error) {
	if _, err := uuid.Parse(in.Key); err != nil {
		return Order{}, Invalid("Identificador de envio inválido.")
	}
	if err := validateItems(in.Items); err != nil {
		return Order{}, err
	}
	if _, err := uuid.Parse(in.ShippingQuoteID); err != nil {
		return Order{}, Invalid("Solicite uma cotação de frete válida.")
	}
	a := in.Address
	if !cep.MatchString(a.CEP) || len(strings.TrimSpace(a.Street)) < 2 || len(a.Street) > 200 || len(strings.TrimSpace(a.Number)) < 1 || len(a.Number) > 20 || len(strings.TrimSpace(a.City)) < 2 || len(a.City) > 100 || !strings.Contains("|AC|AL|AP|AM|BA|CE|DF|ES|GO|MA|MT|MS|MG|PA|PB|PR|PE|PI|RJ|RN|RS|RO|RR|SC|SP|SE|TO|", "|"+a.State+"|") || len(a.State) != 2 {
		return Order{}, Invalid("Confira o endereço de entrega.")
	}
	return s.orders.Create(ctx, user, in)
}
func (s *Service) Orders(ctx context.Context, user string) ([]Order, error) {
	return s.orders.List(ctx, user)
}
func (s *Service) Order(ctx context.Context, user, id string) (Order, error) {
	if _, err := uuid.Parse(id); err != nil {
		return Order{}, NotFound
	}
	return s.orders.Find(ctx, user, id)
}
func (s *Service) Cancel(ctx context.Context, user, id string) (Order, error) {
	if _, err := uuid.Parse(id); err != nil {
		return Order{}, NotFound
	}
	return s.orders.Cancel(ctx, user, id)
}

type PaymentGateway interface {
	Process(context.Context, Order) (string, error)
}

func (s *Service) Pay(ctx context.Context, user, id string) (Order, error) {
	if _, err := uuid.Parse(id); err != nil {
		return Order{}, NotFound
	}
	return s.orders.Payment(ctx, user, id, s.gateway)
}

type Variant struct {
	Finish string `json:"finish"`
	Stock  int    `json:"stock"`
}

func (s *Service) Variants(ctx context.Context, id string) ([]Variant, error) {
	return s.orders.Variants(ctx, id)
}

func validateItems(items []Item) error {
	if len(items) < 1 || len(items) > 30 {
		return Invalid("Selecione de 1 a 30 acabamentos.")
	}
	seen := map[string]bool{}
	for _, i := range items {
		key := i.ID + "\x00" + i.Finish
		if i.Quantity < 1 || i.Quantity > 10 || len(i.ID) > 100 || len(i.Finish) > 100 || seen[key] {
			return Invalid("Itens ou quantidades inválidos.")
		}
		seen[key] = true
	}

	return nil
}

func (s *Service) authenticationBudget(ctx context.Context, email string) error {
	allowed, err := s.identity.AllowAuthentication(ctx, tokenHash(email))
	if err != nil {
		return err
	}
	if !allowed {
		return &Failure{429, "auth_rate_limited", "Muitas tentativas. Aguarde um minuto e tente novamente."}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case s.hashSlots <- struct{}{}:
		return nil
	default:
		return &Failure{503, "auth_busy", "O serviço está ocupado. Tente novamente em instantes."}
	}
}

// Configured by Fx before the HTTP listener starts.
func (s *Service) SetRegistrationPreparer(notify func(Customer) (*AccountDelivery, error)) {
	s.registered = notify
}

var passwordHashSlots = make(chan struct{}, 2)

func PasswordHashSlots() chan struct{} { return passwordHashSlots }

type AccountDelivery struct {
	TokenHash, Email, Subject, Body string
	ExpiresAt                       time.Time
}
