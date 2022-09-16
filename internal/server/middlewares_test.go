package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tinfoil-knight/gargoyle/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func TestResponseHeaderModification(t *testing.T) {
	handler := http.HandlerFunc(dummyHandler)

	t.Run("default Server header is present", func(t *testing.T) {
		cfg := config.HeaderCfg{}
		req := httptest.NewRequest(http.MethodGet, "http://testing.com", nil)
		res := httptest.NewRecorder()

		useHeaderModifier(handler, cfg).ServeHTTP(res, req)

		want := "Gargoyle"
		got := res.Result().Header.Get("Server")

		if !strings.Contains(got, want) {
			t.Errorf("expected %q to have %q", got, want)
		}
	})

	t.Run("default Server header can be removed", func(t *testing.T) {
		cfg := config.HeaderCfg{
			Remove: []string{"Server"},
		}
		req := httptest.NewRequest(http.MethodGet, "http://testing.com", nil)
		res := httptest.NewRecorder()

		useHeaderModifier(handler, cfg).ServeHTTP(res, req)

		want := ""
		got := res.Result().Header.Get("Server")

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("adding a new header works", func(t *testing.T) {
		cfg := config.HeaderCfg{
			Add: map[string]string{
				"Foo": "Bar",
			},
		}
		req := httptest.NewRequest(http.MethodGet, "http://testing.com", nil)
		res := httptest.NewRecorder()

		useHeaderModifier(handler, cfg).ServeHTTP(res, req)

		want := "Bar"
		got := res.Result().Header.Get("Foo")

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("removing an existing header works", func(t *testing.T) {
		cfg := config.HeaderCfg{
			Remove: []string{"X-Request-Id"},
		}
		req := httptest.NewRequest(http.MethodGet, "http://testing.com", nil)
		res := httptest.NewRecorder()
		res.Header().Add("X-Request-Id", "1")

		useHeaderModifier(handler, cfg).ServeHTTP(res, req)

		want := ""
		got := res.Result().Header.Get("X-Request-Id")

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestAuth(t *testing.T) {
	handler := http.HandlerFunc(dummyHandler)
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	cfg := config.AuthConfig{
		BasicAuth: map[string][]byte{
			"first_user": hash,
		},
	}

	t.Run("basic auth throws error without password", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://testing.com", nil)
		res := httptest.NewRecorder()

		auth(handler, cfg).ServeHTTP(res, req)

		want := http.StatusUnauthorized
		got := res.Result().StatusCode

		if got != want {
			t.Errorf("expected %d, got %d", want, got)
		}
	})

	t.Run("basic auth throws error on wrong password", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://testing.com", nil)
		req.SetBasicAuth("first_user", "not_secret")
		res := httptest.NewRecorder()

		auth(handler, cfg).ServeHTTP(res, req)

		want := http.StatusUnauthorized
		got := res.Result().StatusCode

		if got != want {
			t.Errorf("expected %d, got %d", want, got)
		}
	})

	t.Run("basic auth works for correct username and password", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://testing.com", nil)
		req.SetBasicAuth("first_user", "secret")
		res := httptest.NewRecorder()

		auth(handler, cfg).ServeHTTP(res, req)

		want := http.StatusOK
		got := res.Result().StatusCode

		if got != want {
			t.Errorf("expected %d, got %d", want, got)
		}
	})
}

func dummyHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Hello World!"))
}
