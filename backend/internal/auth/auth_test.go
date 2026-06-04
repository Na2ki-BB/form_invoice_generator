package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewFromEnv(t *testing.T) {
	t.Run("default is local", func(t *testing.T) {
		t.Setenv("APP_AUTH_MODE", "")
		authenticator, err := NewFromEnv()
		if err != nil {
			t.Fatalf("NewFromEnv() error = %v", err)
		}
		if authenticator.mode != ModeLocal {
			t.Fatalf("mode = %q, want %q", authenticator.mode, ModeLocal)
		}
	})

	t.Run("cognito requires issuer and audience", func(t *testing.T) {
		t.Setenv("APP_AUTH_MODE", "cognito")
		t.Setenv("APP_COGNITO_ISSUER", "")
		t.Setenv("APP_COGNITO_AUDIENCE", "")
		if _, err := NewFromEnv(); err == nil {
			t.Fatal("NewFromEnv() error = nil, want error")
		}
	})

	t.Run("unknown mode is rejected", func(t *testing.T) {
		t.Setenv("APP_AUTH_MODE", "unknown")
		if _, err := NewFromEnv(); err == nil {
			t.Fatal("NewFromEnv() error = nil, want error")
		}
	})
}

func TestLocalAuthenticator(t *testing.T) {
	authenticator := &Authenticator{mode: ModeLocal}
	request := httptest.NewRequest(http.MethodGet, "/admin/products", nil)
	request.RemoteAddr = "127.0.0.1:12345"
	request.Header.Set("X-Local-Admin", "true")

	if err := authenticator.AuthenticateAdmin(request); err != nil {
		t.Fatalf("AuthenticateAdmin() error = %v", err)
	}
}

func TestLocalAuthenticatorRejectsMissingHeader(t *testing.T) {
	authenticator := &Authenticator{mode: ModeLocal}
	request := httptest.NewRequest(http.MethodGet, "/admin/products", nil)
	request.RemoteAddr = "127.0.0.1:12345"

	if err := authenticator.AuthenticateAdmin(request); err == nil {
		t.Fatal("AuthenticateAdmin() error = nil, want error")
	}
}

func TestCognitoVerifier(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	const keyID = "test-key"
	const issuer = "https://cognito-idp.ap-northeast-1.amazonaws.com/test-pool"
	const audience = "test-client"
	now := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks{Keys: []jwk{publicJWK(keyID, &privateKey.PublicKey)}})
	}))
	defer jwksServer.Close()

	verifier, err := NewCognitoVerifier(
		issuer,
		audience,
		WithJWKSURL(jwksServer.URL),
		WithHTTPClient(jwksServer.Client()),
		WithNow(func() time.Time { return now }),
	)
	if err != nil {
		t.Fatalf("NewCognitoVerifier() error = %v", err)
	}

	token := signedToken(t, privateKey, keyID, map[string]any{
		"iss":       issuer,
		"exp":       now.Add(time.Hour).Unix(),
		"client_id": audience,
		"token_use": "access",
	})

	if err := verifier.VerifyToken(context.Background(), token); err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}
}

func TestCognitoVerifierRejectsWrongAudience(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	const keyID = "test-key"
	const issuer = "https://cognito-idp.ap-northeast-1.amazonaws.com/test-pool"
	now := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks{Keys: []jwk{publicJWK(keyID, &privateKey.PublicKey)}})
	}))
	defer jwksServer.Close()

	verifier, err := NewCognitoVerifier(
		issuer,
		"expected-client",
		WithJWKSURL(jwksServer.URL),
		WithHTTPClient(jwksServer.Client()),
		WithNow(func() time.Time { return now }),
	)
	if err != nil {
		t.Fatalf("NewCognitoVerifier() error = %v", err)
	}

	token := signedToken(t, privateKey, keyID, map[string]any{
		"iss":       issuer,
		"exp":       now.Add(time.Hour).Unix(),
		"client_id": "other-client",
		"token_use": "access",
	})

	if err := verifier.VerifyToken(context.Background(), token); err == nil {
		t.Fatal("VerifyToken() error = nil, want error")
	}
}

func publicJWK(keyID string, key *rsa.PublicKey) jwk {
	return jwk{
		KeyType:  "RSA",
		KeyID:    keyID,
		Use:      "sig",
		Modulus:  base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		Exponent: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
	}
}

func signedToken(t *testing.T, privateKey *rsa.PrivateKey, keyID string, claims map[string]any) string {
	t.Helper()

	header := map[string]any{"alg": "RS256", "kid": keyID}
	signingInput := encodeJSON(t, header) + "." + encodeJSON(t, claims)
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func encodeJSON(t *testing.T, value any) string {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}
