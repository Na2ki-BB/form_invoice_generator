package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"form-invoice-generator/backend/internal/auth"
	"form-invoice-generator/backend/internal/database"
	formrepository "form-invoice-generator/backend/internal/form"
	"form-invoice-generator/backend/internal/pricing"
	"form-invoice-generator/backend/internal/product"
	"form-invoice-generator/backend/internal/submission"
)

func TestRequireAdminLocal(t *testing.T) {
	handler := requireAdmin(auth.NewLocalAuthenticator(), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("rejects request without local admin header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/admin/products", nil)
		request.RemoteAddr = "127.0.0.1:12345"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
	})

	t.Run("allows request with local admin header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/admin/products", nil)
		request.RemoteAddr = "127.0.0.1:12345"
		request.Header.Set("X-Local-Admin", "true")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
	})

	t.Run("preflight allows authorization header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodOptions, "/admin/products", nil)
		request.RemoteAddr = "198.51.100.1:12345"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
		}
		if got := response.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Authorization") {
			t.Fatalf("Access-Control-Allow-Headers = %q, want Authorization", got)
		}
	})
}

func TestConfiguredAllowedOrigin(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("APP_CORS_ALLOWED_ORIGIN", "")
		if got := configuredAllowedOrigin(); got != defaultAllowedOrigin {
			t.Fatalf("configuredAllowedOrigin() = %q, want %q", got, defaultAllowedOrigin)
		}
	})

	t.Run("environment variable", func(t *testing.T) {
		t.Setenv("APP_CORS_ALLOWED_ORIGIN", "https://example.test")
		if got := configuredAllowedOrigin(); got != "https://example.test" {
			t.Fatalf("configuredAllowedOrigin() = %q, want custom origin", got)
		}
	})
}

func TestPublicAPIsIntegration(t *testing.T) {
	if os.Getenv("RUN_DATABASE_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_DATABASE_INTEGRATION_TESTS=1 to run PostgreSQL integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := database.Open(ctx)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	handler := newHandler(
		submission.NewRepository(db),
		formrepository.NewRepository(db),
		pricing.NewRepository(db),
		pricing.NewRuleRepository(db),
		product.NewRepository(db),
		auth.NewLocalAuthenticator(),
	)

	t.Run("public form", func(t *testing.T) {
		response := serveRequest(handler, http.MethodGet, "/public/forms/default", "")
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
		}
	})

	t.Run("public CORS preflight", func(t *testing.T) {
		response := serveRequest(handler, http.MethodOptions, "/public/forms/default", "")
		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusNoContent, response.Body.String())
		}
		if got := response.Header().Get("Access-Control-Allow-Origin"); got != defaultAllowedOrigin {
			t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, defaultAllowedOrigin)
		}
	})

	t.Run("old development invoice endpoint is removed", func(t *testing.T) {
		response := serveRequest(handler, http.MethodGet, "/invoice/download", "")
		if response.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusNotFound, response.Body.String())
		}
	})

	t.Run("discount quote", func(t *testing.T) {
		response := serveRequest(handler, http.MethodPost, "/public/forms/default/quote", `{"items":[{"productId":"ofuda","quantity":2}]}`)
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
		}
		var quote struct {
			Items       []pricing.Item `json:"items"`
			TotalAmount int            `json:"totalAmount"`
		}
		if err := json.NewDecoder(response.Body).Decode(&quote); err != nil {
			t.Fatalf("decode quote: %v", err)
		}
		if len(quote.Items) != 1 || quote.Items[0].UnitPrice != 950 || quote.TotalAmount != 1900 {
			t.Fatalf("quote = %+v, want ofuda unitPrice=950 totalAmount=1900", quote)
		}
	})

	t.Run("unknown form quote", func(t *testing.T) {
		response := serveRequest(handler, http.MethodPost, "/public/forms/missing/quote", `{"items":[{"productId":"ofuda","quantity":2}]}`)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
		}
	})

	t.Run("unknown JSON field is rejected", func(t *testing.T) {
		response := serveRequest(handler, http.MethodPost, "/public/forms/default/quote", `{"items":[{"productId":"ofuda","quantity":2}],"unexpected":true}`)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
		}
	})

	t.Run("duplicate product is rejected", func(t *testing.T) {
		response := serveRequest(handler, http.MethodPost, "/public/forms/default/quote", `{"items":[{"productId":"ofuda","quantity":1},{"productId":"ofuda","quantity":2}]}`)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
		}
	})

	t.Run("required customer field", func(t *testing.T) {
		response := serveRequest(handler, http.MethodPost, "/submissions", `{"formSlug":"default","customerName":" ","postalCode":"100-0001","address":"test address","phone":"000-0000","items":[{"productId":"prayer-a","quantity":1}]}`)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
		}
	})

	t.Run("submission returns JSON instead of invoice file", func(t *testing.T) {
		response := serveRequest(handler, http.MethodPost, "/submissions", `{"formSlug":"default","customerName":"API Test","postalCode":"100-0001","address":"test address","phone":"000-0000","items":[{"productId":"ofuda","quantity":3}]}`)
		if response.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body.String())
		}
		if got := response.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
			t.Fatalf("Content-Type = %q, want JSON", got)
		}
		var created struct {
			ID            int64  `json:"id"`
			InvoiceNumber string `json:"invoiceNumber"`
		}
		if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
			t.Fatalf("decode submission response: %v", err)
		}
		if created.ID == 0 || created.InvoiceNumber == "" {
			t.Fatalf("created submission response = %+v, want id and invoiceNumber", created)
		}
		if _, err := db.Exec(ctx, "DELETE FROM submission_items WHERE submission_id = $1", created.ID); err != nil {
			t.Fatalf("delete submission items: %v", err)
		}
		if _, err := db.Exec(ctx, "DELETE FROM submissions WHERE id = $1", created.ID); err != nil {
			t.Fatalf("delete submission: %v", err)
		}
	})
}

func serveRequest(handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
