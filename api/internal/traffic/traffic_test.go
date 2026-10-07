package traffic

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTrustedBFFSignature(t *testing.T) {
	secret := strings.Repeat("s", 32)
	app := fiber.New()
	app.Post("/api/v1/test", func(c fiber.Ctx) error {
		client, err := Client(c, secret, true)
		if err != nil {
			return err
		}
		return c.SendString(client)
	})
	for _, test := range []struct {
		age      int64
		tamper   bool
		expected int
	}{{0, false, 200}, {0, true, 403}, {-60, false, 403}, {60, false, 403}} {
		stamp := fmt.Sprint(time.Now().Unix() + test.age)
		client := "203.0.113.4"
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(stamp + "\nPOST\n/api/v1/test\n" + client))
		signature := fmt.Sprintf("%x", mac.Sum(nil))
		if test.tamper {
			client = "203.0.113.5"
		}
		request := httptest.NewRequest("POST", "/api/v1/test", nil)
		request.Header.Set("X-Forma-Client", client)
		request.Header.Set("X-Forma-Timestamp", stamp)
		request.Header.Set("X-Forma-Signature", signature)
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != test.expected {
			t.Fatalf("signature expected %d got %d", test.expected, response.StatusCode)
		}
	}
}
