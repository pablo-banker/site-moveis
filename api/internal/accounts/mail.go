package accounts

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type MailConfig struct {
	Host, Port, User, Password, From, Mode, Origin string
	Key                                            []byte
}

func NewMailConfig() (MailConfig, error) {
	get := func(k, d string) string {
		if v := os.Getenv(k); v != "" {
			return v
		}
		return d
	}
	c := MailConfig{Host: get("SMTP_HOST", "localhost"), Port: get("SMTP_PORT", "1025"), User: os.Getenv("SMTP_USER"), Password: os.Getenv("SMTP_PASSWORD"), From: get("SMTP_FROM", "Forma <conta@localhost>"), Mode: get("SMTP_TLS_MODE", "none"), Origin: get("SITE_ORIGIN", "http://127.0.0.1:5173")}
	u, err := url.Parse(c.Origin)
	if err != nil || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return c, fmt.Errorf("invalid SITE_ORIGIN")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return c, fmt.Errorf("invalid SITE_ORIGIN protocol")
	}
	if _, err = mail.ParseAddress(c.From); err != nil || strings.ContainsAny(c.From, "\r\n") {
		return c, fmt.Errorf("invalid SMTP_FROM")
	}
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return c, fmt.Errorf("invalid SMTP_PORT")
	}
	if c.Mode != "none" && c.Mode != "starttls" && c.Mode != "tls" {
		return c, fmt.Errorf("invalid SMTP_TLS_MODE")
	}
	key := get("EMAIL_OUTBOX_KEY", "")
	if os.Getenv("APP_ENV") == "production" {
		if u.Scheme != "https" || os.Getenv("SITE_ORIGIN") == "" || os.Getenv("SMTP_HOST") == "" || c.Mode == "none" || key == "" {
			return c, fmt.Errorf("production requires HTTPS SITE_ORIGIN, SMTP_HOST, SMTP TLS and EMAIL_OUTBOX_KEY")
		}
	}
	if key == "" {
		c.Key = make([]byte, 32)
	} else {
		c.Key, err = base64.StdEncoding.DecodeString(key)
		if err != nil || len(c.Key) != 32 {
			return c, fmt.Errorf("EMAIL_OUTBOX_KEY must be 32 random bytes in base64")
		}
	}
	return c, nil
}
func encodeBody(key []byte, body string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(body), []byte("forma-email-v1"))), nil
}
func DecodeBody(key []byte, body string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(body)
	if err != nil || len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("invalid encrypted email")
	}
	plain, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], []byte("forma-email-v1"))
	return string(plain), err
}
func (c MailConfig) send(recipient, subject, body string) error {
	from, err := mail.ParseAddress(c.From)
	if err != nil {
		return err
	}
	to, err := mail.ParseAddress(recipient)
	if err != nil || strings.ContainsAny(recipient, "\r\n") {
		return fmt.Errorf("invalid recipient")
	}
	addr := net.JoinHostPort(c.Host, c.Port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(8 * time.Second))
	tlsCfg := &tls.Config{ServerName: c.Host, MinVersion: tls.VersionTLS12}
	if c.Mode == "tls" {
		conn = tls.Client(conn, tlsCfg)
		if err = conn.(*tls.Conn).Handshake(); err != nil {
			return err
		}
	}
	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	if c.Mode == "starttls" {
		if err = client.StartTLS(tlsCfg); err != nil {
			return err
		}
	}
	if c.User != "" {
		if err = client.Auth(smtp.PlainAuth("", c.User, c.Password, c.Host)); err != nil {
			return err
		}
	}
	if err = client.Mail(from.Address); err != nil {
		return err
	}
	if err = client.Rcpt(to.Address); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	message := "From: " + from.String() + "\r\nTo: " + to.String() + "\r\nSubject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + strings.ReplaceAll(body, "\n", "\r\n") + "\r\n"
	if _, err = w.Write([]byte(message)); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
func DeliverOne(ctx context.Context, pool *pgxpool.Pool, cfg MailConfig, log *zap.Logger) (bool, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	var id, to, subject, encrypted string
	var attempts int
	err = tx.QueryRow(ctx, `SELECT id::text,recipient,subject,body,attempts FROM email_outbox WHERE sent_at IS NULL AND attempts<10 AND available_at<=now() ORDER BY available_at LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&id, &to, &subject, &encrypted, &attempts)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	body, deliveryErr := DecodeBody(cfg.Key, encrypted)
	if deliveryErr == nil {
		deliveryErr = cfg.send(to, subject, body)
	}
	if deliveryErr == nil {
		_, err = tx.Exec(ctx, `UPDATE email_outbox SET sent_at=now(),body='' WHERE id=$1`, id)
	} else {
		delay := time.Duration(1<<min(attempts, 6)) * time.Minute
		_, err = tx.Exec(ctx, `UPDATE email_outbox SET attempts=attempts+1,available_at=$2 WHERE id=$1`, id, time.Now().Add(delay))
		log.Warn("email delivery retry", zap.String("message_id", id), zap.Int("attempt", attempts+1))
	}
	if err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}
func RunMailer(lc fx.Lifecycle, pool *pgxpool.Pool, cfg MailConfig, log *zap.Logger) {
	var cancel context.CancelFunc
	done := make(chan struct{})
	lc.Append(fx.Hook{OnStart: func(context.Context) error {
		ctx, stop := context.WithCancel(context.Background())
		cancel = stop
		go func() {
			defer close(done)
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				for i := 0; i < 10 && ctx.Err() == nil; i++ {
					work, stop := context.WithTimeout(ctx, 12*time.Second)
					found, err := DeliverOne(work, pool, cfg, log)
					stop()
					if err != nil && ctx.Err() == nil {
						log.Error("email outbox failed")
					}
					if err != nil || !found {
						break
					}
				}
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}
		}()
		return nil
	}, OnStop: func(ctx context.Context) error {
		if cancel != nil {
			cancel()
		}
		select {
		case <-done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
}
